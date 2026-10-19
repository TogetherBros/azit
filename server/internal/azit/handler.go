package azit

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/account"
	"github.com/TogetherBros/azit/server/internal/auth"
	"github.com/TogetherBros/azit/server/internal/changelog"
	"github.com/TogetherBros/azit/server/internal/httperr"
	"github.com/TogetherBros/azit/server/internal/live"
	"github.com/TogetherBros/azit/server/internal/member"
	"github.com/TogetherBros/azit/server/internal/storage"
	"github.com/gofiber/fiber/v3"
)

// Onliner 는 "지금 앱을 켜 두었나" 를 아는 쪽 (hub.Hub).
type Onliner interface {
	Online(accountID int64) bool
}

// UnreadCounter 는 방마다 안 읽은 수를 세는 쪽 (message.Repository). room.UnreadCounter 와 같은 모양.
type UnreadCounter interface {
	CountUnread(ctx context.Context, azitID, accountID int64) (map[int64]int64, error)
}

type Handler struct {
	accountRepo *account.Repository
	memberRepo  *member.Repository
	azitRepo    *Repository
	r2          *storage.R2
	online      Onliner       // 홈 카드 "접속 N명"
	unread      UnreadCounter // 홈 카드 안 읽은 수
	live        *live.Live    // 기본 정보가 바뀌면 멤버들에게 알림 (nil 이면 안 보냄)
	logs        *changelog.Repository
}

func NewHandler(accountRepo *account.Repository, memberRepo *member.Repository, azitRepo *Repository, r2 *storage.R2, online Onliner, unread UnreadCounter, lv *live.Live, logs *changelog.Repository) *Handler {
	return &Handler{
		accountRepo: accountRepo,
		memberRepo:  memberRepo,
		azitRepo:    azitRepo,
		r2:          r2,
		online:      online,
		unread:      unread,
		live:        lv,
		logs:        logs,
	}
}

// stats 는 홈 카드에 붙는 숫자: 지금 접속한 멤버 수, 내가 안 읽은 메시지 수 (모든 방 합).
func (h *Handler) stats(ctx context.Context, azitID, accountID int64) (online, unread int64, err error) {
	ids, err := h.memberRepo.ActiveIDs(ctx, azitID)
	if err != nil {
		return 0, 0, err
	}
	for _, id := range ids {
		if h.online != nil && h.online.Online(id) {
			online++
		}
	}
	if h.unread != nil {
		perRoom, err := h.unread.CountUnread(ctx, azitID, accountID)
		if err != nil {
			return 0, 0, err
		}
		for _, n := range perRoom {
			unread += n
		}
	}
	return online, unread, nil
}

// Create 는 아지트를 만들고, 만든 사람을 super_admin 멤버로 넣는다. POST /v1/azits
func (h *Handler) Create(c fiber.Ctx) error {
	var req proto.CreateAzitRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}

	ctx := c.Context()
	accountID := auth.AccountID(c)

	acc, err := h.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return err
	}

	input := CreateInput{
		MasterID:  accountID,
		Nickname:  acc.Nickname,
		Name:      req.Name,
		Color:     req.Color,
		BannerKey: req.BannerKey,
	}

	input.Normalize()
	if err := input.Validate(); err != nil {
		return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, err.Error())
	}

	if input.BannerKey != "" {
		if !storage.OwnedBy(input.BannerKey, storage.KindAzitBanner, accountID) {
			return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, "내 사진이 아니에요")
		}

		if _, _, err := h.r2.Head(ctx, input.BannerKey); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, "업로드가 아직 안 끝났어요")
			}
			return err
		}
	}

	a, err := h.azitRepo.Create(ctx, input)
	if err != nil {
		return err
	}

	bannerURL, err := h.bannerURL(ctx, a.BannerKey)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(proto.AzitResponse{
		AzitID:      a.AzitID,
		Name:        a.Name,
		Color:       a.Color,
		BannerURL:   bannerURL,
		MemberCount: 1,
		MyRole:      string(member.RoleSuperAdmin),
		MyNickname:  input.Nickname,
	})
}

// List 는 내가 들어가 있는 아지트 목록. GET /v1/azits
// 쿼리는 아지트 개수와 상관없이 두 번이다 (목록 1번, 인원 수 1번).
func (h *Handler) List(c fiber.Ctx) error {
	ctx := c.Context()
	accountID := auth.AccountID(c)

	azits, err := h.azitRepo.FindMyAzits(ctx, accountID)
	if err != nil {
		return err
	}

	ids := make([]int64, 0, len(azits))
	for _, a := range azits {
		ids = append(ids, a.AzitID)
	}

	counts, err := h.memberRepo.CountActive(ctx, ids)
	if err != nil {
		return err
	}

	// make 로 만들어야 아지트가 없을 때 null 이 아니라 [] 로 나간다.
	res := make([]proto.AzitResponse, 0, len(azits))
	for _, a := range azits {
		bannerURL, err := h.bannerURL(ctx, a.BannerKey)
		if err != nil {
			return err
		}

		// 내 역할·닉네임, 접속 수·안 읽은 수. 들어간 아지트는 몇 개 안 돼서 하나씩 읽는다.
		me, err := h.memberRepo.FindActive(ctx, a.AzitID, accountID)
		if err != nil {
			return err
		}
		online, unread, err := h.stats(ctx, a.AzitID, accountID)
		if err != nil {
			return err
		}

		res = append(res, proto.AzitResponse{
			AzitID:      a.AzitID,
			Name:        a.Name,
			Color:       a.Color,
			Description: a.Description,
			BannerURL:   bannerURL,
			MemberCount: counts[a.AzitID],
			OnlineCount: online,
			Unread:      unread,
			MyRole:      string(me.Role),
			MyNickname:  me.NickName,
		})
	}

	return c.JSON(res)
}

// infoSummary 는 변경 기록 한 줄 문구. 무엇이 바뀌었는지 나열한다.
func infoSummary(old, cur Azit) string {
	var parts []string
	if old.Name != cur.Name {
		parts = append(parts, fmt.Sprintf("이름을 '%s'(으)로", cur.Name))
	}
	if old.Color != cur.Color {
		parts = append(parts, "색")
	}
	if old.Description != cur.Description {
		parts = append(parts, "소개")
	}
	if old.BannerKey != cur.BannerKey {
		parts = append(parts, "배너")
	}
	if len(parts) == 0 {
		return "기본 정보를 저장함"
	}
	return "아지트 " + strings.Join(parts, ", ") + " 바꿈"
}

// bannerURL 은 배너 키로 보기 링크를 만든다. 키가 없으면 "".
func (h *Handler) bannerURL(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	return h.r2.PresignGet(ctx, key)
}

// Routes 는 /v1/azits 아래 경로를 붙인다. r 에는 인증 미들웨어가 걸려 있어야 한다.
// requireMember 는 /:id 경로에 끼우는 아지트 멤버 확인 미들웨어.
func (h *Handler) Routes(r fiber.Router, requireMember fiber.Handler) {
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Patch("/:id", requireMember, h.Update)
}

// Update 는 아지트 기본 정보(이름, 색, 소개, 배너)를 바꾼다. 관리자 이상만. PATCH /v1/azits/{id}
// 보낸 칸만 바꾸고, 바뀐 아지트를 목록과 같은 모양으로 돌려준다.
func (h *Handler) Update(c fiber.Ctx) error {
	ctx := c.Context()
	me := member.Current(c)
	if !me.Role.AtLeast(member.RoleAdmin) {
		return httperr.New(fiber.StatusForbidden, proto.ErrCodeForbidden, "관리자만 바꿀 수 있어요")
	}

	var req proto.UpdateAzitRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}
	input := UpdateInput{Name: req.Name, Color: req.Color, Description: req.Description, BannerKey: req.BannerKey}
	input.Normalize()
	if err := input.Validate(); err != nil {
		return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, err.Error())
	}

	// 새 배너는 내가 올린 아지트 배너여야 하고, R2 에 다 올라가 있어야 한다. "" 는 배너 지우기
	if input.BannerKey != nil && *input.BannerKey != "" {
		key := *input.BannerKey
		if !storage.OwnedBy(key, storage.KindAzitBanner, me.AccountID) {
			return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, "내 사진이 아니에요")
		}
		if _, _, err := h.r2.Head(ctx, key); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, "업로드가 아직 안 끝났어요")
			}
			return err
		}
	}

	old, err := h.azitRepo.FindByID(ctx, me.AzitID) // 되돌리기용 바꾸기 전 값
	if err != nil {
		return err
	}
	changes := input.Changes()
	if err := h.azitRepo.Update(ctx, me.AzitID, changes); err != nil {
		return err
	}

	a, err := h.azitRepo.FindByID(ctx, me.AzitID)
	if err != nil {
		return err
	}
	if len(changes) > 0 {
		before := map[string]any{}
		for k := range changes {
			switch k {
			case "name":
				before[k] = old.Name
			case "color":
				before[k] = old.Color
			case "description":
				before[k] = old.Description
			case "banner_key":
				before[k] = old.BannerKey
			}
		}
		h.logs.Record(ctx, changelog.Entry{
			AzitID: me.AzitID, ActorID: me.AccountID, ActorNick: me.NickName,
			Kind: changelog.KindAzitInfo, Summary: infoSummary(old, a),
		}, before)
	}
	counts, err := h.memberRepo.CountActive(ctx, []int64{a.AzitID})
	if err != nil {
		return err
	}
	bannerURL, err := h.bannerURL(ctx, a.BannerKey)
	if err != nil {
		return err
	}
	online, unread, err := h.stats(ctx, a.AzitID, me.AccountID)
	if err != nil {
		return err
	}

	// 다른 멤버들의 왼쪽 아지트 이름·색도 바로 바뀌게
	h.live.AzitChanged(ctx, a.AzitID, proto.ChangedInfo)

	return c.JSON(proto.AzitResponse{
		AzitID:      a.AzitID,
		Name:        a.Name,
		Color:       a.Color,
		Description: a.Description,
		BannerURL:   bannerURL,
		MemberCount: counts[a.AzitID],
		OnlineCount: online,
		Unread:      unread,
		MyRole:      string(me.Role),
		MyNickname:  me.NickName,
	})
}
