package changelog

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/TogetherBros/azit/server/internal/db"
)

// 기록 → 목록 → 되돌리기 → 두 번째 되돌리기는 거절, 되돌리는 법이 없는 종류도 거절.
func TestRecordListUndo(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := conn.DB()
	t.Cleanup(func() { sqlDB.Close() })
	if err := conn.AutoMigrate(&Entry{}); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(conn)
	ctx := t.Context()

	var undone struct {
		name   string
		target int64
	}
	repo.Handle(KindRoomRename, func(_ context.Context, e Entry, _ int64) error {
		var before struct{ Name string }
		if err := DecodeBefore(e, &before); err != nil {
			return err
		}
		undone.name, undone.target = before.Name, e.TargetID
		return nil
	})

	repo.Record(ctx, Entry{AzitID: 1, ActorID: 1, ActorNick: "주인", Kind: KindRoomCreate, TargetID: 7, Summary: "만듦"}, nil)
	repo.Record(ctx, Entry{AzitID: 1, ActorID: 1, ActorNick: "주인", Kind: KindRoomRename, TargetID: 7, Summary: "이름"}, map[string]string{"name": "옛 방"})
	repo.Record(ctx, Entry{AzitID: 2, Kind: KindRoomRename}, map[string]string{"name": "남의 아지트"})

	list, err := repo.List(ctx, 1, 10)
	if err != nil || len(list) != 2 || list[0].Kind != KindRoomRename {
		t.Fatalf("목록 = %+v, err %v (최신 순, 아지트 1 것만)", list, err)
	}
	if !repo.Undoable(KindRoomRename) || repo.Undoable(KindRoomCreate) {
		t.Error("되돌릴 수 있는 종류 판단이 틀림")
	}

	e, err := repo.Undo(ctx, 1, list[0].EntryID, 2, "관리자")
	if err != nil {
		t.Fatalf("되돌리기 실패: %v", err)
	}
	if undone.name != "옛 방" || undone.target != 7 || e.UndoneAt == nil || e.UndoneByNick != "관리자" {
		t.Errorf("되돌린 값 = %+v, 기록 = %+v", undone, e)
	}

	if _, err := repo.Undo(ctx, 1, list[0].EntryID, 2, "관리자"); !errors.Is(err, ErrAlreadyUndone) {
		t.Errorf("두 번 되돌리기 err = %v, want ErrAlreadyUndone", err)
	}
	if _, err := repo.Undo(ctx, 1, list[1].EntryID, 2, "관리자"); !errors.Is(err, ErrNotUndoable) {
		t.Errorf("방 만들기 되돌리기 err = %v, want ErrNotUndoable", err)
	}
	if _, err := repo.Undo(ctx, 2, list[0].EntryID, 2, "관리자"); !errors.Is(err, ErrNotFound) {
		t.Errorf("다른 아지트 기록 err = %v, want ErrNotFound", err)
	}

	// nil 저장소는 아무것도 안 하고 터지지도 않는다 (테스트에서 기록이 필요 없을 때)
	var none *Repository
	none.Record(ctx, Entry{}, nil)
}
