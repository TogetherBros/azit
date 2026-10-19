package invite

import (
	"context"
	"errors"
	"time"

	"github.com/TogetherBros/azit/server/internal/member"
	"github.com/TogetherBros/azit/server/internal/notification"
	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("없는 초대 코드예요")
	ErrUnusable      = errors.New("만료됐거나 더 쓸 수 없는 초대 코드예요")
	ErrAlreadyMember = errors.New("이미 들어와 있는 아지트예요")
	ErrBanned        = errors.New("이 아지트에는 들어갈 수 없어요")
)

// createTries 는 코드가 겹쳤을 때 다시 만들어 보는 횟수. 32^8 이라 실제로 겹칠 일은 거의 없다.
const createTries = 5

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 는 새 코드를 붙여 초대를 저장한다. 코드가 겹치면 새로 만들어 다시 시도한다.
func (r *Repository) Create(ctx context.Context, inv *Invite) error {
	for range createTries {
		code, err := NewCode()
		if err != nil {
			return err
		}
		inv.Code = code

		err = r.db.WithContext(ctx).Create(inv).Error
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			inv.InviteID = 0
			continue
		}
		return err
	}
	return errors.New("초대 코드를 만들지 못했어요")
}

// FindByCode 는 코드로 초대를 찾는다. code 는 NormalizeCode 를 거친 값이어야 한다.
func (r *Repository) FindByCode(ctx context.Context, code string) (Invite, error) {
	var inv Invite
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&inv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Invite{}, ErrNotFound
	}
	return inv, err
}

// ListUsable 은 그 아지트에서 지금 쓸 수 있는 초대만 최신순으로 돌려준다.
func (r *Repository) ListUsable(ctx context.Context, azitID int64, now time.Time) ([]Invite, error) {
	var all []Invite
	err := r.db.WithContext(ctx).
		Where("azit_id = ? AND revoked_at IS NULL", azitID).
		Order("created_at DESC").
		Find(&all).Error
	if err != nil {
		return nil, err
	}

	usable := make([]Invite, 0, len(all))
	for _, inv := range all {
		if inv.Usable(now) {
			usable = append(usable, inv)
		}
	}
	return usable, nil
}

// Revoke 는 초대를 취소한다. 줄은 지우지 않는다 (코드 재사용 막기).
func (r *Repository) Revoke(ctx context.Context, inviteID int64, now time.Time) error {
	return r.db.WithContext(ctx).
		Model(&Invite{}).
		Where("invite_id = ? AND revoked_at IS NULL", inviteID).
		Update("revoked_at", now).Error
}

// notifyJoin 은 원래 있던 멤버들(들어온 사람 빼고)에게 "누가 들어왔어요" 알림을 남긴다. Join 트랜잭션 안에서 부른다.
// 저장된 알림들을 돌려준다 (커밋 뒤 실시간으로 보내려고).
func notifyJoin(tx *gorm.DB, azitID, joinerID int64, nickname string) ([]notification.Notification, error) {
	var ids []int64
	err := tx.Model(&member.Member{}).
		Where("azit_id = ? AND left_at IS NULL AND account_id <> ?", azitID, joinerID).
		Pluck("account_id", &ids).Error
	if err != nil {
		return nil, err
	}

	notes := make([]notification.Notification, 0, len(ids))
	for _, id := range ids {
		notes = append(notes, notification.Notification{
			AccountID: id,
			Kind:      notification.KindJoin,
			AzitID:    azitID,
			ActorID:   joinerID,
			ActorNick: nickname,
		})
	}
	if err := notification.Insert(tx, notes); err != nil {
		return nil, err
	}
	return notes, nil
}

// JoinInput 은 초대 코드로 들어올 때 필요한 값.
type JoinInput struct {
	Code      string // NormalizeCode 를 거친 값
	AccountID int64
	Nickname  string // 계정 닉네임. 아지트 안 닉네임으로 복사해 둔다
	Now       time.Time
}

// Join 은 초대 코드로 아지트에 들어간다. 한 트랜잭션 안에서
// 1) 코드 확인 2) 이미 멤버인지 확인 3) 사용 횟수 +1 (조건부) 4) 멤버 줄 만들기/살리기 를 한다.
// 돌려주는 값은 들어간 아지트의 ID 와, 원래 멤버들에게 남긴 입장 알림들 (커밋 뒤 실시간으로 보내려고).
func (r *Repository) Join(ctx context.Context, in JoinInput) (int64, []notification.Notification, error) {
	var azitID int64
	var notes []notification.Notification

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var inv Invite
		err := tx.Where("code = ?", in.Code).First(&inv).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		azitID = inv.AzitID

		// 이미 들어와 있으면 코드가 다 쓰였어도 "이미 멤버" 로 알려준다.
		var m member.Member
		err = tx.Where("azit_id = ? AND account_id = ?", inv.AzitID, in.AccountID).First(&m).Error
		found := err == nil
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if found && m.LeftAt == nil {
			return ErrAlreadyMember
		}

		if !inv.Usable(in.Now) {
			return ErrUnusable
		}

		// 막힌 사람은 새 코드로도 못 들어온다.
		banned, err := member.IsBanned(tx, inv.AzitID, in.AccountID)
		if err != nil {
			return err
		}
		if banned {
			return ErrBanned
		}

		// 여러 명이 동시에 들어와도 MaxUses 를 넘지 않게, 조건을 건 UPDATE 한 번으로 올린다.
		res := tx.Model(&Invite{}).
			Where("invite_id = ? AND revoked_at IS NULL AND (max_uses = 0 OR uses < max_uses)", inv.InviteID).
			UpdateColumn("uses", gorm.Expr("uses + 1"))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrUnusable
		}

		if found {
			// 나갔다 다시 들어오는 사람: 새 줄 대신 기존 줄을 살린다. 역할은 일반으로 되돌린다.
			err = tx.Model(&member.Member{}).
				Where("azit_id = ? AND account_id = ?", inv.AzitID, in.AccountID).
				Updates(map[string]any{
					"left_at":     nil,
					"joined_at":   in.Now,
					"role":        member.RoleMember,
					"nick_name":   in.Nickname,
					"invite_code": inv.Code,
				}).Error
		} else {
			err = tx.Create(&member.Member{
				AzitID:     inv.AzitID,
				AccountID:  in.AccountID,
				NickName:   in.Nickname,
				Role:       member.RoleMember,
				InviteCode: inv.Code,
			}).Error
		}
		if err != nil {
			return err
		}

		notes, err = notifyJoin(tx, inv.AzitID, in.AccountID, in.Nickname)
		return err
	})

	return azitID, notes, err
}
