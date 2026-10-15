package azit

import (
	"context"
	"errors"

	"github.com/TogetherBros/azit/server/internal/member"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("아지트가 없어요")

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// FindByID 는 아지트 하나를 찾는다. 없으면 ErrNotFound.
func (r *Repository) FindByID(ctx context.Context, id int64) (Azit, error) {
	var a Azit
	err := r.db.WithContext(ctx).First(&a, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Azit{}, ErrNotFound
	}
	return a, err
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
		azit = Azit{
			MasterID:  input.MasterID,
			Name:      input.Name,
			Color:     input.Color,
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
