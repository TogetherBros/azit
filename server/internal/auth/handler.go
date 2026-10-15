package auth

import (
	"context"
	"errors"
	"strconv"

	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/account"
	"github.com/TogetherBros/azit/server/internal/auth/refresh"
	"github.com/TogetherBros/azit/server/internal/httperr"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	tokenProvider *TokenProvider
	refreshRepo   *refresh.Repository
	accountRepo   *account.Repository
	google        *GoogleClient
}

func NewHandler(tokenProvider *TokenProvider, refreshRepo *refresh.Repository, accountRepo *account.Repository, google *GoogleClient) *Handler {
	return &Handler{
		tokenProvider: tokenProvider,
		refreshRepo:   refreshRepo,
		accountRepo:   accountRepo,
		google:        google,
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
	r.Post("/logout", h.Logout)
	r.Post("/google", h.Google)
}

// Logout 은 받은 RT 를 지워 더는 갱신할 수 없게 한다.
// 이미 없는 RT 여도 성공으로 본다 (두 번 눌러도 같은 결과).
func (h *Handler) Logout(c fiber.Ctx) error {
	var req proto.RefreshRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}

	hash := h.tokenProvider.HashRefreshToken(req.RefreshToken)

	if err := h.refreshRepo.Delete(c.Context(), hash); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
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
	if errors.Is(err, refresh.ErrInvalid) {
		return httperr.Unauthorized("다시 로그인해 주세요")
	}
	if err != nil {
		return err
	}

	// 멈춘 계정이면 갱신해 주지 않는다
	acc, err := h.accountRepo.FindByID(ctx, accountID)
	if errors.Is(err, account.ErrNotFound) {
		return httperr.Unauthorized("다시 로그인해 주세요")
	}
	if err != nil {
		return err
	}
	if acc.Disabled {
		return httperr.New(fiber.StatusForbidden, proto.ErrCodeAccountDisabled, "멈춘 계정이에요. 아지트 친구에게 물어봐 주세요")
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

// Google 은 앱이 받은 구글 code 로 로그인한다. 처음 온 사람이면 계정을 만든다.
func (h *Handler) Google(c fiber.Ctx) error {
	var req proto.GoogleLoginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}

	ctx := c.Context()

	// 1. code → id_token
	rawIDToken, err := h.google.Exchange(ctx, req.Code, req.CodeVerifier, req.RedirectURI)
	if errors.Is(err, ErrGoogleExchange) {
		return httperr.New(fiber.StatusBadRequest, proto.ErrCodeGoogleExchangeFailed, "구글 로그인을 확인하지 못했어요. 다시 해 주세요")
	}
	if err != nil {
		return err
	}

	// 2. id_token 검증 → 사용자 정보
	user, err := h.google.Verify(ctx, rawIDToken)
	if errors.Is(err, ErrGoogleIDToken) {
		return httperr.New(fiber.StatusBadRequest, proto.ErrCodeGoogleExchangeFailed, "구글 로그인을 확인하지 못했어요. 다시 해 주세요")
	}
	if err != nil {
		return err
	}

	// 3. 계정 찾기, 없으면 만들기
	acc, err := h.accountRepo.FindByProviderSubject(ctx, account.ProviderGoogle, user.Subject)
	if errors.Is(err, account.ErrNotFound) {
		acc = account.Account{
			Provider: account.ProviderGoogle,
			Subject:  user.Subject,
			Email:    user.Email,
		}
		err = h.accountRepo.Create(ctx, &acc)
	}
	if err != nil {
		return err
	}

	// 4. 멈춘 계정은 막기
	if acc.Disabled {
		return httperr.New(fiber.StatusForbidden, proto.ErrCodeAccountDisabled, "멈춘 계정이에요. 아지트 친구에게 물어봐 주세요")
	}

	// 5. 토큰 발급
	tokens, err := h.issueTokens(ctx, acc.ID)
	if err != nil {
		return err
	}

	return c.JSON(tokens)
}
