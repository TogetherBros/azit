package member

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/TogetherBros/azit/server/internal/db"
	"github.com/TogetherBros/azit/server/internal/notification"
)

type testAzit struct {
	AzitID int64 `gorm:"primaryKey"`
	Name   string
}

func (testAzit) TableName() string { return "azits" }

// 내보내면 그 사람은 나간 상태가 되고, 그 사람에게만 "누가 내보냈어요" 알림이 간다.
func TestKick_LeavesAndNotifiesTarget(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := conn.DB()
	t.Cleanup(func() { sqlDB.Close() })
	if err := conn.AutoMigrate(&Member{}, &notification.Notification{}, &testAzit{}); err != nil {
		t.Fatal(err)
	}

	conn.Create(&testAzit{AzitID: 1, Name: "의성"})
	admin := Member{AzitID: 1, AccountID: 1, NickName: "관리자", Role: RoleAdmin}
	conn.Create(&admin)
	conn.Create(&Member{AzitID: 1, AccountID: 2, NickName: "백호", Role: RoleMember})

	repo := NewRepository(conn)
	if err := repo.Kick(t.Context(), admin, 2, time.Now()); err != nil {
		t.Fatalf("Kick 실패: %v", err)
	}

	if _, err := repo.FindActive(t.Context(), 1, 2); err != ErrNotMember {
		t.Errorf("내보낸 뒤 FindActive err = %v, want ErrNotMember", err)
	}

	var notes []notification.Notification
	conn.Find(&notes)
	if len(notes) != 1 {
		t.Fatalf("알림 수 = %d, want 1", len(notes))
	}
	n := notes[0]
	if n.AccountID != 2 || n.Kind != notification.KindKick || n.ActorNick != "관리자" || n.AzitName != "의성" {
		t.Errorf("알림 = %+v", n)
	}
}
