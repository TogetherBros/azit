package profile

import (
	"context"
	"errors"
	"strings"

	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/account"
	"github.com/TogetherBros/azit/server/internal/auth"
	"github.com/TogetherBros/azit/server/internal/httperr"
	"github.com/TogetherBros/azit/server/internal/storage"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	account *account.Repository
	r2      *storage.R2
}

func NewHandler(account *account.Repository, r2 *storage.R2) *Handler {
	return &Handler{account: account, r2: r2}
}

// Routes 는 /v1/me 아래 경로를 붙인다. r 에는 인증 미들웨어가 걸려 있어야 한다.
func (h *Handler) Routes(r fiber.Router) {
	r.Get("/", h.Me)
	r.Patch("/", h.UpdateMe)
	r.Get("/name-check", h.NameCheck)
	r.Put("/avatar", h.SaveAvatar)
	r.Put("/banner", h.SaveBanner)
}

func (h *Handler) SaveAvatar(c fiber.Ctx) error {
	var req proto.SetImageRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}

	accountID := auth.AccountID(c)
	if !storage.OwnedBy(req.Key, storage.KindAvatar, accountID) {
		return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, "내 사진이 아니에요")
	}

	if _, _, err := h.r2.Head(c.Context(), req.Key); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, "업로드가 아직 안 끝났어요")
		}
		return err
	}

	if err := h.account.SaveAvatarKey(c.Context(), accountID, req.Key); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) SaveBanner(c fiber.Ctx) error {
	var req proto.SetImageRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}

	accountID := auth.AccountID(c)
	if !storage.OwnedBy(req.Key, storage.KindBanner, accountID) {
		return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, "내 사진이 아니에요")
	}

	if _, _, err := h.r2.Head(c.Context(), req.Key); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, "업로드가 아직 안 끝났어요")
		}
		return err
	}

	if err := h.account.SaveBannerKey(c.Context(), accountID, req.Key); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Me 는 내 프로필을 돌려준다. 사진은 key 대신 24시간짜리 보기 주소로 바꿔서 준다.
func (h *Handler) Me(c fiber.Ctx) error {
	res, err := h.meResponse(c.Context(), auth.AccountID(c))
	if err != nil {
		return err
	}
	return c.JSON(res)
}

// UpdateMe 는 내 프로필(이름, 닉네임, 소개, 색)을 저장하고, 저장된 프로필을 GET /v1/me 와 같은 모양으로 돌려준다.
func (h *Handler) UpdateMe(c fiber.Ctx) error {
	var req proto.UpdateMeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}

	p := account.ProfileUpdate{
		Name:     req.Name,
		Nickname: req.Nickname,
		Bio:      req.Bio,
		Color:    req.Color,
	}
	p.Normalize()
	if err := p.Validate(); err != nil {
		return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, err.Error())
	}

	ctx := c.Context()
	accountID := auth.AccountID(c)

	err := h.account.UpdateProfile(ctx, accountID, p)
	if errors.Is(err, account.ErrNameTaken) {
		return httperr.New(fiber.StatusConflict, proto.ErrCodeNameTaken, err.Error())
	}
	if err != nil {
		return err
	}

	res, err := h.meResponse(ctx, accountID)
	if err != nil {
		return err
	}
	return c.JSON(res)
}

// NameCheck 는 그 이름을 내가 쓸 수 있는지 알려 준다. 지금 내 이름이면 쓸 수 있다고 본다.
// 쓸 수 없어도 200 으로 답한다. 입력 중에 계속 부르는 API 라 에러로 만들지 않는다.
func (h *Handler) NameCheck(c fiber.Ctx) error {
	name := strings.TrimSpace(c.Query("name"))
	if err := account.ValidateName(name); err != nil {
		return c.JSON(proto.NameCheckResponse{Available: false, Message: err.Error()})
	}

	ctx := c.Context()

	acc, err := h.account.FindByID(ctx, auth.AccountID(c))
	if errors.Is(err, account.ErrNotFound) {
		return httperr.Unauthorized("다시 로그인해 주세요")
	}
	if err != nil {
		return err
	}
	if acc.Name != nil && *acc.Name == name {
		return c.JSON(proto.NameCheckResponse{Available: true})
	}

	taken, err := h.account.IsNameTaken(ctx, name)
	if err != nil {
		return err
	}
	if taken {
		return c.JSON(proto.NameCheckResponse{Available: false, Message: account.ErrNameTaken.Error()})
	}
	return c.JSON(proto.NameCheckResponse{Available: true})
}

// meResponse 는 계정을 읽어 GET /v1/me 응답으로 만든다.
func (h *Handler) meResponse(ctx context.Context, accountID int64) (proto.MeResponse, error) {
	acc, err := h.account.FindByID(ctx, accountID)
	if errors.Is(err, account.ErrNotFound) {
		return proto.MeResponse{}, httperr.Unauthorized("다시 로그인해 주세요")
	}
	if err != nil {
		return proto.MeResponse{}, err
	}

	avatarURL, err := h.imageURL(ctx, acc.AvatarKey)
	if err != nil {
		return proto.MeResponse{}, err
	}

	bannerURL, err := h.imageURL(ctx, acc.BannerKey)
	if err != nil {
		return proto.MeResponse{}, err
	}

	return proto.MeResponse{
		AccountID:    acc.ID,
		Name:         acc.Name,
		Nickname:     acc.Nickname,
		Bio:          acc.Bio,
		Color:        acc.Color,
		AvatarURL:    avatarURL,
		BannerURL:    bannerURL,
		NeedsProfile: acc.Name == nil,
	}, nil
}

// imageURL 은 R2 key 를 보기 주소로 바꾼다. 사진을 안 올렸으면(key 가 "") 빈 문자열.
func (h *Handler) imageURL(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	return h.r2.PresignGet(ctx, key)
}
