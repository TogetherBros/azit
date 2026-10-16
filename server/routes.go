package main

import (
	"log"

	"github.com/TogetherBros/azit/server/internal/auth"
	"github.com/TogetherBros/azit/server/internal/config"
	"github.com/TogetherBros/azit/server/internal/httperr"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

// newApp 은 Fiber 앱을 만들고 미들웨어와 경로를 단다.
func newApp(cfg config.Config, h *handlers, tokenProvider *auth.TokenProvider) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: httperr.Handler,
	})

	// handler 안에서 panic 이 나도 서버가 죽지 않고 500 으로 돌려준다.
	app.Use(recover.New())

	app.Get("/healthz", healthz)

	requireAuth := auth.RequireAuth(tokenProvider)

	api := app.Group("/v1")
	h.auth.Routes(api.Group("/auth"))
	h.upload.Routes(api.Group("/uploads", requireAuth))
	h.profile.Routes(api.Group("/me", requireAuth))
	azits := api.Group("/azits", requireAuth)
	h.azit.Routes(azits)
	h.invite.AzitRoutes(azits)
	h.room.Routes(azits, h.requireMember)
	h.message.Routes(azits, h.requireMember)
	h.meme.Routes(azits, h.requireMember)
	h.invite.Routes(api.Group("/invites", requireAuth))

	if cfg.Dev {
		app.Get("/dev/token", h.auth.DevToken)
		log.Println("⚠️ DEV 모드: /dev/token 이 켜져 있어요")
	}

	return app
}

func healthz(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"ok": true})
}
