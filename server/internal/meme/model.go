package meme

import "time"

// Meme 은 아지트에 저장해 둔 짤 하나 (우리끼리 짤 모음).
// 메시지로 보내거나 맞장구 목록에 넣을 수 있다. 지워도 줄은 남긴다 (변경 기록에서 되돌리려고).
type Meme struct {
	MemeID    int64  `gorm:"primaryKey"`
	AzitID    int64  `gorm:"not null;index"`
	Name      string `gorm:"not null"` // "킹받네" 같은 이름. 짤 찾기에 쓴다
	MediaKey  string `gorm:"not null"` // R2 키 (memes/...)
	CreatedBy int64  `gorm:"not null"` // 올린 사람 accounts.id
	CreatedAt time.Time
	DeletedAt *time.Time
}

// React 는 아지트 맞장구 목록(빠른 선택용)의 한 칸. 아지트마다 최대 10개, 순서가 있다.
// 글자(2자까지) 또는 아지트 짤 하나 중 하나만 채운다.
type React struct {
	ReactID  int64  `gorm:"primaryKey"`
	AzitID   int64  `gorm:"not null;index"`
	Text     string `gorm:"not null;default:''"` // "ㅋㅋ", "헐". 짤이면 ""
	MemeID   *int64 // 짤이면 Meme id. 글자면 nil
	Position int    `gorm:"not null;default:0"` // 목록 순서. 작을수록 앞
}
