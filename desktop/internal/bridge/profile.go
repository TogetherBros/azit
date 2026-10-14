package bridge

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/TogetherBros/azit/desktop/internal/client"
	"github.com/TogetherBros/azit/proto"
)

// ProfileService 는 내 프로필 보기와 사진 올리기를 React 에 공개한다.
// 토큰은 AuthService 가 들고 있어서 거기서 빌려 쓴다.
type ProfileService struct {
	api  *client.Client
	auth *AuthService
}

func NewProfileService(api *client.Client, auth *AuthService) *ProfileService {
	return &ProfileService{api: api, auth: auth}
}

// Me 는 내 프로필을 서버에서 받아 온다.
func (s *ProfileService) Me(ctx context.Context) (proto.MeResponse, error) {
	return s.api.Me(ctx, s.auth.accessToken())
}

// UploadImage 는 사진 하나를 R2 에 올리고 내 프사(kind "avatar")나 배너("banner")로 저장한다.
// 1) 서버에 업로드 주소 받기 → 2) R2 에 바로 PUT → 3) 서버에 key 저장 (서버가 R2 에 진짜 있는지 확인)
// 사진 바이트는 React 에서 base64 문자열로 넘어온다.
func (s *ProfileService) UploadImage(ctx context.Context, kind, contentType, dataBase64 string) error {
	data, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		return fmt.Errorf("사진을 읽지 못했어요: %w", err)
	}

	token := s.auth.accessToken()

	up, err := s.api.CreateUpload(ctx, token, proto.UploadRequest{
		Kind:        kind,
		ContentType: contentType,
		Size:        int64(len(data)),
	})
	if err != nil {
		return err
	}

	if err := s.api.UploadToR2(ctx, up.URL, contentType, data); err != nil {
		return err
	}

	switch kind {
	case "avatar":
		return s.api.SetAvatar(ctx, token, up.Key)
	case "banner":
		return s.api.SetBanner(ctx, token, up.Key)
	default:
		return fmt.Errorf("알 수 없는 사진 종류예요: %s", kind)
	}
}
