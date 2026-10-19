package member

import (
	"context"
	"errors"
	"time"

	"github.com/TogetherBros/azit/server/internal/notification"
	"gorm.io/gorm"
)

// ErrNotMember 는 그 아지트에 들어와 있지 않을 때 (나간 사람 포함).
var ErrNotMember = errors.New("이 아지트 멤버가 아니에요")

// ErrNotBanned 는 막힌 사람이 아닐 때.
var ErrNotBanned = errors.New("막은 사람이 아니에요")

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

// ListActive 는 지금 들어와 있는 멤버 전부. 들어온 순서대로.
func (r *Repository) ListActive(ctx context.Context, azitID int64) ([]Member, error) {
	var members []Member
	err := r.db.WithContext(ctx).
		Where("azit_id = ? AND left_at IS NULL", azitID).
		Order("joined_at, account_id").
		Find(&members).Error
	return members, err
}

// ActiveIDs 는 아지트에 지금 들어와 있는 멤버들의 계정 id.
func (r *Repository) ActiveIDs(ctx context.Context, azitID int64) ([]int64, error) {
	var ids []int64
	err := r.db.WithContext(ctx).Model(&Member{}).
		Where("azit_id = ? AND left_at IS NULL", azitID).
		Pluck("account_id", &ids).Error
	return ids, err
}

// CoMemberIDs 는 나와 같은 아지트에 지금 들어와 있는 사람들 (나 빼고, 겹치면 한 번만).
// "누가 접속했다" 를 알릴 사람들이다.
func (r *Repository) CoMemberIDs(ctx context.Context, accountID int64) ([]int64, error) {
	var ids []int64
	err := r.db.WithContext(ctx).Model(&Member{}).
		Distinct("account_id").
		Where("left_at IS NULL AND account_id <> ?", accountID).
		Where("azit_id IN (?)", r.db.Model(&Member{}).Select("azit_id").Where("account_id = ? AND left_at IS NULL", accountID)).
		Pluck("account_id", &ids).Error
	return ids, err
}

// UpdateNickName 은 이 아지트 안 닉네임을 바꾼다.
func (r *Repository) UpdateNickName(ctx context.Context, azitID, accountID int64, nickname string) error {
	return r.db.WithContext(ctx).Model(&Member{}).
		Where("azit_id = ? AND account_id = ? AND left_at IS NULL", azitID, accountID).
		Update("nick_name", nickname).Error
}

// UpdateRole 은 역할을 바꾼다.
func (r *Repository) UpdateRole(ctx context.Context, azitID, accountID int64, role Role) error {
	return r.db.WithContext(ctx).Model(&Member{}).
		Where("azit_id = ? AND account_id = ? AND left_at IS NULL", azitID, accountID).
		Update("role", role).Error
}

// Kick 은 멤버를 내보내고, 내보내진 사람에게 알림을 남긴다. 한 트랜잭션.
// kicker 는 내보낸 사람 (알림에 "누가" 로 남는다). ban 이면 막기도 같이 한다 (새 코드로도 못 들어오게).
// 저장된 알림을 돌려준다 (커밋 뒤 실시간으로 보내려고).
func (r *Repository) Kick(ctx context.Context, kicker Member, target Member, at time.Time, ban bool) ([]notification.Notification, error) {
	notes := []notification.Notification{{
		AccountID: target.AccountID,
		Kind:      notification.KindKick,
		AzitID:    kicker.AzitID,
		ActorID:   kicker.AccountID,
		ActorNick: kicker.NickName,
	}}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&Member{}).
			Where("azit_id = ? AND account_id = ? AND left_at IS NULL", kicker.AzitID, target.AccountID).
			Update("left_at", at).Error
		if err != nil {
			return err
		}
		if ban {
			err := tx.Save(&Ban{AzitID: kicker.AzitID, AccountID: target.AccountID, Nickname: target.NickName, BannedByNick: kicker.NickName}).Error
			if err != nil {
				return err
			}
		}
		return notification.Insert(tx, notes)
	})
	if err != nil {
		return nil, err
	}
	return notes, nil
}

// Restore 는 내보낸 사람을 다시 들여보낸다 (변경 기록 되돌리기). 막은 것도 푼다.
// 그 사이 스스로 다시 들어와 있으면 아무것도 안 한다.
func (r *Repository) Restore(ctx context.Context, azitID, accountID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("azit_id = ? AND account_id = ?", azitID, accountID).Delete(&Ban{}).Error; err != nil {
			return err
		}
		return tx.Model(&Member{}).
			Where("azit_id = ? AND account_id = ? AND left_at IS NOT NULL", azitID, accountID).
			Updates(map[string]any{"left_at": nil, "role": RoleMember}).Error
	})
}

// IsBanned 는 이 아지트에서 막힌 사람인지.
func IsBanned(tx *gorm.DB, azitID, accountID int64) (bool, error) {
	var n int64
	err := tx.Model(&Ban{}).Where("azit_id = ? AND account_id = ?", azitID, accountID).Count(&n).Error
	return n > 0, err
}

// ListBans 는 막은 사람들 (최근 순).
func (r *Repository) ListBans(ctx context.Context, azitID int64) ([]Ban, error) {
	var out []Ban
	err := r.db.WithContext(ctx).Where("azit_id = ?", azitID).Order("created_at DESC").Find(&out).Error
	return out, err
}

// FindBan 은 막은 줄 하나. 없으면 ErrNotMember 와 같은 404 로 다룬다.
func (r *Repository) FindBan(ctx context.Context, azitID, accountID int64) (Ban, error) {
	var b Ban
	err := r.db.WithContext(ctx).Where("azit_id = ? AND account_id = ?", azitID, accountID).First(&b).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Ban{}, ErrNotBanned
	}
	return b, err
}

// Unban 은 막은 것을 푼다. 다시 들어오려면 새 초대 코드가 필요하다.
func (r *Repository) Unban(ctx context.Context, azitID, accountID int64) error {
	return r.db.WithContext(ctx).Where("azit_id = ? AND account_id = ?", azitID, accountID).Delete(&Ban{}).Error
}

// Reban 은 풀었던 것을 다시 막는다 (변경 기록 되돌리기).
func (r *Repository) Reban(ctx context.Context, b Ban) error {
	return r.db.WithContext(ctx).Save(&b).Error
}

// Leave 는 멤버를 나간 상태로 만든다 (스스로 나가기, 내보내기 둘 다). 줄은 지우지 않는다:
// 예전 메시지에 닉네임이 보여야 하고, 초대 코드로 다시 들어오면 이 줄을 살린다.
func (r *Repository) Leave(ctx context.Context, azitID, accountID int64, at time.Time) error {
	return r.db.WithContext(ctx).Model(&Member{}).
		Where("azit_id = ? AND account_id = ? AND left_at IS NULL", azitID, accountID).
		Update("left_at", at).Error
}
