package member

import "time"

// Member 는 아지트와 계정을 잇는다 (다대다 연결 테이블).
// 나갔다 다시 들어오면 새 줄이 아니라 이 줄을 고친다 (LeftAt = nil, JoinedAt = 지금).
type Member struct {
	AzitID     int64      `gorm:"primaryKey"`
	AccountID  int64      `gorm:"primaryKey;index"`          // "내 아지트 목록"은 account_id 로 찾는다
	NickName   string     `gorm:"not null;default:''"`       // 아지트 안 닉네임. 들어올 때 계정 닉네임을 복사해 두고, 나중에 고칠 수 있다
	Role       Role       `gorm:"not null;default:'member'"` // 역할
	JoinedAt   time.Time  `gorm:"autoCreateTime"`
	LeftAt     *time.Time // 나갔으면 그 시각, 들어와 있으면 nil
	InviteCode string     `gorm:"not null;default:''"` // 들어올 때 쓴 코드. 만든 사람은 ""
}
