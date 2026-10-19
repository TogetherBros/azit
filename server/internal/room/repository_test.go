package room

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/TogetherBros/azit/server/internal/db"
)

// 기본 방(첫 방)은 못 지움 → 다른 방 지우기 → 목록에서 빠짐 → 되살리면 맨 아래로 돌아옴.
func TestDeleteRestore(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := conn.DB()
	t.Cleanup(func() { sqlDB.Close() })
	if err := conn.AutoMigrate(&Room{}); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(conn)
	ctx := t.Context()

	def := Room{AzitID: 1, Name: DefaultName, CreatedBy: 1}
	a := Room{AzitID: 1, Name: "다같이", CreatedBy: 1}
	b := Room{AzitID: 1, Name: "게임방", CreatedBy: 1}
	other := Room{AzitID: 2, Name: "남의 방", CreatedBy: 1}
	for _, r := range []*Room{&def, &a, &b, &other} {
		if err := repo.Create(ctx, r); err != nil {
			t.Fatal(err)
		}
	}

	if id, _ := repo.DefaultID(ctx, 1); id != def.RoomID {
		t.Errorf("DefaultID = %d, want %d", id, def.RoomID)
	}
	if err := repo.Delete(ctx, 1, def.RoomID); !errors.Is(err, ErrDefaultRoom) {
		t.Errorf("기본 방 지우기 err = %v, want ErrDefaultRoom", err)
	}
	if err := repo.Delete(ctx, 1, other.RoomID); !errors.Is(err, ErrNotFound) {
		t.Errorf("다른 아지트 방 지우기 err = %v, want ErrNotFound", err)
	}
	if err := repo.Delete(ctx, 1, a.RoomID); err != nil {
		t.Fatalf("지우기 실패: %v", err)
	}
	list, _ := repo.ListByAzit(ctx, 1)
	if len(list) != 2 || list[1].RoomID != b.RoomID {
		t.Fatalf("지운 뒤 목록 = %+v, 기본 방과 게임방만 남아야 함", list)
	}
	if _, err := repo.FindByID(ctx, a.RoomID); !errors.Is(err, ErrNotFound) {
		t.Errorf("지운 방 FindByID err = %v, want ErrNotFound", err)
	}
	if err := repo.Delete(ctx, 1, b.RoomID); err != nil {
		t.Fatalf("게임방 지우기 실패: %v", err)
	}
	if err := repo.Restore(ctx, 1, b.RoomID); err != nil {
		t.Fatalf("게임방 되살리기 실패: %v", err)
	}

	if err := repo.Restore(ctx, 1, a.RoomID); err != nil {
		t.Fatalf("되살리기 실패: %v", err)
	}
	list, _ = repo.ListByAzit(ctx, 1)
	if len(list) != 3 || list[2].RoomID != a.RoomID {
		t.Fatalf("되살린 뒤 목록 = %+v, 다같이가 맨 아래여야 함", list)
	}
	if err := repo.Restore(ctx, 1, a.RoomID); !errors.Is(err, ErrNotFound) {
		t.Errorf("이미 살아 있는 방 되살리기 err = %v, want ErrNotFound", err)
	}
}
