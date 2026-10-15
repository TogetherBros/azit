package member

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// ErrNotMember 는 그 아지트에 들어와 있지 않을 때 (나간 사람 포함).
var ErrNotMember = errors.New("이 아지트 멤버가 아니에요")

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CountActive 는 아지트마다 지금 들어와 있는(나가지 않은) 인원 수를 센다.
// 결과는 azitID -> 인원. 맵에 없는 아지트는 0명이다.
func (r *Repository) CountActive(ctx context.Context, azitIDs []int64) (map[int64]int64, error) {
	counts := make(map[int64]int64, len(azitIDs))
	if len(azitIDs) == 0 {
		return counts, nil
	}

	var rows []struct {
		AzitID int64
		Count  int64
	}
	err := r.db.WithContext(ctx).
		Model(&Member{}).
		Select("azit_id, COUNT(*) AS count").
		Where("azit_id IN ?", azitIDs).
		Where("left_at IS NULL").
		Group("azit_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		counts[row.AzitID] = row.Count
	}
	return counts, nil
}

// FindActive 는 지금 들어와 있는 멤버 한 명을 찾는다. 없거나 나갔으면 ErrNotMember.
func (r *Repository) FindActive(ctx context.Context, azitID, accountID int64) (Member, error) {
	var m Member
	err := r.db.WithContext(ctx).
		Where("azit_id = ? AND account_id = ? AND left_at IS NULL", azitID, accountID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Member{}, ErrNotMember
	}
	return m, err
}
