package bridge

import (
	"context"
	"log"
	"sync"

	"github.com/TogetherBros/azit/desktop/internal/client"
	"github.com/TogetherBros/azit/desktop/internal/login"
	"github.com/TogetherBros/azit/proto"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// AuthService 는 로그인 관련 기능을 React 에 공개한다.
// 실제 일은 login(브라우저 로그인)과 client(서버 호출)에 맡긴다.
type AuthService struct {
	api      *client.Client
	clientID string

	mu     sync.Mutex
	tokens proto.TokenResponse
}

func NewAuthService(api *client.Client, clientID string) *AuthService {
	return &AuthService{api: api, clientID: clientID}
}

// Login 은 브라우저로 구글 로그인을 하고, 서버에서 토큰을 받아 보관한다.
// React 에서 호출을 취소하면 ctx 가 취소되어 기다리던 로그인도 멈춘다.
func (s *AuthService) Login(ctx context.Context) error {
	grant, err := login.Authorize(ctx, s.clientID, application.Get().Browser.OpenURL)
	if err != nil {
		return err
	}

	tokens, err := s.api.GoogleLogin(ctx, proto.GoogleLoginRequest{
		Code:         grant.Code,
		CodeVerifier: grant.Verifier,
		RedirectURI:  grant.RedirectURI,
	})
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.tokens = tokens
	s.mu.Unlock()
	return nil
}

// IsLoggedIn 은 지금 토큰을 들고 있는지 알려 준다.
func (s *AuthService) IsLoggedIn() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tokens.AccessToken != ""
}

// Logout 은 들고 있던 토큰을 버린다. 서버에 RT 폐기를 알리는 건 실패해도 로그아웃은 된다
// (서버가 꺼져 있어도 앱에서는 나갈 수 있어야 해서). 그 RT 는 서버에서 기한이 지나면 못 쓴다.
func (s *AuthService) Logout(ctx context.Context) {
	s.mu.Lock()
	rt := s.tokens.RefreshToken
	s.tokens = proto.TokenResponse{}
	s.mu.Unlock()

	if rt == "" {
		return
	}
	if err := s.api.Logout(ctx, rt); err != nil {
		log.Printf("로그아웃: 서버에 RT 폐기를 못 알렸어요: %v", err)
	}
}

// accessToken 은 지금 들고 있는 AT. 로그인 전이면 "".
func (s *AuthService) accessToken() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tokens.AccessToken
}
