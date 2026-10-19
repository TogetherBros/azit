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
	if err := conn.AutoMigrate(&Member{}, &Ban{}, &notification.Notification{}, &testAzit{}); err != nil {
		t.Fatal(err)
	}

	conn.Create(&testAzit{AzitID: 1, Name: "의성"})
	admin := Member{AzitID: 1, AccountID: 1, NickName: "관리자", Role: RoleAdmin}
	conn.Create(&admin)
	target := Member{AzitID: 1, AccountID: 2, NickName: "백호", Role: RoleMember}
	conn.Create(&target)

	repo := NewRepository(conn)
	returned, err := repo.Kick(t.Context(), admin, target, time.Now(), false)
	if err != nil {
		t.Fatalf("Kick 실패: %v", err)
	}
	// 돌려받은 알림에 저장된 id 와 이름이 채워져 있어야 실시간으로 그대로 보낼 수 있다
	if len(returned) != 1 || returned[0].NotificationID == 0 || returned[0].AzitName != "의성" {
		t.Errorf("돌려받은 알림 = %+v", returned)
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

// 막고 내보내면 막은 줄이 생기고, 되돌리기(Restore)는 다시 들여보내며 막은 것도 푼다.
// 풀기(Unban) 뒤 다시 막기(Reban)도 된다.
func TestBan_KickRestoreUnban(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := conn.DB()
	t.Cleanup(func() { sqlDB.Close() })
	if err := conn.AutoMigrate(&Member{}, &Ban{}, &notification.Notification{}, &testAzit{}); err != nil {
		t.Fatal(err)
	}
	conn.Create(&testAzit{AzitID: 1, Name: "의성"})
	admin := Member{AzitID: 1, AccountID: 1, NickName: "관리자", Role: RoleAdmin}
	target := Member{AzitID: 1, AccountID: 2, NickName: "백호", Role: RoleMember}
	conn.Create(&admin)
	conn.Create(&target)
	repo := NewRepository(conn)
	ctx := t.Context()

	if _, err := repo.Kick(ctx, admin, target, time.Now(), true); err != nil {
		t.Fatal(err)
	}
	if banned, _ := IsBanned(conn, 1, 2); !banned {
		t.Fatal("막고 내보냈는데 막힌 줄이 없음")
	}
	bans, _ := repo.ListBans(ctx, 1)
	if len(bans) != 1 || bans[0].Nickname != "백호" || bans[0].BannedByNick != "관리자" {
		t.Errorf("막은 목록 = %+v", bans)
	}

	// 되돌리기: 다시 멤버 + 막은 것 풀림
	if err := repo.Restore(ctx, 1, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.FindActive(ctx, 1, 2); err != nil {
		t.Errorf("되돌린 뒤에도 멤버가 아님: %v", err)
	}
	if banned, _ := IsBanned(conn, 1, 2); banned {
		t.Error("되돌렸는데 아직 막혀 있음")
	}

	// 풀기 → 다시 막기
	b := Ban{AzitID: 1, AccountID: 3, Nickname: "남", BannedByNick: "관리자"}
	if err := repo.Reban(ctx, b); err != nil {
		t.Fatal(err)
	}
	if err := repo.Unban(ctx, 1, 3); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.FindBan(ctx, 1, 3); err != ErrNotBanned {
		t.Errorf("푼 뒤 FindBan err = %v, want ErrNotBanned", err)
	}
}

func TestCoMemberIDs(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := conn.DB()
	t.Cleanup(func() { sqlDB.Close() })
	if err := conn.AutoMigrate(&Member{}); err != nil {
		t.Fatal(err)
	}
	left := time.Now()
	conn.Create(&Member{AzitID: 1, AccountID: 1})
	conn.Create(&Member{AzitID: 1, AccountID: 2})
	conn.Create(&Member{AzitID: 2, AccountID: 1})
	conn.Create(&Member{AzitID: 2, AccountID: 2}) // 두 아지트에 같이 있어도 한 번만
	conn.Create(&Member{AzitID: 2, AccountID: 3})
	conn.Create(&Member{AzitID: 2, AccountID: 4, LeftAt: &left}) // 나간 사람 빼고
	conn.Create(&Member{AzitID: 3, AccountID: 5})                // 남의 아지트 빼고

	ids, err := NewRepository(conn).CoMemberIDs(t.Context(), 1)
	if err != nil {
		t.Fatal(err)
	}
	got := map[int64]bool{}
	for _, id := range ids {
		got[id] = true
	}
	if len(ids) != 2 || !got[2] || !got[3] {
		t.Errorf("같은 아지트 사람 = %v, want [2 3]", ids)
	}
}
