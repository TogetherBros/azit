package message

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MarkRead 는 "이 방을 여기까지 읽었다" 를 저장한다. 이미 더 뒤까지 읽었으면 그대로 둔다 (뒤로 가지 않는다).
// 줄이 없으면 새로 만들고, 있으면 고친다 (upsert). SQL 로는:
//
//	INSERT INTO room_reads ... ON CONFLICT (account_id, room_id)
//	DO UPDATE SET last_read_message_id = MAX(room_reads.last_read_message_id, excluded.last_read_message_id)
func (r *Repository) MarkRead(ctx context.Context, accountID, roomID, messageID int64) error {
	read := Read{AccountID: accountID, RoomID: roomID, LastReadMessageID: messageID, UpdatedAt: time.Now()}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "account_id"}, {Name: "room_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"last_read_message_id": gorm.Expr("MAX(room_reads.last_read_message_id, excluded.last_read_message_id)"),
			"updated_at":           gorm.Expr("excluded.updated_at"),
		}),
	}).Create(&read).Error
}

// LastRead 는 이 방을 어디까지 읽었는지. 한 번도 안 읽었으면 0.
func (r *Repository) LastRead(ctx context.Context, accountID, roomID int64) (int64, error) {
	var ids []int64
	err := r.db.WithContext(ctx).Model(&Read{}).
		Where("account_id = ? AND room_id = ?", accountID, roomID).
		Pluck("last_read_message_id", &ids).Error
	if err != nil || len(ids) == 0 {
		return 0, err
	}
	return ids[0], nil
}

// CountUnread 는 아지트의 방마다 내가 안 읽은 메시지 수를 방 id → 수 로 돌려준다.
// 안 읽은 메시지 = 내가 읽은 위치보다 뒤 + 내가 보낸 게 아님 + 지우지 않음. 0 인 방은 map 에 없다.
// room 패키지의 방 목록이 이걸 쓴다 (room.UnreadCounter).
func (r *Repository) CountUnread(ctx context.Context, azitID, accountID int64) (map[int64]int64, error) {
	var rows []struct {
		RoomID int64
		Count  int64
	}
	err := r.db.WithContext(ctx).Table("messages AS m").
		Select("m.room_id AS room_id, COUNT(*) AS count").
		Joins("LEFT JOIN room_reads AS r ON r.room_id = m.room_id AND r.account_id = ?", accountID).
		Where("m.azit_id = ? AND m.sender_id <> ? AND m.deleted_at IS NULL", azitID, accountID).
		Where("m.message_id > COALESCE(r.last_read_message_id, 0)").
		Where("m.room_id IN (SELECT room_id FROM rooms WHERE azit_id = ? AND deleted_at IS NULL)", azitID). // 지운 방은 안 센다
		Group("m.room_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[int64]int64, len(rows))
	for _, row := range rows {
		out[row.RoomID] = row.Count
	}
	return out, nil
}
