package main

import (
	"github.com/TogetherBros/azit/server/internal/account"
	"github.com/TogetherBros/azit/server/internal/auth"
	"github.com/TogetherBros/azit/server/internal/auth/refresh"
	"github.com/TogetherBros/azit/server/internal/azit"
	"github.com/TogetherBros/azit/server/internal/config"
	"github.com/TogetherBros/azit/server/internal/member"
	"github.com/TogetherBros/azit/server/internal/profile"
	"github.com/TogetherBros/azit/server/internal/storage"
	"github.com/TogetherBros/azit/server/internal/upload"
	"gorm.io/gorm"
)

// handlers 는 서버가 쓰는 handler 묶음이다.
type handlers struct {
	auth    *auth.Handler
	profile *profile.Handler
	upload  *upload.Handler
	azit    *azit.Handler
}

// newHandlers 는 repo 와 외부 클라이언트를 만들고 handler 에 꽂아 준다.
// 새 handler 가 생기면 여기서 만들고 routes.go 에서 경로를 단다.
func newHandlers(cfg config.Config, conn *gorm.DB, tokenProvider *auth.TokenProvider) (*handlers, error) {
	accountRepo := account.NewRepository(conn)
	refreshRepo := refresh.NewRepository(conn)
	memberRepo := member.NewRepository(conn)
	azitRepo := azit.NewRepository(conn)

	googleClient := auth.NewGoogleClient(cfg.GoogleClientID, cfg.GoogleClientSecret)
	r2, err := storage.NewR2(cfg.R2AccountID, cfg.R2AccessKeyID, cfg.R2SecretAccessKey, cfg.R2Bucket)
	if err != nil {
		return nil, err
	}

	return &handlers{
		auth:    auth.NewHandler(tokenProvider, refreshRepo, accountRepo, googleClient),
		profile: profile.NewHandler(accountRepo, r2),
		upload:  upload.NewHandler(r2),
		azit:    azit.NewHandler(accountRepo, memberRepo, azitRepo, r2),
	}, nil
}
