package notification

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/TogetherBros/azit/server/internal/db"
	"gorm.io/gorm"
)

// 알림이 이름을 찾아 쓰는 표만 흉내 낸다 (azit, room 패키지를 끌어오지 않으려고).
type testAzit struct {
	AzitID int64 `gorm:"primaryKey"`
	Name   string
}

func (testAzit) TableName() string { return "azits" }

type testRoom struct {
	RoomID int64 `gorm:"primaryKey"`
	Name   string
}

func (testRoom) TableName() string { return "rooms" }

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("DB 열기 실패: %v", err)
	}
	sqlDB, _ := conn.DB()
	t.Cleanup(func() { sqlDB.Close() }) // 윈도우는 열린 파일을 못 지운다
	if err := conn.AutoMigrate(&Notification{}, &testAzit{}, &testRoom{}); err != nil {
		t.Fatalf("AutoMigrate 실패: %v", err)
	}
	conn.Create(&testAzit{AzitID: 1, Name: "의성"})
	conn.Create(&testRoom{RoomID: 7, Name: "게임방"})
	return conn
}

func TestInsert_FillsNamesAndCutsPreview(t *testing.T) {
	conn := newTestDB(t)
	repo := NewRepository(conn)

	long := strings.Repeat("가", PreviewLen+5)
	err := Insert(conn, []Notification{
		{AccountID: 2, Kind: KindShake, AzitID: 1, RoomID: 7, ActorID: 1, ActorNick: "문경", Preview: long},
		{AccountID: 2, Kind: KindJoin, AzitID: 1, ActorID: 3, ActorNick: "백호"},
	})
	if err != nil {
		t.Fatalf("Insert 실패: %v", err)
	}

	got, err := repo.List(t.Context(), 2, ListLimit)
	if err != nil {
		t.Fatalf("List 실패: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("알림 수 = %d, want 2", len(got))
	}
	// 최신 순: join 이 먼저
	if got[0].Kind != KindJoin || got[1].Kind != KindShake {
		t.Errorf("순서 = %s, %s, want join, shake", got[0].Kind, got[1].Kind)
	}
	shake := got[1]
	if shake.AzitName != "의성" || shake.RoomName != "게임방" {
		t.Errorf("이름 = %q, %q", shake.AzitName, shake.RoomName)
	}
	if want := strings.Repeat("가", PreviewLen) + "…"; shake.Preview != want {
		t.Errorf("preview 길이 = %d 글자, want %d + …", len([]rune(shake.Preview)), PreviewLen)
	}
}

func TestListLimitAndReadAll(t *testing.T) {
	conn := newTestDB(t)
	repo := NewRepository(conn)
	ctx := t.Context()

	notes := make([]Notification, 0, ListLimit+3)
	for range ListLimit + 3 {
		notes = append(notes, Notification{AccountID: 2, Kind: KindShake, AzitID: 1})
	}
	notes = append(notes, Notification{AccountID: 9, Kind: KindKick, AzitID: 1}) // 남의 알림
	if err := Insert(conn, notes); err != nil {
		t.Fatal(err)
	}

	got, _ := repo.List(ctx, 2, ListLimit)
	if len(got) != ListLimit {
		t.Errorf("목록 = %d개, want %d", len(got), ListLimit)
	}
	if n, _ := repo.CountUnread(ctx, 2); n != int64(ListLimit+3) {
		t.Errorf("안 읽은 수 = %d, want %d (목록 밖 것도 센다)", n, ListLimit+3)
	}

	if err := repo.ReadAll(ctx, 2); err != nil {
		t.Fatal(err)
	}
	if n, _ := repo.CountUnread(ctx, 2); n != 0 {
		t.Errorf("모두 읽음 뒤 = %d, want 0", n)
	}
	if n, _ := repo.CountUnread(ctx, 9); n != 1 {
		t.Errorf("남의 알림까지 읽음 처리됨: %d, want 1", n)
	}
}
