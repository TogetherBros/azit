package account

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("계정이 없어요")

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByProviderSubject(ctx context.Context, provider Provider, subject string) (Account, error) {
	var a Account
	err := r.db.WithContext(ctx).
		Where("provider = ? AND subject = ?", provider, subject).
		First(&a).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Account{}, ErrNotFound
	}
	if err != nil {
		return Account{}, err
	}
	return a, nil
}

func (r *Repository) Create(ctx context.Context, a *Account) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *Repository) FindByID(ctx context.Context, id int64) (Account, error) {
	var a Account
	err := r.db.WithContext(ctx).First(&a, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Account{}, ErrNotFound
	}
	if err != nil {
		return Account{}, err
	}
	return a, nil
}

// IsNameTaken 은 그 이름(@ 뒤)을 누가 이미 쓰고 있는지 본다.
func (r *Repository) IsNameTaken(ctx context.Context, name string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&Account{}).Where("name = ?", name).Count(&n).Error
	return n > 0, err
}

// SuggestName 은 이메일로 아직 아무도 안 쓰는 이름을 하나 추천한다.
// 추천일 뿐이라, 저장할 때 DB 유니크 인덱스가 한 번 더 막는다.
func (r *Repository) SuggestName(ctx context.Context, email string) (string, error) {
	base := NameBase(email)
	candidate := base
	for i := 0; i < 10; i++ {
		taken, err := r.IsNameTaken(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !taken {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s%d", base, rand.IntN(9000)+1000)
	}
	return "", errors.New("이름을 추천하지 못했어요")
}

var ErrNameTaken = errors.New("이미 누가 쓰고 있는 이름이에요")

// ProfileUpdate 는 사용자가 바꿀 수 있는 프로필 칸만 담는다.
type ProfileUpdate struct {
	Name     string
	Nickname string
	Bio      string
	Color    string
}

// UpdateProfile 은 프로필을 저장한다. 이름이 겹치면 ErrNameTaken.
// 빈 값("")도 저장해야 해서 구조체가 아니라 map 으로 넘긴다.
func (r *Repository) UpdateProfile(ctx context.Context, id int64, p ProfileUpdate) error {
	err := r.db.WithContext(ctx).Model(&Account{ID: id}).Updates(map[string]any{
		"name":     p.Name,
		"nickname": p.Nickname,
		"bio":      p.Bio,
		"color":    p.Color,
	}).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrNameTaken
	}
	return err
}

func (r *Repository) SaveAvatarKey(ctx context.Context, id int64, key string) error {
	return r.db.WithContext(ctx).Model(&Account{ID: id}).Update("avatar_key", key).Error
}

func (r *Repository) SaveBannerKey(ctx context.Context, id int64, key string) error {
	return r.db.WithContext(ctx).Model(&Account{ID: id}).Update("banner_key", key).Error
}

// FindByIDs 는 계정 여러 개를 id → 계정 으로 돌려준다. 없는 id 는 map 에 안 들어간다.
func (r *Repository) FindByIDs(ctx context.Context, ids []int64) (map[int64]Account, error) {
	out := map[int64]Account{}
	if len(ids) == 0 {
		return out, nil
	}
	var accounts []Account
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&accounts).Error; err != nil {
		return nil, err
	}
	for _, a := range accounts {
		out[a.ID] = a
	}
	return out, nil
}
