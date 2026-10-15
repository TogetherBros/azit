package azit

import "time"

type Azit struct {
	AzitID    int64  `gorm:"primaryKey"`
	MasterID  int64  `gorm:"not null;index"`             // 만든 사람 기록용 (accounts.id). 권한은 member.Role 로 본다
	Name      string `gorm:"not null"`                   // 겹쳐도 됨
	Color     string `gorm:"not null;default:'#D4FF3A'"` // 아지트 색
	Code      string `gorm:"not null;uniqueIndex"`       // 초대 코드 6자리
	BannerKey string `gorm:"not null;default:''"`        // R2 배너 키. 없으면 ""
	CreatedAt time.Time
}
