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

// Ban 은 막은 사람. 이 줄이 있으면 새 초대 코드로도 이 아지트에 다시 못 들어온다.
// 풀면 줄을 지운다 (풀어도 자동으로 다시 들어오진 않는다. 새 코드가 필요하다).
type Ban struct {
	AzitID       int64  `gorm:"primaryKey"`
	AccountID    int64  `gorm:"primaryKey"`
	Nickname     string `gorm:"not null;default:''"` // 막을 때의 아지트 닉네임
	BannedByNick string `gorm:"not null;default:''"`
	CreatedAt    time.Time
}
