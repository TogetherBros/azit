package bridge

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/TogetherBros/azit/desktop/internal/client"
	"github.com/TogetherBros/azit/proto"
)

// ProfileService 는 내 프로필 보기·고치기와 사진 올리기를 React 에 공개한다.
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
	var out proto.MeResponse
	err := s.auth.withToken(ctx, func(token string) (err error) {
		out, err = s.api.Me(ctx, token)
		return err
	})
	return out, err
}

// UpdateMe 는 내 프로필(이름, 닉네임, 소개, 색)을 저장하고, 저장된 프로필을 돌려준다.
func (s *ProfileService) UpdateMe(ctx context.Context, req proto.UpdateMeRequest) (proto.MeResponse, error) {
	var out proto.MeResponse
	err := s.auth.withToken(ctx, func(token string) (err error) {
		out, err = s.api.UpdateMe(ctx, token, req)
		return err
	})
	return out, err
}

// CheckName 은 그 이름을 쓸 수 있는지 알려 준다. 지금 내 이름이면 쓸 수 있다고 나온다.
func (s *ProfileService) CheckName(ctx context.Context, name string) (proto.NameCheckResponse, error) {
	var out proto.NameCheckResponse
	err := s.auth.withToken(ctx, func(token string) (err error) {
		out, err = s.api.CheckName(ctx, token, name)
		return err
	})
	return out, err
}

// UploadImage 는 사진 하나를 R2 에 올리고 내 프사(kind "avatar")나 배너("banner")로 저장한다.
// 1) 서버에 업로드 주소 받기 → 2) R2 에 바로 PUT → 3) 서버에 key 저장 (서버가 R2 에 진짜 있는지 확인)
// 사진 바이트는 React 에서 base64 문자열로 넘어온다.
func (s *ProfileService) UploadImage(ctx context.Context, kind, contentType, dataBase64 string) error {
	if kind != "avatar" && kind != "banner" {
		return fmt.Errorf("알 수 없는 사진 종류예요: %s", kind)
	}

	key, err := upload(ctx, s.api, s.auth, kind, contentType, dataBase64)
	if err != nil {
		return err
	}

	return s.auth.withToken(ctx, func(token string) error {
		if kind == "avatar" {
			return s.api.SetAvatar(ctx, token, key)
		}
		return s.api.SetBanner(ctx, token, key)
	})
}

// upload 는 사진 하나를 R2 에 올리고 key 를 돌려준다 (내 사진, 짤, 아지트 배너 공통).
// 1) 서버에 업로드 주소 받기 → 2) R2 에 바로 PUT. key 를 어디에 저장할지는 부른 쪽이 정한다.
func upload(ctx context.Context, api *client.Client, auth *AuthService, kind, contentType, dataBase64 string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		return "", fmt.Errorf("사진을 읽지 못했어요: %w", err)
	}

	var up proto.UploadResponse
	err = auth.withToken(ctx, func(token string) (err error) {
		up, err = api.CreateUpload(ctx, token, proto.UploadRequest{
			Kind:        kind,
			ContentType: contentType,
			Size:        int64(len(data)),
		})
		return err
	})
	if err != nil {
		return "", err
	}

	if err := api.UploadToR2(ctx, up.URL, contentType, data); err != nil {
		return "", err
	}
	return up.Key, nil
}
