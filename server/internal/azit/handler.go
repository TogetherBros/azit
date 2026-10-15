package azit

import (
	"context"
	"errors"

	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/account"
	"github.com/TogetherBros/azit/server/internal/auth"
	"github.com/TogetherBros/azit/server/internal/httperr"
	"github.com/TogetherBros/azit/server/internal/member"
	"github.com/TogetherBros/azit/server/internal/storage"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	accountRepo *account.Repository
	memberRepo  *member.Repository
	azitRepo    *Repository
	r2          *storage.R2
}

func NewHandler(accountRepo *account.Repository, memberRepo *member.Repository, azitRepo *Repository, r2 *storage.R2) *Handler {
	return &Handler{
		accountRepo: accountRepo,
		memberRepo:  memberRepo,
		azitRepo:    azitRepo,
		r2:          r2,
	}
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

		res = append(res, proto.AzitResponse{
			AzitID:      a.AzitID,
			Name:        a.Name,
			Color:       a.Color,
			BannerURL:   bannerURL,
			MemberCount: counts[a.AzitID],
		})
	}

	return c.JSON(res)
}

// bannerURL 은 배너 키로 보기 링크를 만든다. 키가 없으면 "".
func (h *Handler) bannerURL(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	return h.r2.PresignGet(ctx, key)
}

// Routes 는 /v1/azits 아래 경로를 붙인다. r 에는 인증 미들웨어가 걸려 있어야 한다.
func (h *Handler) Routes(r fiber.Router) {
	r.Post("/", h.Create)
	r.Get("/", h.List)
}
