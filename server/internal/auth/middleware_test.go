package auth

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

func TestRequireAuth_ValidToken(t *testing.T) {
	p := NewTokenProvider("test-secret", time.Minute, time.Hour)
	token, _ := p.IssueAccessToken(14)

	app := fiber.New()
	app.Get("/me", RequireAuth(p), func(c fiber.Ctx) error {
		return c.SendString("통과")
	})

	req := httptest.NewRequest("GET", "/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("요청 실패: %v", err)
	}
	if res.StatusCode != 200 {
		t.Errorf("status = %d, want 200", res.StatusCode)
	}
}
