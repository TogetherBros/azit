package room

import (
	"time"

	"gorm.io/gorm"
)

const DefaultName = "아지트"

type Room struct {
	RoomID    int64  `gorm:"primaryKey"`
	AzitID    int64  `gorm:"not null;index"`
	Name      string `gorm:"not null"`
	Position  int    `gorm:"not null;default:0"` // 방 순서. 0부터 시작. 같은 아지트 내에서만 유효
	CreatedBy int64  `gorm:"not null;index"`     // 만든 사람 기록용 (accounts.id)
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"` // 지운 방. 메시지는 그대로 두고 방만 숨긴다 (변경 기록에서 되돌리면 다시 보인다)
}
