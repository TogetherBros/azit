package message

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MaxReactionKinds 는 한 메시지에 달 수 있는 맞장구 종류 수. 같은 맞장구를 여러 명이 다는 건 1종류로 센다.
const MaxReactionKinds = 10

var ErrReactionLimit = errors.New("이 메시지에는 맞장구를 더 달 수 없어요 (10종류까지)")

// AddReaction 은 맞장구를 단다. 이미 달려 있으면 아무 일도 안 한다 (두 번 눌러도 같은 결과).
// 이 메시지에 없던 새 종류인데 이미 10종류면 ErrReactionLimit.
// 세기와 넣기를 한 트랜잭션에서 해서, 센 다음 넣기 전에 다른 요청이 끼어드는 걸 줄인다.
func (r *Repository) AddReaction(ctx context.Context, rc Reaction) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 이 종류(글자 또는 짤)가 이미 누군가에 의해 달려 있는지
		var same int64
		if err := tx.Model(&Reaction{}).
			Where("message_id = ? AND text = ? AND meme_id = ?", rc.MessageID, rc.Text, rc.MemeID).
			Count(&same).Error; err != nil {
			return err
		}

		// 새 종류면 지금 몇 종류인지 센다
		if same == 0 {
			var kinds int64
			sub := tx.Model(&Reaction{}).Select("DISTINCT text, meme_id").Where("message_id = ?", rc.MessageID)
			if err := tx.Table("(?) AS k", sub).Count(&kinds).Error; err != nil {
				return err
			}
			if kinds >= MaxReactionKinds {
				return ErrReactionLimit
			}
		}

		// 내가 이미 단 거면 PK 가 겹친다. 에러 대신 그냥 넘어간다 (ON CONFLICT DO NOTHING).
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rc).Error
	})
}

// RemoveReaction 은 내가 단 맞장구를 뗀다. 안 달려 있었어도 에러가 아니다.
func (r *Repository) RemoveReaction(ctx context.Context, messageID, accountID int64, text string, memeID int64) error {
	return r.db.WithContext(ctx).
		Where("message_id = ? AND account_id = ? AND text = ? AND meme_id = ?", messageID, accountID, text, memeID).
		Delete(&Reaction{}).Error
}

// ReactionsByMessages 는 메시지 여러 개에 달린 맞장구 줄들을 메시지 id → 줄 목록 으로 돌려준다. 단 순서대로.
func (r *Repository) ReactionsByMessages(ctx context.Context, messageIDs []int64) (map[int64][]Reaction, error) {
	out := map[int64][]Reaction{}
	if len(messageIDs) == 0 {
		return out, nil
	}
	var rows []Reaction
	err := r.db.WithContext(ctx).
		Where("message_id IN ?", messageIDs).
		Order("created_at, account_id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, rc := range rows {
		out[rc.MessageID] = append(out[rc.MessageID], rc)
	}
	return out, nil
}
