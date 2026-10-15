package message

type Kind string

const (
	KindText   Kind = "text"
	KindImage  Kind = "image"
	KindSystem Kind = "system"
	KindMeme   Kind = "meme"
)

type Message struct {
	MessageID int64  `gorm:"primaryKey"`
	RoomID    int64  `gorm:"not null;index"`
	SenderID  int64  `gorm:"not null;index"`
	Content   string `gorm:"not null"`
	CreatedAt int64  `gorm:"autoCreateTime:nano"`
}
