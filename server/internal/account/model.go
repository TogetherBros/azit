package account

import "time"

// Provider 는 어느 로그인 서비스로 가입했는지. DB 에는 글자 그대로 저장된다.
type Provider string

const (
	ProviderGoogle Provider = "google"
)

type Account struct {
	ID        int64    `gorm:"primaryKey"`
	Provider  Provider `gorm:"not null;uniqueIndex:idx_provider_subject"`
	Subject   string   `gorm:"not null;uniqueIndex:idx_provider_subject"` // 구글이면 id_token 의 sub
	Email     string   `gorm:"not null"`
	Name      *string  `gorm:"uniqueIndex"`                // 이름 (중복 불가, @ 와 함께 보임). 프로필 만들기 전엔 nil
	Nickname  string   `gorm:"not null;default:''"`        // 닉네임 (겹쳐도 됨). 처음엔 구글 이름
	Bio       string   `gorm:"not null;default:''"`        // 자기소개 (190자)
	Color     string   `gorm:"not null;default:'#D4FF3A'"` // 내 색
	AvatarKey string   `gorm:"not null;default:''"`        // R2 사진 키
	BannerKey string   `gorm:"not null;default:''"`        // R2 배너 키
	Disabled  bool     `gorm:"not null;default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
