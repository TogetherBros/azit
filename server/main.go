package main

import (
	"log"

	"github.com/TogetherBros/azit/server/internal/account"
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

	if err := conn.AutoMigrate(&account.Account{}); err != nil {
		log.Fatal(err)
	}

	app := fiber.New()

	app.Get("/healthz", healthz)

	log.Printf("Listening on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}

func healthz(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"ok": true})
}
