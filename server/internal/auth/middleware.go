package auth

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

const accountIDKey = "accountID"

func RequireAuth(tokenProvider *TokenProvider) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := c.Get("Authorization")
		token, ok := bearerToken(header)
		if !ok {
			return fiber.ErrUnauthorized
		}

		accountID, err := tokenProvider.ParseAccessToken(token)
		if err != nil {
			return fiber.ErrUnauthorized
		}

		c.Locals(accountIDKey, accountID)

		return c.Next()
	}
}

func bearerToken(header string) (string, bool) {
	token, ok := strings.CutPrefix(header, "Bearer ")
	if !ok || token == "" {
		return "", false
	}
	return token, true
}

func AccountID(c fiber.Ctx) int64 {
	id, _ := c.Locals(accountIDKey).(int64)
	return id
}
