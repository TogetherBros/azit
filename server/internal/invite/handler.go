package invite

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"time"

	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/account"
	"github.com/TogetherBros/azit/server/internal/auth"
	"github.com/TogetherBros/azit/server/internal/azit"
	"github.com/TogetherBros/azit/server/internal/httperr"
	"github.com/TogetherBros/azit/server/internal/member"
	"github.com/TogetherBros/azit/server/internal/storage"
	"github.com/gofiber/fiber/v3"
)

// expiresInOptions 는 고를 수 있는 만료 시간. 0 은 무제한.
var expiresInOptions = map[string]time.Duration{
	"30m":   30 * time.Minute,
	"1h":    time.Hour,
	"1d":    24 * time.Hour,
	"7d":    7 * 24 * time.Hour,
	"never": 0,
}

const defaultExpiresIn = "7d"

// maxUsesOptions 는 고를 수 있는 사용 횟수. 0 은 무제한.
var maxUsesOptions = []int{0, 1, 5, 10}

type Handler struct {
	inviteRepo  *Repository
	azitRepo    *azit.Repository
	memberRepo  *member.Repository
	accountRepo *account.Repository
	r2          *storage.R2
}

func NewHandler(inviteRepo *Repository, azitRepo *azit.Repository, memberRepo *member.Repository, accountRepo *account.Repository, r2 *storage.R2) *Handler {
	return &Handler{
		inviteRepo:  inviteRepo,
		azitRepo:    azitRepo,
		memberRepo:  memberRepo,
		accountRepo: accountRepo,
		r2:          r2,
	}
}

// AzitRoutes 는 /v1/azits 그룹 아래에 아지트별 초대 경로를 붙인다.
func (h *Handler) AzitRoutes(r fiber.Router) {
	r.Post("/:id/invites", h.Create)
	r.Get("/:id/invites", h.List)
}

// Routes 는 /v1/invites 아래 경로를 붙인다. r 에는 인증 미들웨어가 걸려 있어야 한다.
func (h *Handler) Routes(r fiber.Router) {
	r.Get("/:code", h.Preview)
	r.Post("/:code/join", h.Join)
	r.Delete("/:code", h.Revoke)
}

// Create 는 초대 코드를 만든다. 관리자 이상만. POST /v1/azits/{id}/invites
func (h *Handler) Create(c fiber.Ctx) error {
	azitID, err := azitIDParam(c)
	if err != nil {
		return err
	}

	var req proto.CreateInviteRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}

	if req.ExpiresIn == "" {
		req.ExpiresIn = defaultExpiresIn
	}
	expiresIn, ok := expiresInOptions[req.ExpiresIn]
	if !ok {
		return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, "고를 수 없는 만료 시간이에요")
	}
	if !slices.Contains(maxUsesOptions, req.MaxUses) {
		return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, "고를 수 없는 사용 횟수예요")
	}

	ctx := c.Context()
	accountID := auth.AccountID(c)

	if _, err := h.requireAdmin(ctx, azitID, accountID); err != nil {
		return err
	}

	now := time.Now().UTC()
	inv := Invite{
		AzitID:    azitID,
		CreatedBy: accountID,
		MaxUses:   req.MaxUses,
	}
	if expiresIn > 0 {
		expiresAt := now.Add(expiresIn)
		inv.ExpiresAt = &expiresAt
	}

	if err := h.inviteRepo.Create(ctx, &inv); err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(toResponse(inv))
}

// List 는 그 아지트에서 지금 쓸 수 있는 초대 코드 목록. 관리자 이상만. GET /v1/azits/{id}/invites
func (h *Handler) List(c fiber.Ctx) error {
	azitID, err := azitIDParam(c)
	if err != nil {
		return err
	}

	ctx := c.Context()
	if _, err := h.requireAdmin(ctx, azitID, auth.AccountID(c)); err != nil {
		return err
	}

	invites, err := h.inviteRepo.ListUsable(ctx, azitID, time.Now().UTC())
	if err != nil {
		return err
	}

	res := make([]proto.InviteResponse, 0, len(invites))
	for _, inv := range invites {
		res = append(res, toResponse(inv))
	}
	return c.JSON(res)
}

// Revoke 는 초대 코드를 취소한다. 만든 사람이나 관리자 이상만. DELETE /v1/invites/{code}
func (h *Handler) Revoke(c fiber.Ctx) error {
	ctx := c.Context()
	accountID := auth.AccountID(c)

	inv, err := h.inviteRepo.FindByCode(ctx, NormalizeCode(c.Params("code")))
	if errors.Is(err, ErrNotFound) {
		return httperr.New(fiber.StatusNotFound, proto.ErrCodeNotFound, err.Error())
	}
	if err != nil {
		return err
	}

	m, err := h.memberRepo.FindActive(ctx, inv.AzitID, accountID)
	if errors.Is(err, member.ErrNotMember) {
		// 멤버가 아니면 코드가 있다는 것도 알려주지 않는다.
		return httperr.New(fiber.StatusNotFound, proto.ErrCodeNotFound, ErrNotFound.Error())
	}
	if err != nil {
		return err
	}
	if inv.CreatedBy != accountID && !m.Role.AtLeast(member.RoleAdmin) {
		return httperr.New(fiber.StatusForbidden, proto.ErrCodeForbidden, "만든 사람이나 관리자만 취소할 수 있어요")
	}

	if err := h.inviteRepo.Revoke(ctx, inv.InviteID, time.Now().UTC()); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// Preview 는 들어가기 전에 어떤 아지트인지 보여준다. GET /v1/invites/{code}
func (h *Handler) Preview(c fiber.Ctx) error {
	ctx := c.Context()
	accountID := auth.AccountID(c)

	inv, err := h.inviteRepo.FindByCode(ctx, NormalizeCode(c.Params("code")))
	if errors.Is(err, ErrNotFound) {
		return httperr.New(fiber.StatusNotFound, proto.ErrCodeNotFound, err.Error())
	}
	if err != nil {
		return err
	}

	_, err = h.memberRepo.FindActive(ctx, inv.AzitID, accountID)
	alreadyMember := err == nil
	if err != nil && !errors.Is(err, member.ErrNotMember) {
		return err
	}
	// 이미 들어와 있으면 코드가 다 쓰였어도 미리보기는 보여준다 (앱이 "이미 멤버" 로 안내).
	if !alreadyMember && !inv.Usable(time.Now().UTC()) {
		return httperr.New(fiber.StatusNotFound, proto.ErrCodeInviteUnusable, ErrUnusable.Error())
	}

	a, err := h.azitRepo.FindByID(ctx, inv.AzitID)
	if err != nil {
		return err
	}
	azitRes, err := h.azitResponse(ctx, a)
	if err != nil {
		return err
	}

	inviter, err := h.inviterNickname(ctx, inv)
	if err != nil {
		return err
	}

	return c.JSON(proto.InvitePreviewResponse{
		Code:            FormatCode(inv.Code),
		Azit:            azitRes,
		InviterNickname: inviter,
		AlreadyMember:   alreadyMember,
	})
}

// Join 은 초대 코드로 아지트에 들어간다. POST /v1/invites/{code}/join
func (h *Handler) Join(c fiber.Ctx) error {
	ctx := c.Context()
	accountID := auth.AccountID(c)

	acc, err := h.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return err
	}

	azitID, err := h.inviteRepo.Join(ctx, JoinInput{
		Code:      NormalizeCode(c.Params("code")),
		AccountID: accountID,
		Nickname:  acc.Nickname,
		Now:       time.Now().UTC(),
	})
	switch {
	case errors.Is(err, ErrNotFound):
		return httperr.New(fiber.StatusNotFound, proto.ErrCodeNotFound, err.Error())
	case errors.Is(err, ErrUnusable):
		return httperr.New(fiber.StatusNotFound, proto.ErrCodeInviteUnusable, err.Error())
	case errors.Is(err, ErrAlreadyMember):
		return httperr.New(fiber.StatusConflict, proto.ErrCodeAlreadyMember, err.Error())
	case err != nil:
		return err
	}

	a, err := h.azitRepo.FindByID(ctx, azitID)
	if err != nil {
		return err
	}
	res, err := h.azitResponse(ctx, a)
	if err != nil {
		return err
	}
	return c.JSON(res)
}

// requireAdmin 은 지금 들어와 있는 관리자 이상인지 본다.
// 멤버가 아니면 404 (아지트가 있다는 것도 숨김), 일반 멤버면 403.
func (h *Handler) requireAdmin(ctx context.Context, azitID, accountID int64) (member.Member, error) {
	m, err := h.memberRepo.FindActive(ctx, azitID, accountID)
	if errors.Is(err, member.ErrNotMember) {
		return member.Member{}, httperr.New(fiber.StatusNotFound, proto.ErrCodeNotFound, azit.ErrNotFound.Error())
	}
	if err != nil {
		return member.Member{}, err
	}
	if !m.Role.AtLeast(member.RoleAdmin) {
		return member.Member{}, httperr.New(fiber.StatusForbidden, proto.ErrCodeForbidden, "관리자만 할 수 있어요")
	}
	return m, nil
}

// inviterNickname 은 초대 만든 사람의 아지트 안 닉네임. 그 사람이 나갔으면 계정 닉네임.
func (h *Handler) inviterNickname(ctx context.Context, inv Invite) (string, error) {
	m, err := h.memberRepo.FindActive(ctx, inv.AzitID, inv.CreatedBy)
	if err == nil {
		return m.NickName, nil
	}
	if !errors.Is(err, member.ErrNotMember) {
		return "", err
	}

	acc, err := h.accountRepo.FindByID(ctx, inv.CreatedBy)
	if errors.Is(err, account.ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return acc.Nickname, nil
}

// azitResponse 는 아지트 한 개를 응답 모양으로 바꾼다 (인원 수, 배너 링크 포함).
func (h *Handler) azitResponse(ctx context.Context, a azit.Azit) (proto.AzitResponse, error) {
	counts, err := h.memberRepo.CountActive(ctx, []int64{a.AzitID})
	if err != nil {
		return proto.AzitResponse{}, err
	}

	bannerURL := ""
	if a.BannerKey != "" {
		bannerURL, err = h.r2.PresignGet(ctx, a.BannerKey)
		if err != nil {
			return proto.AzitResponse{}, err
		}
	}

	return proto.AzitResponse{
		AzitID:      a.AzitID,
		Name:        a.Name,
		Color:       a.Color,
		Description: a.Description,
		BannerURL:   bannerURL,
		MemberCount: counts[a.AzitID],
	}, nil
}

func toResponse(inv Invite) proto.InviteResponse {
	return proto.InviteResponse{
		Code:      FormatCode(inv.Code),
		AzitID:    inv.AzitID,
		CreatedBy: inv.CreatedBy,
		ExpiresAt: inv.ExpiresAt,
		MaxUses:   inv.MaxUses,
		Uses:      inv.Uses,
		CreatedAt: inv.CreatedAt,
	}
}

// azitIDParam 은 경로의 {id} 를 숫자로 읽는다.
func azitIDParam(c fiber.Ctx) (int64, error) {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, fiber.ErrBadRequest
	}
	return id, nil
}
