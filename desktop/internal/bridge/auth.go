package bridge

import (
	"context"
	"errors"
	"log"
	"net/http"
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

	refreshMu sync.Mutex // 갱신은 한 번에 하나만
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

// withToken 은 AT 를 붙여 서버를 부른다. AT 가 만료돼 401 이 오면 RT 로 한 번 갱신하고 다시 부른다.
// (AT 는 15분짜리라, 앱을 켜 두면 금방 만료된다)
func (s *AuthService) withToken(ctx context.Context, call func(token string) error) error {
	token := s.accessToken()
	err := call(token)

	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusUnauthorized {
		return err
	}

	if err := s.refresh(ctx, token); err != nil {
		return err
	}
	return call(s.accessToken())
}

// refresh 는 RT 로 토큰을 새로 받는다. 화면이 여러 API 를 한꺼번에 불러 동시에 401 이 나도
// 한 번만 갱신하도록, 그 사이 다른 쪽이 이미 갱신했으면(AT 가 바뀌었으면) 그냥 돌아간다.
func (s *AuthService) refresh(ctx context.Context, usedToken string) error {
	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()

	s.mu.Lock()
	cur := s.tokens
	s.mu.Unlock()

	if cur.AccessToken != usedToken {
		return nil
	}
	if cur.RefreshToken == "" {
		return errors.New("다시 로그인해 주세요")
	}

	tokens, err := s.api.Refresh(ctx, cur.RefreshToken)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.tokens = tokens
	s.mu.Unlock()
	return nil
}
