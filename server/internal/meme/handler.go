package meme

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/changelog"
	"github.com/TogetherBros/azit/server/internal/httperr"
	"github.com/TogetherBros/azit/server/internal/live"
	"github.com/TogetherBros/azit/server/internal/member"
	"github.com/TogetherBros/azit/server/internal/storage"
	"github.com/gofiber/fiber/v3"
)

// PhotoSource 는 채팅에 올라온 사진을 찾는 쪽 (message.Repository).
// meme 이 message 를 import 하면 순환 import 라서 (message 가 이미 meme 을 쓴다) 인터페이스로 받는다.
type PhotoSource interface {
	PhotoKey(ctx context.Context, azitID, messageID int64) (string, error)
}

type Handler struct {
	repo   *Repository
	r2     *storage.R2
	photos PhotoSource           // 채팅 사진 → 짤로 저장
	live   *live.Live            // 짤·맞장구 목록이 바뀌면 멤버들에게 알림 (nil 이면 안 보냄)
	logs   *changelog.Repository // 변경 기록 (nil 이면 안 남김)
}

func NewHandler(repo *Repository, r2 *storage.R2, photos PhotoSource, lv *live.Live, logs *changelog.Repository) *Handler {
	return &Handler{repo: repo, r2: r2, photos: photos, live: lv, logs: logs}
}

// Routes 는 /v1/azits 그룹 아래에 짤, 맞장구 목록 경로를 단다.
func (h *Handler) Routes(r fiber.Router, requireMember fiber.Handler) {
	r.Get("/:id/memes", requireMember, h.ListMemes)
	r.Post("/:id/memes", requireMember, h.CreateMeme)
	r.Post("/:id/memes/from-message", requireMember, h.SaveFromMessage)
	r.Delete("/:id/memes/:memeId", requireMember, h.DeleteMeme)
	r.Get("/:id/reacts", requireMember, h.ListReacts)
	r.Put("/:id/reacts", requireMember, h.ReplaceReacts)
}

// ListMemes 는 아지트 짤 모음. 멤버면 누구나. GET /v1/azits/{id}/memes
func (h *Handler) ListMemes(c fiber.Ctx) error {
	me := member.Current(c)

	memes, err := h.repo.List(c.Context(), me.AzitID)
	if err != nil {
		return err
	}

	res := make([]proto.MemeResponse, 0, len(memes))
	for _, m := range memes {
		r, err := h.memeResponse(c.Context(), m)
		if err != nil {
			return err
		}
		res = append(res, r)
	}
	return c.JSON(res)
}

// CreateMeme 은 업로드한 짤을 아지트 짤 모음에 넣는다. 멤버면 누구나. POST /v1/azits/{id}/memes
// 순서: POST /v1/uploads (kind=meme) → R2 에 PUT → 여기로 key 를 보낸다.
func (h *Handler) CreateMeme(c fiber.Ctx) error {
	me := member.Current(c)

	var req proto.CreateMemeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}

	name := NormalizeName(req.Name)
	if err := ValidateName(name); err != nil {
		return badRequest(err)
	}

	// 내가 올린 짤 폴더(memes/{내 id}/...) 의 키인지. 남이 올린 파일 키를 가로채지 못하게.
	if !storage.OwnedBy(req.Key, storage.KindMeme, me.AccountID) {
		return badRequest(errors.New("올린 짤을 찾을 수 없어요"))
	}
	// 실제로 R2 에 올라가 있는지. 업로드 주소만 받고 안 올린 경우를 막는다.
	if _, _, err := h.r2.Head(c.Context(), req.Key); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return badRequest(errors.New("올린 짤을 찾을 수 없어요"))
		}
		return err
	}

	return h.create(c, me, name, req.Key)
}

// create 는 짤 한 줄을 넣고 기록·알림을 남긴 뒤 응답한다. 올리기와 "채팅 사진을 짤로" 가 같이 쓴다.
func (h *Handler) create(c fiber.Ctx, me member.Member, name, key string) error {
	m := Meme{
		AzitID:    me.AzitID,
		Name:      name,
		MediaKey:  key,
		CreatedBy: me.AccountID,
	}
	if err := h.repo.Create(c.Context(), &m); err != nil {
		return err
	}
	h.logs.Record(c.Context(), changelog.Entry{
		AzitID: me.AzitID, ActorID: me.AccountID, ActorNick: me.NickName,
		Kind: changelog.KindMemeCreate, TargetID: m.MemeID, Summary: fmt.Sprintf("'%s' 짤을 올림", m.Name),
	}, nil)
	h.live.AzitChanged(c.Context(), me.AzitID, proto.ChangedMemes)

	res, err := h.memeResponse(c.Context(), m)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(res)
}

// SaveFromMessage 는 채팅에 올라온 사진을 아지트 짤 모음에 넣는다. 멤버면 누구나.
// POST /v1/azits/{id}/memes/from-message {"message_id": 812, "name": "킹받네"}
// 사진 파일은 보낸 사람 폴더에 있어서, 내 짤 폴더(memes/{내 id}/)로 복사해 둔다 (사진 메시지가 지워져도 짤은 남게).
func (h *Handler) SaveFromMessage(c fiber.Ctx) error {
	me := member.Current(c)

	var req proto.SaveMemeFromMessageRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}
	name := NormalizeName(req.Name)
	if err := ValidateName(name); err != nil {
		return badRequest(err)
	}

	src, err := h.photos.PhotoKey(c.Context(), me.AzitID, req.MessageID)
	if err != nil {
		return badRequest(err)
	}
	size, contentType, err := h.r2.Head(c.Context(), src)
	if err != nil {
		return err
	}
	dst, err := storage.NewKey(storage.KindMeme, me.AccountID, contentType, size)
	if err != nil {
		return badRequest(err)
	}
	if err := h.r2.Copy(c.Context(), src, dst); err != nil {
		return err
	}
	return h.create(c, me, name, dst)
}

// DeleteMeme 은 짤을 지운다. 올린 사람이나 관리자 이상만. DELETE /v1/azits/{id}/memes/{memeId}
// 맞장구 목록에서도 빠진다. 이미 메시지에 달린 짤 맞장구는 그대로 보인다.
func (h *Handler) DeleteMeme(c fiber.Ctx) error {
	me := member.Current(c)

	memeID := fiber.Params[int64](c, "memeId")
	m, err := h.repo.FindActive(c.Context(), me.AzitID, memeID)
	if errors.Is(err, ErrNotFound) {
		return httperr.New(fiber.StatusNotFound, proto.ErrCodeNotFound, ErrNotFound.Error())
	}
	if err != nil {
		return err
	}

	if m.CreatedBy != me.AccountID && !me.Role.AtLeast(member.RoleAdmin) {
		return httperr.New(fiber.StatusForbidden, proto.ErrCodeForbidden, "올린 사람이나 관리자만 지울 수 있어요")
	}

	if err := h.repo.SoftDelete(c.Context(), memeID); err != nil {
		return err
	}
	h.logs.Record(c.Context(), changelog.Entry{
		AzitID: me.AzitID, ActorID: me.AccountID, ActorNick: me.NickName,
		Kind: changelog.KindMemeDelete, TargetID: memeID, Summary: fmt.Sprintf("'%s' 짤을 뺌", m.Name),
	}, nil)
	h.live.AzitChanged(c.Context(), me.AzitID, proto.ChangedMemes)
	h.live.AzitChanged(c.Context(), me.AzitID, proto.ChangedReacts) // 맞장구 목록에서도 빠졌다
	return c.SendStatus(fiber.StatusNoContent)
}

// ListReacts 는 아지트 맞장구 목록 (메시지에 맞장구 달 때 빠르게 고르는 칸들). 멤버면 누구나. GET /v1/azits/{id}/reacts
func (h *Handler) ListReacts(c fiber.Ctx) error {
	me := member.Current(c)
	return h.writeReacts(c, me.AzitID)
}

// ReplaceReacts 는 아지트 맞장구 목록을 통째로 바꾼다. 관리자 이상만. PUT /v1/azits/{id}/reacts
// 바뀐 목록을 돌려준다.
func (h *Handler) ReplaceReacts(c fiber.Ctx) error {
	me := member.Current(c)
	if !me.Role.AtLeast(member.RoleAdmin) {
		return httperr.New(fiber.StatusForbidden, proto.ErrCodeForbidden, "관리자만 할 수 있어요")
	}

	var req proto.ReplaceReactsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}
	if len(req.Reacts) > MaxReacts {
		return badRequest(ErrTooManyReacts)
	}

	// 칸마다 검사하면서 모델로 옮긴다. 같은 맞장구가 두 번 들어오면 400.
	type key struct {
		text   string
		memeID int64
	}
	seen := map[key]bool{}
	reacts := make([]React, 0, len(req.Reacts))
	var memeIDs []int64
	for _, item := range req.Reacts {
		text := NormalizeReactText(item.Text)
		if err := ValidateReact(text, item.MemeID); err != nil {
			return badRequest(err)
		}
		k := key{text, item.MemeID}
		if seen[k] {
			return badRequest(ErrDuplicateReact)
		}
		seen[k] = true

		react := React{Text: text}
		if item.MemeID > 0 {
			id := item.MemeID
			react.MemeID = &id
			memeIDs = append(memeIDs, id)
		}
		reacts = append(reacts, react)
	}

	// 짤 칸이 전부 이 아지트의 지우지 않은 짤인지
	if len(memeIDs) > 0 {
		n, err := h.repo.CountActiveAmong(c.Context(), me.AzitID, memeIDs)
		if err != nil {
			return err
		}
		if n != int64(len(memeIDs)) {
			return badRequest(ErrNotFound)
		}
	}

	before, err := h.repo.ListReacts(c.Context(), me.AzitID) // 되돌리기용 옛 목록
	if err != nil {
		return err
	}
	if err := h.repo.ReplaceReacts(c.Context(), me.AzitID, reacts); err != nil {
		return err
	}
	old := make([]proto.ReactItem, 0, len(before))
	for _, r := range before {
		item := proto.ReactItem{Text: r.Text}
		if r.MemeID != nil {
			item.MemeID = *r.MemeID
		}
		old = append(old, item)
	}
	h.logs.Record(c.Context(), changelog.Entry{
		AzitID: me.AzitID, ActorID: me.AccountID, ActorNick: me.NickName,
		Kind: changelog.KindReacts, Summary: "맞장구 목록을 바꿈",
	}, old)
	h.live.AzitChanged(c.Context(), me.AzitID, proto.ChangedReacts)
	return h.writeReacts(c, me.AzitID)
}

// writeReacts 는 아지트 맞장구 목록을 읽어 응답한다. 짤 칸에는 보기 링크를 붙인다.
func (h *Handler) writeReacts(c fiber.Ctx, azitID int64) error {
	reacts, err := h.repo.ListReacts(c.Context(), azitID)
	if err != nil {
		return err
	}

	var memeIDs []int64
	for _, r := range reacts {
		if r.MemeID != nil {
			memeIDs = append(memeIDs, *r.MemeID)
		}
	}
	memes, err := h.repo.FindByIDs(c.Context(), slices.Compact(slices.Sorted(slices.Values(memeIDs))))
	if err != nil {
		return err
	}

	res := make([]proto.ReactResponse, 0, len(reacts))
	for _, r := range reacts {
		item := proto.ReactResponse{ReactID: r.ReactID, Text: r.Text, Position: r.Position}
		if r.MemeID != nil {
			item.MemeID = *r.MemeID
			if m, ok := memes[*r.MemeID]; ok {
				item.MemeURL, err = h.r2.PresignGet(c.Context(), m.MediaKey)
				if err != nil {
					return err
				}
			}
		}
		res = append(res, item)
	}
	return c.JSON(res)
}

func (h *Handler) memeResponse(ctx context.Context, m Meme) (proto.MemeResponse, error) {
	url, err := h.r2.PresignGet(ctx, m.MediaKey)
	if err != nil {
		return proto.MemeResponse{}, err
	}
	return proto.MemeResponse{
		MemeID:    m.MemeID,
		Name:      m.Name,
		URL:       url,
		CreatedBy: m.CreatedBy,
		CreatedAt: m.CreatedAt,
	}, nil
}

func badRequest(err error) error {
	return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, err.Error())
}
