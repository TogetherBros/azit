package main

import (
	"log"

	"github.com/TogetherBros/azit/server/internal/account"
	"github.com/TogetherBros/azit/server/internal/auth"
	"github.com/TogetherBros/azit/server/internal/auth/refresh"
	"github.com/TogetherBros/azit/server/internal/config"
	"github.com/TogetherBros/azit/server/internal/db"
	"github.com/TogetherBros/azit/server/internal/httperr"
	"github.com/TogetherBros/azit/server/internal/profile"
	"github.com/TogetherBros/azit/server/internal/storage"
	"github.com/TogetherBros/azit/server/internal/upload"
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

	app := fiber.New(fiber.Config{
		ErrorHandler: httperr.Handler,
	})

	app.Get("/healthz", healthz)

	refreshRepo := refresh.NewRepository(conn)
	accountRepo := account.NewRepository(conn)
	googleClient := auth.NewGoogleClient(cfg.GoogleClientID, cfg.GoogleClientSecret)
	authHandler := auth.NewHandler(tokenProvider, refreshRepo, accountRepo, googleClient)

	api := app.Group("/v1")
	authHandler.Routes(api.Group("/auth"))

	r2, err := storage.NewR2(cfg.R2AccountID, cfg.R2AccessKeyID, cfg.R2SecretAccessKey, cfg.R2Bucket)
	if err != nil {
		log.Fatal(err)
	}
	uploadHandler := upload.NewHandler(r2)
	profileHandler := profile.NewHandler(accountRepo, r2)

	api.Post("/uploads", auth.RequireAuth(tokenProvider), uploadHandler.Create)
	api.Get("/me", auth.RequireAuth(tokenProvider), profileHandler.Me)
	api.Put("/me/avatar", auth.RequireAuth(tokenProvider), profileHandler.SaveAvatar)
	api.Put("/me/banner", auth.RequireAuth(tokenProvider), profileHandler.SaveBanner)

	if cfg.Dev {
		app.Get("/dev/token", authHandler.DevToken)
		log.Println("⚠️ DEV 모드: /dev/token 이 켜져 있어요")
	}
	log.Printf("Listening on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}

func healthz(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"ok": true})
}
