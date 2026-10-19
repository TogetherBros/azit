// Package changelog 는 아지트 설정에서 "누가 무엇을 바꿨나" 를 남기고, 되돌린다.
//
// 바꾸는 쪽(아지트·방·멤버·짤 핸들러)은 저장이 끝난 뒤 Record 로 한 줄을 남긴다.
// 되돌리기는 종류(Kind)마다 다르게 해야 해서, 실제 되돌리는 함수(Undoer)는 wire.go 가 종류별로 꽂아 준다.
// (changelog 가 room·member·meme 를 import 하면 서로 물고 물려서, 반대로 함수만 받는다)
package changelog

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"gorm.io/gorm"
)

// 기록 종류.
const (
	KindAzitInfo    = "azit.info"     // 아지트 이름·색·소개·배너. Before = 바꾸기 전 값들
	KindRoomCreate  = "room.create"   // 방 만들기 (되돌리기 없음: 지우려면 방 지우기로)
	KindRoomRename  = "room.rename"   // 방 이름. Before = {"name": 옛 이름}
	KindRoomOrder   = "room.order"    // 방 순서. Before = {"room_ids": 옛 순서}
	KindRoomDelete  = "room.delete"   // 방 지우기. 되돌리면 다시 살린다 (맨 아래로)
	KindMemberRole  = "member.role"   // 역할. Before = {"role": 옛 역할}
	KindMemberKick  = "member.kick"   // 내보내기 (막기 포함). 되돌리면 다시 들여보내고 막은 것도 푼다
	KindMemberUnban = "member.unban"  // 막은 것 풀기. 되돌리면 다시 막는다
	KindMemeCreate  = "meme.create"   // 짤 올리기. 되돌리면 뺀다
	KindMemeDelete  = "meme.delete"   // 짤 빼기. 되돌리면 다시 넣는다
	KindReacts      = "reacts.update" // 맞장구 목록. Before = 옛 목록
)

var (
	ErrNotFound      = errors.New("변경 기록이 없어요")
	ErrAlreadyUndone = errors.New("이미 되돌린 기록이에요")
	ErrNotUndoable   = errors.New("되돌릴 수 없는 기록이에요")
)

// Entry 는 기록 한 줄.
type Entry struct {
	EntryID      int64  `gorm:"primaryKey"`
	AzitID       int64  `gorm:"not null;index"`
	ActorID      int64  `gorm:"not null"`
	ActorNick    string `gorm:"not null;default:''"` // 그때의 아지트 닉네임
	Kind         string `gorm:"not null"`
	TargetID     int64  `gorm:"not null;default:0"`  // 방 id, 계정 id, 짤 id …
	Summary      string `gorm:"not null;default:''"` // 화면에 보일 한 줄 ("방 이름을 '게임방' 으로")
	Before       string `gorm:"not null;default:''"` // 되돌릴 때 쓰는 이전 값 (JSON)
	CreatedAt    time.Time
	UndoneAt     *time.Time
	UndoneByNick string `gorm:"not null;default:''"`
}

func (Entry) TableName() string { return "change_logs" }

// Undoer 는 한 종류의 기록을 되돌린다. actor 는 되돌리는 사람 (되돌리기 자체도 그 사람이 한 일이 된다).
type Undoer func(ctx context.Context, e Entry, actorID int64) error

type Repository struct {
	db      *gorm.DB
	undoers map[string]Undoer
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db, undoers: map[string]Undoer{}}
}

// Handle 은 kind 기록을 되돌리는 함수를 꽂는다 (wire.go).
func (r *Repository) Handle(kind string, f Undoer) {
	r.undoers[kind] = f
}

// Undoable 은 이 종류가 되돌릴 수 있는 것인지.
func (r *Repository) Undoable(kind string) bool {
	if r == nil {
		return false
	}
	_, ok := r.undoers[kind]
	return ok
}

// Record 는 기록 한 줄을 남긴다. 실패해도 원래 일은 이미 끝났으니 로그만 남긴다.
// before 는 되돌릴 때 쓸 값 (JSON 으로 바꿔 저장). r 이 nil 이면 아무것도 안 한다 (테스트).
func (r *Repository) Record(ctx context.Context, e Entry, before any) {
	if r == nil {
		return
	}
	if before != nil {
		b, err := json.Marshal(before)
		if err != nil {
			log.Printf("변경 기록 실패: %v", err)
			return
		}
		e.Before = string(b)
	}
	if err := r.db.WithContext(ctx).Create(&e).Error; err != nil {
		log.Printf("변경 기록 실패: %v", err)
	}
}

// List 는 아지트 기록을 최신 순으로 limit 개.
func (r *Repository) List(ctx context.Context, azitID int64, limit int) ([]Entry, error) {
	var out []Entry
	err := r.db.WithContext(ctx).Where("azit_id = ?", azitID).Order("entry_id DESC").Limit(limit).Find(&out).Error
	return out, err
}

// Undo 는 기록 하나를 되돌리고 "되돌림" 표시를 남긴다.
func (r *Repository) Undo(ctx context.Context, azitID, entryID, actorID int64, actorNick string) (Entry, error) {
	var e Entry
	err := r.db.WithContext(ctx).Where("azit_id = ? AND entry_id = ?", azitID, entryID).First(&e).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Entry{}, ErrNotFound
	}
	if err != nil {
		return Entry{}, err
	}
	if e.UndoneAt != nil {
		return Entry{}, ErrAlreadyUndone
	}
	undo, ok := r.undoers[e.Kind]
	if !ok {
		return Entry{}, ErrNotUndoable
	}
	if err := undo(ctx, e, actorID); err != nil {
		return Entry{}, err
	}

	now := time.Now()
	err = r.db.WithContext(ctx).Model(&Entry{}).Where("entry_id = ?", e.EntryID).
		Updates(map[string]any{"undone_at": now, "undone_by_nick": actorNick}).Error
	e.UndoneAt, e.UndoneByNick = &now, actorNick
	return e, err
}

// DecodeBefore 는 Before JSON 을 v 로 푼다. Undoer 들이 쓴다.
func DecodeBefore(e Entry, v any) error {
	if e.Before == "" {
		return ErrNotUndoable
	}
	return json.Unmarshal([]byte(e.Before), v)
}
