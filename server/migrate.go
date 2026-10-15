package main

import (
	"github.com/TogetherBros/azit/server/internal/account"
	"github.com/TogetherBros/azit/server/internal/auth/refresh"
	"github.com/TogetherBros/azit/server/internal/azit"
	"github.com/TogetherBros/azit/server/internal/invite"
	"github.com/TogetherBros/azit/server/internal/member"
	"github.com/TogetherBros/azit/server/internal/room"
	"gorm.io/gorm"
)

// migrate 는 테이블을 만들거나 고친다. 새 모델이 생기면 여기에만 추가한다.
func migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&account.Account{},
		&refresh.Token{},
		&azit.Azit{},
		&member.Member{},
		&invite.Invite{},
		&room.Room{},
	)
}
