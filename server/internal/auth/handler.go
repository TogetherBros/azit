package auth

import (
	"context"
	"strconv"

	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/auth/refresh"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	tokenProvider *TokenProvider
	refreshRepo   *refresh.Repository
}

func NewHandler(tokenProvider *TokenProvider, refreshRepo *refresh.Repository) *Handler {
	return &Handler{
		tokenProvider: tokenProvider,
		refreshRepo:   refreshRepo,
	}
}

// issueTokens 는 AT 와 RT 를 새로 만들고, RT 는 DB 에 저장한다.
func (h *Handler) issueTokens(ctx context.Context, accountID int64) (proto.TokenResponse, error) {
	at, err := h.tokenProvider.IssueAccessToken(accountID)
	if err != nil {
		return proto.TokenResponse{}, err
	}

	plain, hash, expiresAt := h.tokenProvider.NewRefreshToken()

	if err := h.refreshRepo.Save(ctx, accountID, hash, expiresAt); err != nil {
		return proto.TokenResponse{}, err
	}

	return proto.TokenResponse{
		AccessToken:  at,
		RefreshToken: plain,
	}, nil
}

// Routes 는 /v1/auth 아래 경로를 붙인다.
func (h *Handler) Routes(r fiber.Router) {
	r.Post("/refresh", h.Refresh)
}

// Refresh 는 RT 를 받아 새 AT·RT 를 준다. 쓴 RT 는 지워진다 (회전).
func (h *Handler) Refresh(c fiber.Ctx) error {
	var req proto.RefreshRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}

	ctx := c.Context()

	// DB 에는 해시만 있으니 해시로 찾는다
	hash := h.tokenProvider.HashRefreshToken(req.RefreshToken)

	accountID, err := h.refreshRepo.Consume(ctx, hash)
	if err != nil {
		return err
	}

	tokens, err := h.issueTokens(ctx, accountID)
	if err != nil {
		return err
	}

	return c.JSON(tokens)
}

func (h *Handler) DevToken(c fiber.Ctx) error {
	accountID, err := strconv.ParseInt(c.Query("account_id"), 10, 64)
	if err != nil {
		return fiber.ErrBadRequest
	}

	tokens, err := h.issueTokens(c.Context(), accountID)
	if err != nil {
		return err
	}

	return c.JSON(tokens)
}
