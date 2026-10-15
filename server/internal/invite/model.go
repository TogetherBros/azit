package invite

import "time"

// Invite 는 아지트 초대 코드 한 개. 만료되거나 취소돼도 지우지 않는다 (코드 재사용 막기).
type Invite struct {
	InviteID  int64      `gorm:"primaryKey"`
	AzitID    int64      `gorm:"not null;index"`
	Code      string     `gorm:"not null;uniqueIndex"` // 8자리, 하이픈 없이 저장
	CreatedBy int64      `gorm:"not null"`             // 만든 사람 (accounts.id)
	ExpiresAt *time.Time // nil 이면 무제한
	MaxUses   int        `gorm:"not null;default:0"` // 0 이면 무제한
	Uses      int        `gorm:"not null;default:0"`
	RevokedAt *time.Time // 취소했으면 그 시각
	CreatedAt time.Time
}

// Usable 은 지금 이 코드로 들어올 수 있는지 본다 (취소, 만료, 횟수 다 씀이 아니면 true).
func (i Invite) Usable(now time.Time) bool {
	if i.RevokedAt != nil {
		return false
	}
	if i.ExpiresAt != nil && !now.Before(*i.ExpiresAt) {
		return false
	}
	if i.MaxUses > 0 && i.Uses >= i.MaxUses {
		return false
	}
	return true
}
