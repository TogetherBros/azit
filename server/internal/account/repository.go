package account

import (
	"context"
	"errors"

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
