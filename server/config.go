package main

import (
	"os"
	"time"
)

type Config struct {
	Port           string
	DBDSN          string
	GoogleClientID string
	JWTSecret      string
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
}

func LoadConfig() Config {
	return Config{
		Port:           getenv("PORT", "8081"),
		DBDSN:          os.Getenv("DB_DSN"),
		GoogleClientID: os.Getenv("GOOGLE_CLIENT_ID"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		AccessTTL:      15 * time.Minute,    // Step 3에서 ACCESS_TOKEN_TTL_SEC 로 바꿈
		RefreshTTL:     30 * 24 * time.Hour, // Step 3에서 REFRESH_TOKEN_TTL_SEC 로 바꿈
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
