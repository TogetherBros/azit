package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := OpenDB(cfg.DBDSN)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("DB 연결 성공: %s", cfg.DBDSN)
	_ = db // A2-3 에서 AutoMigrate 에 쓰면 지운다

	app := fiber.New()

	app.Get("/healthz", healthz)

	log.Printf("Listening on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}

func healthz(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"ok": true})
}
