package main

import (
	"log"
	"strconv"

	"github.com/TogetherBros/azit/server/internal/account"
	"github.com/TogetherBros/azit/server/internal/auth"
	"github.com/TogetherBros/azit/server/internal/auth/refresh"
	"github.com/TogetherBros/azit/server/internal/config"
	"github.com/TogetherBros/azit/server/internal/db"
	"github.com/gofiber/fiber/v3"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	conn, err := db.Open(cfg.DBDSN)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("DB 연결 성공: %s", cfg.DBDSN)

	if err := conn.AutoMigrate(&account.Account{}, &refresh.Token{}); err != nil {
		log.Fatal(err)
	}

	tokenProvider := auth.NewTokenProvider(cfg.JWTSecret, cfg.AccessTTL, cfg.RefreshTTL)

	app := fiber.New()

	app.Get("/healthz", healthz)
	app.Get("/v1/me", auth.RequireAuth(tokenProvider), me)

	if cfg.Dev {
		app.Get("/dev/token", devToken(tokenProvider))
		log.Println("⚠️ DEV 모드: /dev/token 이 켜져 있어요")
	}

	log.Printf("Listening on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}

func healthz(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"ok": true})
}

func me(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"account_id": auth.AccountID(c)})
}

func devToken(tokenProvider *auth.TokenProvider) fiber.Handler {
	return func(c fiber.Ctx) error {
		accountID, err := strconv.ParseInt(c.Query("account_id"), 10, 64)
		if err != nil {
			return fiber.ErrBadRequest
		}

		at, err := tokenProvider.IssueAccessToken(accountID)
		if err != nil {
			return err
		}

		return c.JSON(fiber.Map{"access_token": at})
	}
}
