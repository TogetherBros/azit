package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	Port               string
	DBDSN              string
	GoogleClientID     string
	GoogleClientSecret string
	JWTSecret          string
	AccessTTL          time.Duration
	RefreshTTL         time.Duration
	Dev                bool
	R2AccountID        string
	R2AccessKeyID      string
	R2SecretAccessKey  string
	R2Bucket           string
}

func Load() (Config, error) {
	if err := loadDotEnv(".env"); err != nil {
		return Config{}, err
	}

	cfg := Config{
		Port:               getenv("PORT", "8081"),
		DBDSN:              getenv("DB_DSN", "azit.db"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		AccessTTL:          15 * time.Minute,    // Step 3에서 ACCESS_TOKEN_TTL_SEC 로 바꿈
		RefreshTTL:         30 * 24 * time.Hour, // Step 3에서 REFRESH_TOKEN_TTL_SEC 로 바꿈
		Dev:                os.Getenv("DEV") == "1",
		R2AccountID:        os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKeyID:      os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey:  os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2Bucket:           os.Getenv("R2_BUCKET"),
	}

	// 빈 비밀키로도 JWT 서명은 되기 때문에, 아예 서버가 안 뜨게 막는다.
	if cfg.JWTSecret == "" {
		return Config{}, errors.New("JWT_SECRET 환경 변수가 비어 있어요")
	}

	if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" {
		return Config{}, errors.New("GOOGLE_CLIENT_ID / GOOGLE_CLIENT_SECRET 환경 변수가 비어 있어요")
	}

	if cfg.R2AccountID == "" || cfg.R2AccessKeyID == "" || cfg.R2SecretAccessKey == "" || cfg.R2Bucket == "" {
		return Config{}, errors.New("R2_ACCOUNT_ID / R2_ACCESS_KEY_ID / R2_SECRET_ACCESS_KEY / R2_BUCKET 환경 변수가 비어 있어요")
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
