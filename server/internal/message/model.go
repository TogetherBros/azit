package message

import "time"

// Kind 는 메시지 종류.
type Kind string

const (
	KindText   Kind = "text"
	KindPhoto  Kind = "photo"  // 직접 올린 사진. Body 에 R2 키 (photos/...)
	KindMeme   Kind = "meme"   // 짤. Body 에 R2 키 (memes/...)
	KindSystem Kind = "system" // "OO님이 들어왔어요" 같은 서버 메시지. SenderID 는 0
)

// Message 는 방에 올라온 메시지 한 개.
// 지워도 줄은 남기고 DeletedAt 만 채운다 ("삭제된 메시지" 자리를 남기려고).
type Message struct {
	MessageID int64  `gorm:"primaryKey"`
	AzitID    int64  `gorm:"not null;index"` // 실시간으로 뿌릴 때 아지트를 바로 알려고 같이 저장
	RoomID    int64  `gorm:"not null;index"` // 불러오기가 "이 방 + id < before" 라서 인덱스
	SenderID  int64  `gorm:"not null;uniqueIndex:idx_sender_temp,priority:1"`
	TempID    string `gorm:"not null;uniqueIndex:idx_sender_temp,priority:2"` // 앱이 만든 번호표. 같은 사람 + 같은 번호표면 재전송
	Kind      Kind   `gorm:"not null;default:'text'"`
	Body      string `gorm:"not null"`
	MemeID    int64  `gorm:"not null;default:0"` // 짤 메시지면 보낸 짤 id. Body 에는 그때의 R2 키를 같이 남긴다 (짤을 빼도 메시지는 보이게)
	ReplyToID *int64 // 답장이면 원래 메시지 id
	ShakeAll  bool   `gorm:"not null;default:false"` // @전체 로 불렀는지 (온라인인 전원을 흔든다). 사람별은 Shake 테이블
	CreatedAt time.Time
	EditedAt  *time.Time
	DeletedAt *time.Time
}

// Read 는 누가 어느 방을 어디까지 읽었는지. 계정 + 방 하나에 한 줄.
type Read struct {
	AccountID         int64 `gorm:"primaryKey;autoIncrement:false"`
	RoomID            int64 `gorm:"primaryKey;autoIncrement:false"`
	LastReadMessageID int64 `gorm:"not null;default:0"`
	UpdatedAt         time.Time
}

// TableName 은 테이블 이름을 room_reads 로 정한다. 안 정하면 GORM 이 reads 로 만든다.
func (Read) TableName() string {
	return "room_reads"
}

// Shake 는 메시지에서 @이름 으로 부른 사람. 부르면 무조건 그 사람 창이 흔들린다. 메시지 + 사람 하나에 한 줄.
// @전체 는 여기 줄을 만들지 않고 Message.ShakeAll 로 표시한다 (20명 줄을 매번 만들지 않으려고).
type Shake struct {
	MessageID int64 `gorm:"primaryKey;autoIncrement:false"`
	TargetID  int64 `gorm:"primaryKey;autoIncrement:false;index"` // 불린 사람 accounts.id. "나를 부른 메시지" 찾을 때 쓴다
}

// Reaction 은 메시지에 단 맞장구 하나 (글자 또는 짤).
// 한 사람이 여러 개 달 수 있고, 같은 걸 두 번은 못 단다. 화면에는 같은 맞장구끼리 묶어서 "ㅋㅋ 2" 로 센다.
// 아지트 맞장구 목록(meme.React)의 id 가 아니라 내용을 저장한다. 목록에 없는 짤도 달 수 있고, 목록을 바꿔도 이미 단 건 안 깨진다.
type Reaction struct {
	MessageID int64  `gorm:"primaryKey;autoIncrement:false"`
	AccountID int64  `gorm:"primaryKey;autoIncrement:false"`
	Text      string `gorm:"primaryKey;default:''"`                    // 글자 맞장구 ("ㅋㅋ"). 짤이면 ""
	MemeID    int64  `gorm:"primaryKey;autoIncrement:false;default:0"` // 짤 맞장구면 meme.Meme id. 글자면 0
	CreatedAt time.Time
}
