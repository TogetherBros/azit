package config

import (
	"errors"
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

func Load() (Config, error) {
	cfg := Config{
		Port:           getenv("PORT", "8081"),
		DBDSN:          getenv("DB_DSN", "azit.db"),
		GoogleClientID: os.Getenv("GOOGLE_CLIENT_ID"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		AccessTTL:      15 * time.Minute,    // Step 3에서 ACCESS_TOKEN_TTL_SEC 로 바꿈
		RefreshTTL:     30 * 24 * time.Hour, // Step 3에서 REFRESH_TOKEN_TTL_SEC 로 바꿈
	}

	// 빈 비밀키로도 JWT 서명은 되기 때문에, 아예 서버가 안 뜨게 막는다.
	if cfg.JWTSecret == "" {
		return Config{}, errors.New("JWT_SECRET 환경 변수가 비어 있어요")
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
