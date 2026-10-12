package main

import "time"

// Provider 는 어느 로그인 서비스로 가입했는지. DB 에는 글자 그대로 저장된다.
type Provider string

const (
	ProviderGoogle Provider = "google"
)

type Account struct {
	ID        uint     `gorm:"primaryKey"`
	Provider  Provider `gorm:"not null;uniqueIndex:idx_provider_subject"`
	Subject   string   `gorm:"not null;uniqueIndex:idx_provider_subject"` // 구글이면 id_token 의 sub
	Email     string   `gorm:"not null"`
	Name      string   `gorm:"not null"`
	Disabled  bool     `gorm:"not null;default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
