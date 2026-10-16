package message

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ErrNotFound 는 찾는 메시지가 없을 때.
var ErrNotFound = errors.New("message: not found")

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 는 메시지와 @로 부른 사람(Shake)들을 한 트랜잭션으로 저장한다.
// 성공하면 m.MessageID, m.CreatedAt 이 채워진다. shakeIDs 중복 제거는 부르는 쪽에서 한다.
func (r *Repository) Create(ctx context.Context, m *Message, shakeIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(m).Error; err != nil {
			return err
		}

		if len(shakeIDs) == 0 {
			return nil // 부른 사람이 없으면 여기서 끝
		}

		// 부른 사람이 있으면 shakes 테이블에 한 번에 넣는다.
		shakes := make([]Shake, 0, len(shakeIDs))
		for _, id := range shakeIDs {
			shakes = append(shakes, Shake{MessageID: m.MessageID, TargetID: id})
		}

		return tx.Create(&shakes).Error
	})
}

// FindByTempID 는 보낸 사람 + 번호표로 이미 저장된 메시지를 찾는다. 재전송이 왔을 때 쓴다.
// 없으면 ErrNotFound.
func (r *Repository) FindByTempID(ctx context.Context, senderID int64, tempID string) (Message, error) {
	var m Message
	err := r.db.WithContext(ctx).
		Where("sender_id = ? AND temp_id = ?", senderID, tempID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Message{}, ErrNotFound
	}
	return m, err
}

// FindByID 는 메시지 하나를 찾는다. 없으면 ErrNotFound.
func (r *Repository) FindByID(ctx context.Context, messageID int64) (Message, error) {
	var m Message
	err := r.db.WithContext(ctx).First(&m, messageID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Message{}, ErrNotFound
	}
	return m, err
}

// ShakeIDsByMessages 는 메시지 여러 개의 @로 부른 사람들을 메시지 id → account id 목록 으로 돌려준다.
// 목록 화면에서 메시지마다 쿼리하지 않고 한 번에 가져오려고.
func (r *Repository) ShakeIDsByMessages(ctx context.Context, messageIDs []int64) (map[int64][]int64, error) {
	out := map[int64][]int64{}
	if len(messageIDs) == 0 {
		return out, nil
	}
	var shakes []Shake
	err := r.db.WithContext(ctx).
		Where("message_id IN ?", messageIDs).
		Order("message_id, target_id").
		Find(&shakes).Error
	if err != nil {
		return nil, err
	}
	for _, s := range shakes {
		out[s.MessageID] = append(out[s.MessageID], s.TargetID)
	}
	return out, nil
}

// ListBefore 는 방 메시지를 최신 순으로 limit 개 가져온다. before 보다 id 가 작은 것만 (before 가 0 이면 맨 최신부터).
// 위로 스크롤할 때 "지금 맨 위 메시지 id" 를 before 로 넘기면 그 위 페이지가 나온다.
// 지운 메시지도 가져온다 ("삭제된 메시지" 자리를 그리려고).
func (r *Repository) ListBefore(ctx context.Context, roomID, before int64, limit int) ([]Message, error) {
	q := r.db.WithContext(ctx).Where("room_id = ?", roomID)
	if before > 0 {
		q = q.Where("message_id < ?", before)
	}
	var messages []Message
	err := q.Order("message_id DESC").Limit(limit).Find(&messages).Error
	return messages, err
}

// UpdateBody 는 본문을 고치고 고친 시각을 남긴다.
func (r *Repository) UpdateBody(ctx context.Context, messageID int64, body string, at time.Time) error {
	return r.db.WithContext(ctx).Model(&Message{}).
		Where("message_id = ?", messageID).
		Updates(map[string]any{"body": body, "edited_at": at}).Error
}

// SoftDelete 는 메시지를 지운다. 줄은 남기고 DeletedAt 만 채운다.
func (r *Repository) SoftDelete(ctx context.Context, messageID int64, at time.Time) error {
	return r.db.WithContext(ctx).Model(&Message{}).
		Where("message_id = ?", messageID).
		Update("deleted_at", at).Error
}
