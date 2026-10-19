package meme

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("짤이 없어요")

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// List 는 아지트 짤 모음. 지운 건 빼고 최신 순.
func (r *Repository) List(ctx context.Context, azitID int64) ([]Meme, error) {
	var memes []Meme
	err := r.db.WithContext(ctx).
		Where("azit_id = ? AND deleted_at IS NULL", azitID).
		Order("meme_id DESC").
		Find(&memes).Error
	return memes, err
}

// Create 는 짤을 저장한다. 성공하면 m.MemeID, m.CreatedAt 이 채워진다.
func (r *Repository) Create(ctx context.Context, m *Meme) error {
	return r.db.WithContext(ctx).Create(m).Error
}

// FindActive 는 이 아지트의 지우지 않은 짤 하나. 없거나, 다른 아지트 것이거나, 지웠으면 ErrNotFound.
func (r *Repository) FindActive(ctx context.Context, azitID, memeID int64) (Meme, error) {
	var m Meme
	err := r.db.WithContext(ctx).
		Where("meme_id = ? AND azit_id = ? AND deleted_at IS NULL", memeID, azitID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Meme{}, ErrNotFound
	}
	return m, err
}

// FindByIDs 는 짤 여러 개를 id → 짤 로 돌려준다. 지운 짤도 포함한다
// (이미 달린 맞장구나 보낸 짤은 짤을 지워도 계속 보여야 해서).
func (r *Repository) FindByIDs(ctx context.Context, memeIDs []int64) (map[int64]Meme, error) {
	out := map[int64]Meme{}
	if len(memeIDs) == 0 {
		return out, nil
	}
	var memes []Meme
	if err := r.db.WithContext(ctx).Where("meme_id IN ?", memeIDs).Find(&memes).Error; err != nil {
		return nil, err
	}
	for _, m := range memes {
		out[m.MemeID] = m
	}
	return out, nil
}

// SoftDelete 는 짤을 지운다 (줄은 남기고 DeletedAt 만). 아지트 맞장구 목록에 있던 칸도 같이 뺀다.
func (r *Repository) SoftDelete(ctx context.Context, memeID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Meme{}).Where("meme_id = ?", memeID).
			Update("deleted_at", time.Now()).Error; err != nil {
			return err
		}
		return tx.Where("meme_id = ?", memeID).Delete(&React{}).Error
	})
}

// Undelete 는 지운 짤을 되살린다 (변경 기록 되돌리기). 맞장구 목록에서 빠진 칸은 돌아오지 않는다.
func (r *Repository) Undelete(ctx context.Context, azitID, memeID int64) error {
	return r.db.WithContext(ctx).Model(&Meme{}).
		Where("azit_id = ? AND meme_id = ?", azitID, memeID).
		Update("deleted_at", nil).Error
}

// ListReacts 는 아지트 맞장구 목록. 순서대로.
func (r *Repository) ListReacts(ctx context.Context, azitID int64) ([]React, error) {
	var reacts []React
	err := r.db.WithContext(ctx).
		Where("azit_id = ?", azitID).
		Order("position, react_id").
		Find(&reacts).Error
	return reacts, err
}

// ReplaceReacts 는 아지트 맞장구 목록을 통째로 바꾼다. 받은 순서대로 Position 0, 1, 2 ...
// 지우고 새로 넣는 걸 한 트랜잭션으로 해서, 중간에 실패해도 목록이 비지 않는다.
func (r *Repository) ReplaceReacts(ctx context.Context, azitID int64, reacts []React) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("azit_id = ?", azitID).Delete(&React{}).Error; err != nil {
			return err
		}
		if len(reacts) == 0 {
			return nil
		}
		for i := range reacts {
			reacts[i].ReactID = 0 // 새 줄로 넣는다
			reacts[i].AzitID = azitID
			reacts[i].Position = i
		}
		return tx.Create(&reacts).Error
	})
}

// CountActiveAmong 은 memeIDs 중 이 아지트의 지우지 않은 짤이 몇 개인지 센다.
// "맞장구 목록에 넣은 짤이 전부 우리 아지트 짤인가" 를 쿼리 한 번으로 볼 때 쓴다.
// HasReactText 는 아지트 맞장구 목록에 이 글자 칸이 있는지 본다.
func (r *Repository) HasReactText(ctx context.Context, azitID int64, text string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&React{}).
		Where("azit_id = ? AND text = ?", azitID, text).
		Count(&n).Error
	return n > 0, err
}

func (r *Repository) CountActiveAmong(ctx context.Context, azitID int64, memeIDs []int64) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&Meme{}).
		Where("azit_id = ? AND meme_id IN ? AND deleted_at IS NULL", azitID, memeIDs).
		Count(&n).Error
	return n, err
}
