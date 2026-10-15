package azit

import (
	"context"
	"crypto/rand"
	"errors"

	"github.com/TogetherBros/azit/server/internal/member"
	"gorm.io/gorm"
)

var ErrCodeGeneration = errors.New("초대 코드를 만들지 못했어요")

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindMyAzits(ctx context.Context, accountID int64) ([]Azit, error) {
	var azits []Azit

	err := r.db.WithContext(ctx).
		Model(&Azit{}).
		Joins("JOIN members ON members.azit_id = azits.azit_id").
		Where("members.account_id = ?", accountID).
		Where("members.left_at IS NULL").
		Order("azits.created_at DESC").
		Find(&azits).Error

	return azits, err
}

type CreateInput struct {
	MasterID  int64
	Nickname  string // 만든 사람의 계정 닉네임. 아지트 안 닉네임으로 복사해 둔다
	Name      string
	Color     string
	BannerKey string
}

func (r *Repository) Create(ctx context.Context, input CreateInput) (Azit, error) {
	var azit Azit

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		code, err := newInviteCode()
		if err != nil {
			return err
		}

		azit = Azit{
			MasterID:  input.MasterID,
			Name:      input.Name,
			Color:     input.Color,
			Code:      code,
			BannerKey: input.BannerKey,
		}
		if err := tx.Create(&azit).Error; err != nil {
			return err
		}

		return tx.Create(&member.Member{
			AzitID:    azit.AzitID,
			AccountID: input.MasterID,
			NickName:  input.Nickname,
			Role:      member.RoleSuperAdmin,
		}).Error
	})

	return azit, err
}

const inviteCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func newInviteCode() (string, error) {
	code := make([]byte, 6)
	random := make([]byte, len(code))
	if _, err := rand.Read(random); err != nil {
		return "", ErrCodeGeneration
	}

	for i := range code {
		code[i] = inviteCodeAlphabet[int(random[i])%len(inviteCodeAlphabet)]
	}
	return string(code), nil
}
