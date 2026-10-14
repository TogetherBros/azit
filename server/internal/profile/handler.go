package profile

import (
	"context"
	"errors"

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
	ctx := c.Context()

	acc, err := h.account.FindByID(ctx, auth.AccountID(c))
	if errors.Is(err, account.ErrNotFound) {
		return httperr.Unauthorized("다시 로그인해 주세요")
	}
	if err != nil {
		return err
	}

	avatarURL, err := h.imageURL(ctx, acc.AvatarKey)
	if err != nil {
		return err
	}

	bannerURL, err := h.imageURL(ctx, acc.BannerKey)
	if err != nil {
		return err
	}

	return c.JSON(proto.MeResponse{
		AccountID:    acc.ID,
		Name:         acc.Name,
		Nickname:     acc.Nickname,
		Bio:          acc.Bio,
		Color:        acc.Color,
		AvatarURL:    avatarURL,
		BannerURL:    bannerURL,
		NeedsProfile: acc.Name == nil,
	})
}

// imageURL 은 R2 key 를 보기 주소로 바꾼다. 사진을 안 올렸으면(key 가 "") 빈 문자열.
func (h *Handler) imageURL(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	return h.r2.PresignGet(ctx, key)
}
