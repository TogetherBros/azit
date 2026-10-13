package refresh

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ErrInvalid 는 RT 가 없거나, 만료됐거나, 이미 쓰인 경우.
var ErrInvalid = errors.New("refresh token 이 없거나 만료됐어요")

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Save(ctx context.Context, accountID int64, hash string, expiresAt time.Time) error {
	t := Token{AccountID: accountID, TokenHash: hash, ExpiresAt: expiresAt}
	return r.db.WithContext(ctx).Create(&t).Error
}

// Consume 은 유효한 RT 를 찾아 지우고 주인 accountID 를 돌려준다.
// 지우면서 확인하므로 같은 RT 로 두 번 갱신할 수 없다 (회전).
func (r *Repository) Consume(ctx context.Context, hash string) (int64, error) {
	var t Token
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND expires_at > ?", hash, time.Now()).
		First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, ErrInvalid
	}
	if err != nil {
		return 0, err
	}

	res := r.db.WithContext(ctx).Delete(&Token{}, t.ID)
	if res.Error != nil {
		return 0, res.Error
	}
	// 찾은 뒤 지우기 전에 다른 요청이 먼저 지웠다 = 동시에 같은 RT 를 썼다
	if res.RowsAffected == 0 {
		return 0, ErrInvalid
	}

	return t.AccountID, nil
}

// Delete 는 로그아웃할 때 RT 하나를 지운다. 없어도 에러가 아니다.
func (r *Repository) Delete(ctx context.Context, hash string) error {
	return r.db.WithContext(ctx).Where("token_hash = ?", hash).Delete(&Token{}).Error
}
