package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	app.Get("/healthz", healthz)

	log.Fatal(app.Listen(":8081"))
}

func healthz(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"ok": true})
}
