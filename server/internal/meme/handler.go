package meme

import (
	"context"
	"errors"
	"slices"

	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/httperr"
	"github.com/TogetherBros/azit/server/internal/member"
	"github.com/TogetherBros/azit/server/internal/storage"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	repo *Repository
	r2   *storage.R2
}

func NewHandler(repo *Repository, r2 *storage.R2) *Handler {
	return &Handler{repo: repo, r2: r2}
}

// Routes 는 /v1/azits 그룹 아래에 짤, 맞장구 목록 경로를 단다.
func (h *Handler) Routes(r fiber.Router, requireMember fiber.Handler) {
	r.Get("/:id/memes", requireMember, h.ListMemes)
	r.Post("/:id/memes", requireMember, h.CreateMeme)
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

	m := Meme{
		AzitID:    me.AzitID,
		Name:      name,
		MediaKey:  req.Key,
		CreatedBy: me.AccountID,
	}
	if err := h.repo.Create(c.Context(), &m); err != nil {
		return err
	}

	res, err := h.memeResponse(c.Context(), m)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(res)
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

	if err := h.repo.ReplaceReacts(c.Context(), me.AzitID, reacts); err != nil {
		return err
	}
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
