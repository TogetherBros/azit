package main

import (
	"log"

	"github.com/TogetherBros/azit/server/internal/auth"
	"github.com/TogetherBros/azit/server/internal/config"
	"github.com/TogetherBros/azit/server/internal/db"
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

	if err := migrate(conn); err != nil {
		log.Fatal(err)
	}

	tokenProvider := auth.NewTokenProvider(cfg.JWTSecret, cfg.AccessTTL, cfg.RefreshTTL)

	h, err := newHandlers(cfg, conn, tokenProvider)
	if err != nil {
		log.Fatal(err)
	}

	app := newApp(cfg, h, tokenProvider)

	log.Printf("Listening on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
