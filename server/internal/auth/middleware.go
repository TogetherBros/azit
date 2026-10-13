package auth

import "github.com/gofiber/fiber/v3"

func RequireAuth(tokenProvider TokenProvider) fiber.Handler {
	return func(c fiber.Ctx) error {
		//TODO : 토큰 검사하는 로직 추가 예정

		return c.Next()
	}
}
