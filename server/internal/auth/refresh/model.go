package refresh

import "time"

// Token 은 발급한 RT 하나. 원문은 저장하지 않고 해시만 둔다.
type Token struct {
	ID        int64     `gorm:"primaryKey"`
	AccountID int64     `gorm:"not null;index"`
	TokenHash string    `gorm:"not null;uniqueIndex"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
}

// TableName 이 없으면 GORM 은 구조체 이름을 따라 "tokens" 테이블을 만든다.
func (Token) TableName() string {
	return "refresh_tokens"
}
