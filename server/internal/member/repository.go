package member

import (
	"context"

	"gorm.io/gorm"
)

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
