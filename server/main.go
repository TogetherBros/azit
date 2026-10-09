package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
)

func main() {
	cfg := LoadConfig()

	app := fiber.New()

	app.Get("/healthz", healthz)

	log.Printf("Listening on port %d", cfg.Port)
	log.Fatal(app.Listen(":8081"))
}

func healthz(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"ok": true})
}
