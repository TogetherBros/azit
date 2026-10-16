package notification

import (
	"context"
	"unicode/utf8"

	"gorm.io/gorm"
)

// PreviewLen 은 알림에 남기는 메시지 앞부분 글자 수.
const PreviewLen = 40

// ListLimit 은 목록으로 돌려주는 최근 알림 수.
const ListLimit = 50

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Insert 는 알림 여러 줄을 tx 안에서 저장한다. 메시지 보내기·초대 코드로 들어오기·내보내기의
// 트랜잭션 안에서 불러서, 이벤트가 롤백되면 알림도 같이 사라지게 한다.
// 아지트 이름과 방 이름은 여기서 채운다 (부르는 쪽마다 따로 찾지 않게).
func Insert(tx *gorm.DB, notes []Notification) error {
	if len(notes) == 0 {
		return nil
	}

	azitNames := map[int64]string{}
	roomNames := map[int64]string{}
	for i := range notes {
		n := &notes[i]
		if n.AzitName == "" {
			name, ok := azitNames[n.AzitID]
			if !ok {
				if err := tx.Table("azits").Where("azit_id = ?", n.AzitID).Select("name").Scan(&name).Error; err != nil {
					return err
				}
				azitNames[n.AzitID] = name
			}
			n.AzitName = name
		}
		if n.RoomID > 0 && n.RoomName == "" {
			name, ok := roomNames[n.RoomID]
			if !ok {
				if err := tx.Table("rooms").Where("room_id = ?", n.RoomID).Select("name").Scan(&name).Error; err != nil {
					return err
				}
				roomNames[n.RoomID] = name
			}
			n.RoomName = name
		}
		n.Preview = cut(n.Preview, PreviewLen)
	}
	return tx.Create(&notes).Error
}

// List 는 내 알림을 최신 순으로 limit 개.
func (r *Repository) List(ctx context.Context, accountID int64, limit int) ([]Notification, error) {
	var out []Notification
	err := r.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		Order("notification_id DESC").
		Limit(limit).
		Find(&out).Error
	return out, err
}

// CountUnread 는 안 읽은 알림 수.
func (r *Repository) CountUnread(ctx context.Context, accountID int64) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&Notification{}).
		Where("account_id = ? AND is_read = ?", accountID, false).
		Count(&n).Error
	return n, err
}

// ReadAll 은 내 알림을 전부 읽음으로 바꾼다.
func (r *Repository) ReadAll(ctx context.Context, accountID int64) error {
	return r.db.WithContext(ctx).Model(&Notification{}).
		Where("account_id = ? AND is_read = ?", accountID, false).
		Update("is_read", true).Error
}

// cut 은 글자(rune) 기준으로 앞 n 자만 남긴다. 잘랐으면 끝에 … 를 붙인다.
func cut(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n]) + "…"
}
