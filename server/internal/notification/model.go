package notification

import "time"

// Kind 는 알림 종류.
type Kind string

const (
	KindShake Kind = "shake" // 누가 나를 @로 불렀다 (@전체 포함)
	KindJoin  Kind = "join"  // 내 아지트에 누가 초대 코드로 들어왔다
	KindKick  Kind = "kick"  // 내가 아지트에서 내보내졌다
)

// Notification 은 한 사람에게 가는 알림 한 줄.
// 아지트·방 이름과 보낸 사람 닉네임은 만들 때 그대로 적어 둔다 (나중에 이름이 바뀌거나 내보내져도 알림 문구가 안 깨지게).
type Notification struct {
	NotificationID int64  `gorm:"primaryKey"`
	AccountID      int64  `gorm:"not null;index:idx_notif_account,priority:1"` // 받는 사람
	Kind           Kind   `gorm:"not null"`
	AzitID         int64  `gorm:"not null"`
	AzitName       string `gorm:"not null;default:''"`
	RoomID         int64  `gorm:"not null;default:0"` // shake 면 그 방, 아니면 0
	RoomName       string `gorm:"not null;default:''"`
	MessageID      int64  `gorm:"not null;default:0"`  // shake 면 그 메시지, 아니면 0
	ActorID        int64  `gorm:"not null;default:0"`  // 흔든 사람 / 들어온 사람 / 내보낸 사람
	ActorNick      string `gorm:"not null;default:''"` // 그때의 아지트 닉네임
	Preview        string `gorm:"not null;default:''"` // shake 면 메시지 앞부분
	IsRead         bool   `gorm:"not null;default:false;index:idx_notif_account,priority:2"`
	CreatedAt      time.Time
}
