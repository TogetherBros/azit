package room

import (
	"context"
	"errors"
	"slices"

	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("방을 찾을 수 없어요")
	ErrBadOrder = errors.New("방 순서가 맞지 않아요. 목록을 새로 받아 주세요")
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ListByAzit 는 아지트의 방을 순서대로 돌려준다. 순서가 같으면 먼저 만든 방이 위.
func (r *Repository) ListByAzit(ctx context.Context, azitID int64) ([]Room, error) {
	var rooms []Room
	err := r.db.WithContext(ctx).
		Where("azit_id = ?", azitID).
		Order("position, room_id").
		Find(&rooms).Error
	return rooms, err
}

// Create 는 방을 맨 아래에 만든다. room.AzitID, Name, CreatedBy 를 채워서 넘기면
// Position 과 RoomID 는 여기서 채워진다.
func (r *Repository) Create(ctx context.Context, room *Room) error {
	position, err := r.nextPosition(ctx, room.AzitID)
	if err != nil {
		return err
	}
	room.Position = position
	return r.db.WithContext(ctx).Create(room).Error
}

// nextPosition 은 그 아지트에서 가장 큰 position + 1. 방이 없으면 0.
func (r *Repository) nextPosition(ctx context.Context, azitID int64) (int, error) {
	var max int
	err := r.db.WithContext(ctx).
		Model(&Room{}).
		Where("azit_id = ?", azitID).
		Select("COALESCE(MAX(position), -1)").
		Scan(&max).Error
	if err != nil {
		return 0, err
	}
	return max + 1, nil
}

// FindByID 는 방 하나를 찾는다. 없으면 ErrNotFound.
func (r *Repository) FindByID(ctx context.Context, roomID int64) (Room, error) {
	var room Room
	err := r.db.WithContext(ctx).First(&room, roomID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Room{}, ErrNotFound
	}
	return room, err
}

// Rename 은 방 이름을 바꾼다.
func (r *Repository) Rename(ctx context.Context, roomID int64, name string) error {
	return r.db.WithContext(ctx).
		Model(&Room{}).
		Where("room_id = ?", roomID).
		Update("name", name).Error
}

// Reorder 는 아지트의 방 순서를 roomIDs 순서대로 0, 1, 2... 로 다시 매긴다.
// roomIDs 는 그 아지트 방 전부를 빠짐없이, 겹침 없이 담아야 한다. 아니면 ErrBadOrder.
func (r *Repository) Reorder(ctx context.Context, azitID int64, roomIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current []int64
		if err := tx.Model(&Room{}).Where("azit_id = ?", azitID).Pluck("room_id", &current).Error; err != nil {
			return err
		}

		// 보낸 목록과 지금 방 목록이 같은 집합인지 본다 (정렬해서 비교).
		want := slices.Sorted(slices.Values(roomIDs))
		slices.Sort(current)
		if !slices.Equal(want, current) {
			return ErrBadOrder
		}

		for i, id := range roomIDs {
			if err := tx.Model(&Room{}).Where("room_id = ?", id).Update("position", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
