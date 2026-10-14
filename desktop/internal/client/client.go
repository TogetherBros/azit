package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/TogetherBros/azit/proto"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
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
	err := c.postJSON(ctx, "/v1/auth/google", req, &out)
	return out, err
}

func (c *Client) postJSON(ctx context.Context, path string, in, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

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

	return json.NewDecoder(resp.Body).Decode(out)
}
