package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/TogetherBros/azit/proto"
)

type Client struct {
	baseURL string
	http    *http.Client
	upload  *http.Client // R2 업로드용. 사진이 커서 시간을 넉넉히 준다
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
		upload:  &http.Client{Timeout: 60 * time.Second},
	}
}

type APIError struct {
	Status  int
	Code    string
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}

func (c *Client) GoogleLogin(ctx context.Context, req proto.GoogleLoginRequest) (proto.TokenResponse, error) {
	var out proto.TokenResponse
	err := c.do(ctx, http.MethodPost, "/v1/auth/google", "", req, &out)
	return out, err
}

// Me 는 내 프로필을 받는다. 사진은 24시간짜리 보기 주소로 온다.
func (c *Client) Me(ctx context.Context, token string) (proto.MeResponse, error) {
	var out proto.MeResponse
	err := c.do(ctx, http.MethodGet, "/v1/me", token, nil, &out)
	return out, err
}

// UpdateMe 는 내 프로필(이름, 닉네임, 소개, 색)을 저장하고, 저장된 프로필을 받는다.
func (c *Client) UpdateMe(ctx context.Context, token string, req proto.UpdateMeRequest) (proto.MeResponse, error) {
	var out proto.MeResponse
	err := c.do(ctx, http.MethodPatch, "/v1/me", token, req, &out)
	return out, err
}

// CheckName 은 그 이름(@ 뒤)을 내가 쓸 수 있는지 물어본다.
func (c *Client) CheckName(ctx context.Context, token, name string) (proto.NameCheckResponse, error) {
	var out proto.NameCheckResponse
	err := c.do(ctx, http.MethodGet, "/v1/me/name-check?name="+url.QueryEscape(name), token, nil, &out)
	return out, err
}

// Logout 은 서버에 RT 를 버리라고 알린다. 그 RT 로는 더 갱신할 수 없다.
func (c *Client) Logout(ctx context.Context, refreshToken string) error {
	return c.do(ctx, http.MethodPost, "/v1/auth/logout", "", proto.RefreshRequest{RefreshToken: refreshToken}, nil)
}

// CreateUpload 는 서버에서 R2 업로드 주소(10분짜리)와 key 를 받는다.
func (c *Client) CreateUpload(ctx context.Context, token string, req proto.UploadRequest) (proto.UploadResponse, error) {
	var out proto.UploadResponse
	err := c.do(ctx, http.MethodPost, "/v1/uploads", token, req, &out)
	return out, err
}

// SetAvatar, SetBanner 는 R2 에 올린 파일의 key 를 내 프사·배너로 저장한다.
func (c *Client) SetAvatar(ctx context.Context, token, key string) error {
	return c.do(ctx, http.MethodPut, "/v1/me/avatar", token, proto.SetImageRequest{Key: key}, nil)
}

func (c *Client) SetBanner(ctx context.Context, token, key string) error {
	return c.do(ctx, http.MethodPut, "/v1/me/banner", token, proto.SetImageRequest{Key: key}, nil)
}

// UploadToR2 는 서버가 준 주소로 파일을 바로 올린다. 서버를 거치지 않는다.
// 주소 서명에 형식과 크기가 들어 있어서, 요청할 때 말한 것과 똑같이 보내야 한다.
func (c *Client) UploadToR2(ctx context.Context, url, contentType string, data []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := c.upload.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("사진을 올리지 못했어요 (R2 %s): %s", resp.Status, msg)
	}
	return nil
}

// do 는 서버 API 한 번 부르기. token 이 있으면 Authorization 헤더를 붙인다.
// in 이 nil 이면 본문 없이, out 이 nil 이면 응답 본문을 읽지 않는다.
func (c *Client) do(ctx context.Context, method, path, token string, in, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var e proto.ErrorResponse
		if json.NewDecoder(resp.Body).Decode(&e) == nil && e.Error.Message != "" {
			return &APIError{
				Status:  resp.StatusCode,
				Code:    e.Error.Code,
				Message: e.Error.Message,
			}
		}
		return fmt.Errorf("API error: %s", resp.Status)
	}

	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
