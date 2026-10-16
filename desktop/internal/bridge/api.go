package bridge

import (
	"context"
	"fmt"
	"strings"

	"github.com/TogetherBros/azit/desktop/internal/client"
)

// APIService 는 아지트·방·메시지·짤 같은 서버 API 를 React 에 그대로 열어 준다.
// API 마다 Go 함수를 만들지 않고, React(api/*.ts)가 경로와 본문을 정해서 Call 로 부른다.
// 토큰 붙이기와 AT 갱신은 여기서 한다.
type APIService struct {
	api  *client.Client
	auth *AuthService
}

func NewAPIService(api *client.Client, auth *AuthService) *APIService {
	return &APIService{api: api, auth: auth}
}

// Call 은 서버 API 하나를 부른다. path 는 "/v1/..." 로 시작해야 한다.
// body 는 JSON 으로 보낼 값 (없으면 null), 응답은 JSON 을 푼 값 그대로 (본문이 없으면 null).
func (s *APIService) Call(ctx context.Context, method, path string, body any) (any, error) {
	if !strings.HasPrefix(path, "/v1/") {
		return nil, fmt.Errorf("잘못된 경로예요: %s", path)
	}

	var out any
	err := s.auth.withToken(ctx, func(token string) error {
		return s.api.Call(ctx, method, path, token, body, &out)
	})
	return out, err
}

// Upload 는 사진 하나를 R2 에 올리고 key 를 돌려준다. kind 는 "meme", "photo", "azit_banner".
// 받은 key 는 짤 등록(POST /v1/azits/{id}/memes), 사진 메시지 보내기, 아지트 만들기에 넣는다.
func (s *APIService) Upload(ctx context.Context, kind, contentType, dataBase64 string) (string, error) {
	if kind != "meme" && kind != "photo" && kind != "azit_banner" {
		return "", fmt.Errorf("알 수 없는 사진 종류예요: %s", kind)
	}
	return upload(ctx, s.api, s.auth, kind, contentType, dataBase64)
}
