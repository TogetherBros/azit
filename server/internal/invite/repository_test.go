package invite

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/TogetherBros/azit/server/internal/db"
	"github.com/TogetherBros/azit/server/internal/member"
	"github.com/TogetherBros/azit/server/internal/notification"
)

type testAzit struct {
	AzitID int64 `gorm:"primaryKey"`
	Name   string
}

func (testAzit) TableName() string { return "azits" }

// 코드로 들어오면 원래 있던 멤버(나간 사람 빼고)에게만 입장 알림이 간다. 들어온 사람에게는 안 간다.
func TestJoin_NotifiesExistingMembers(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := conn.DB()
	t.Cleanup(func() { sqlDB.Close() })
	if err := conn.AutoMigrate(&Invite{}, &member.Member{}, &member.Ban{}, &notification.Notification{}, &testAzit{}); err != nil {
		t.Fatal(err)
	}

	left := time.Now()
	conn.Create(&testAzit{AzitID: 1, Name: "의성"})
	conn.Create(&member.Member{AzitID: 1, AccountID: 1, NickName: "주인", Role: member.RoleSuperAdmin})
	conn.Create(&member.Member{AzitID: 1, AccountID: 2, NickName: "백호", Role: member.RoleMember})
	conn.Create(&member.Member{AzitID: 1, AccountID: 3, NickName: "나간 사람", Role: member.RoleMember, LeftAt: &left})
	conn.Create(&Invite{AzitID: 1, Code: "ABCD2345", CreatedBy: 1})

	repo := NewRepository(conn)
	_, returned, err := repo.Join(t.Context(), JoinInput{Code: "ABCD2345", AccountID: 4, Nickname: "새친구", Now: time.Now()})
	if err != nil {
		t.Fatalf("Join 실패: %v", err)
	}
	// 커밋 뒤 실시간으로 보낼 수 있게, 저장된 알림(id 채워짐)을 돌려준다
	if len(returned) != 2 || returned[0].NotificationID == 0 {
		t.Errorf("돌려받은 알림 = %+v, want 2개 (id 채워짐)", returned)
	}

	var notes []notification.Notification
	conn.Order("account_id").Find(&notes)
	if len(notes) != 2 || notes[0].AccountID != 1 || notes[1].AccountID != 2 {
		t.Fatalf("알림 받은 사람 = %+v, want 1, 2", notes)
	}
	for _, n := range notes {
		if n.Kind != notification.KindJoin || n.ActorID != 4 || n.ActorNick != "새친구" || n.AzitName != "의성" {
			t.Errorf("알림 = %+v", n)
		}
	}
	// 막힌 사람은 새 코드로도 못 들어온다
	conn.Create(&member.Ban{AzitID: 1, AccountID: 9, Nickname: "막힘"})
	if _, _, err := repo.Join(t.Context(), JoinInput{Code: "ABCD2345", AccountID: 9, Nickname: "막힘", Now: time.Now()}); err != ErrBanned {
		t.Errorf("막힌 사람 Join err = %v, want ErrBanned", err)
	}
}
