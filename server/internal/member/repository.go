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

// CountActiveAmong 은 accountIDs 중 지금 이 아지트에 들어와 있는 사람이 몇 명인지 센다.
// "부른 사람이 전부 멤버인가" 를 한 번의 쿼리로 확인할 때 쓴다 (센 수 == len(accountIDs) 면 전부 멤버).
func (r *Repository) CountActiveAmong(ctx context.Context, azitID int64, accountIDs []int64) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&Member{}).
		Where("azit_id = ? AND account_id IN ? AND left_at IS NULL", azitID, accountIDs).
		Count(&n).Error
	return n, err
}

// FindByAccountIDs 는 이 아지트의 멤버 여러 명을 account id → 멤버 로 돌려준다.
// 나간 사람도 포함한다 (나간 사람이 예전에 보낸 메시지에도 닉네임은 보여야 해서).
func (r *Repository) FindByAccountIDs(ctx context.Context, azitID int64, accountIDs []int64) (map[int64]Member, error) {
	out := map[int64]Member{}
	if len(accountIDs) == 0 {
		return out, nil
	}
	var members []Member
	err := r.db.WithContext(ctx).
		Where("azit_id = ? AND account_id IN ?", azitID, accountIDs).
		Find(&members).Error
	if err != nil {
		return nil, err
	}
	for _, m := range members {
		out[m.AccountID] = m
	}
	return out, nil
}
