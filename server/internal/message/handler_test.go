package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/account"
	"github.com/TogetherBros/azit/server/internal/auth"
	"github.com/TogetherBros/azit/server/internal/azit"
	"github.com/TogetherBros/azit/server/internal/db"
	"github.com/TogetherBros/azit/server/internal/httperr"
	"github.com/TogetherBros/azit/server/internal/member"
	"github.com/TogetherBros/azit/server/internal/meme"
	"github.com/TogetherBros/azit/server/internal/notification"
	"github.com/TogetherBros/azit/server/internal/room"
	"github.com/TogetherBros/azit/server/internal/storage"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// sendEnv 는 메시지 보내기 테스트용 서버. 아지트 1 에 멤버 한 명, 방 하나, 짤 하나가 있다.
type sendEnv struct {
	db       *gorm.DB
	app      *fiber.App
	token    string
	roomID   int64
	memeID   int64 // 아지트 1 짤
	other    int64 // 아지트 2 짤
	friend   int64 // 아지트 1 의 다른 멤버 (@로 부를 수 있음)
	outsider int64 // 아지트 2 에만 있는 사람 (@로 부를 수 없음)
}

func newSendEnv(t *testing.T) sendEnv {
	t.Helper()

	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("DB 열기 실패: %v", err)
	}
	// 윈도우는 열린 파일을 못 지워서, TempDir 정리 전에 닫는다 (Cleanup 은 나중에 단 것부터 돈다)
	sqlDB, _ := conn.DB()
	t.Cleanup(func() { sqlDB.Close() })

	if err := conn.AutoMigrate(&account.Account{}, &member.Member{}, &room.Room{}, &Message{}, &Read{}, &Shake{}, &Reaction{}, &meme.Meme{}, &notification.Notification{}, &azit.Azit{}); err != nil {
		t.Fatalf("AutoMigrate 실패: %v", err)
	}

	acc := account.Account{Provider: account.ProviderGoogle, Subject: "sub-1", Email: "a@test.com", Nickname: "문경"}
	must(t, conn.Create(&acc).Error)
	must(t, conn.Create(&member.Member{AzitID: 1, AccountID: acc.ID, NickName: "문경", Role: member.RoleMember}).Error)

	friend := account.Account{Provider: account.ProviderGoogle, Subject: "sub-2", Email: "b@test.com", Nickname: "백호"}
	outsider := account.Account{Provider: account.ProviderGoogle, Subject: "sub-3", Email: "c@test.com", Nickname: "남"}
	must(t, conn.Create(&friend).Error)
	must(t, conn.Create(&outsider).Error)
	must(t, conn.Create(&member.Member{AzitID: 1, AccountID: friend.ID, NickName: "백호", Role: member.RoleMember}).Error)
	must(t, conn.Create(&member.Member{AzitID: 2, AccountID: outsider.ID, NickName: "남", Role: member.RoleMember}).Error)

	roomRepo := room.NewRepository(conn)
	rm := room.Room{AzitID: 1, Name: "아지트", CreatedBy: acc.ID}
	must(t, roomRepo.Create(t.Context(), &rm))

	mine := meme.Meme{AzitID: 1, Name: "킹받네", MediaKey: fmt.Sprintf("memes/%d/A.webp", acc.ID), CreatedBy: acc.ID}
	other := meme.Meme{AzitID: 2, Name: "남의 짤", MediaKey: "memes/99/B.webp", CreatedBy: 99}
	must(t, conn.Create(&mine).Error)
	must(t, conn.Create(&other).Error)

	r2, err := storage.NewR2("test", "key", "secret", "bucket") // 링크 서명만 하고 네트워크는 안 탄다
	must(t, err)

	memberRepo := member.NewRepository(conn)
	h := NewHandler(NewRepository(conn), roomRepo, memberRepo, account.NewRepository(conn), meme.NewRepository(conn), r2)

	p := auth.NewTokenProvider("test-secret", time.Minute, time.Hour)
	token, err := p.IssueAccessToken(acc.ID)
	must(t, err)

	app := fiber.New(fiber.Config{ErrorHandler: httperr.Handler}) // 실제 서버처럼 httperr 를 상태 코드로
	h.Routes(app.Group("/v1/azits", auth.RequireAuth(p)), member.RequireMember(memberRepo))

	return sendEnv{db: conn, app: app, token: token, roomID: rm.RoomID, memeID: mine.MemeID, other: other.MemeID, friend: friend.ID, outsider: outsider.ID}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// send 는 메시지를 보내고 상태 코드와 (성공이면) 응답을 돌려준다.
func (e sendEnv) send(t *testing.T, req proto.SendMessageRequest) (int, proto.MessageResponse) {
	t.Helper()
	body, _ := json.Marshal(req)
	r := httptest.NewRequest("POST", fmt.Sprintf("/v1/azits/1/rooms/%d/messages", e.roomID), bytes.NewReader(body))
	r.Header.Set("Authorization", "Bearer "+e.token)
	r.Header.Set("Content-Type", "application/json")
	res, err := e.app.Test(r)
	must(t, err)

	var out proto.MessageResponse
	if res.StatusCode == 201 {
		must(t, json.NewDecoder(res.Body).Decode(&out))
	}
	return res.StatusCode, out
}

func TestSendMessage_Meme(t *testing.T) {
	e := newSendEnv(t)

	status, m := e.send(t, proto.SendMessageRequest{TempID: "t-1", Kind: "meme", MemeID: e.memeID})
	if status != 201 {
		t.Fatalf("status = %d, want 201", status)
	}
	if m.Kind != "meme" || m.MemeID != e.memeID {
		t.Errorf("kind, meme_id = %q, %d, want meme, %d", m.Kind, m.MemeID, e.memeID)
	}
	if m.Body != "" {
		t.Errorf("body = %q, want \"\" (R2 키를 내보내면 안 됨)", m.Body)
	}
	if m.MediaURL == "" {
		t.Error("media_url 이 비었어요")
	}

	// 같은 번호표로 다시 보내면 같은 메시지
	status, again := e.send(t, proto.SendMessageRequest{TempID: "t-1", Kind: "meme", MemeID: e.memeID})
	if status != 201 || again.MessageID != m.MessageID {
		t.Errorf("재전송: status %d, id %d, want 201, %d", status, again.MessageID, m.MessageID)
	}
}

func TestSendMessage_Rejects(t *testing.T) {
	e := newSendEnv(t)

	cases := []struct {
		name string
		req  proto.SendMessageRequest
	}{
		{"다른 아지트 짤", proto.SendMessageRequest{TempID: "r-1", Kind: "meme", MemeID: e.other}},
		{"없는 짤", proto.SendMessageRequest{TempID: "r-2", Kind: "meme", MemeID: 12345}},
		{"남의 사진 키", proto.SendMessageRequest{TempID: "r-3", Kind: "photo", PhotoKey: "photos/99/X.webp"}},
		{"사진 폴더가 아닌 키", proto.SendMessageRequest{TempID: "r-4", Kind: "photo", PhotoKey: "memes/1/A.webp"}},
		{"모르는 종류", proto.SendMessageRequest{TempID: "r-5", Kind: "video", Body: "안녕"}},
		{"빈 글자", proto.SendMessageRequest{TempID: "r-6", Body: "  "}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if status, _ := e.send(t, tc.req); status != 400 {
				t.Errorf("status = %d, want 400", status)
			}
		})
	}
}

func TestSendMessage_TextStillWorks(t *testing.T) {
	e := newSendEnv(t)

	status, m := e.send(t, proto.SendMessageRequest{TempID: "x-1", Body: "안녕"})
	if status != 201 {
		t.Fatalf("status = %d, want 201", status)
	}
	if m.Kind != "text" || m.Body != "안녕" || m.MediaURL != "" {
		t.Errorf("got kind %q body %q media_url %q", m.Kind, m.Body, m.MediaURL)
	}
}

func TestSendMessage_Shake(t *testing.T) {
	e := newSendEnv(t)

	// 같은 사람을 두 번 불러도 한 번으로
	status, m := e.send(t, proto.SendMessageRequest{TempID: "s-1", Body: "@백호 점심", ShakeIDs: []int64{e.friend, e.friend}})
	if status != 201 {
		t.Fatalf("status = %d, want 201", status)
	}
	if len(m.ShakeIDs) != 1 || m.ShakeIDs[0] != e.friend {
		t.Errorf("shake_ids = %v, want [%d]", m.ShakeIDs, e.friend)
	}

	// 다른 아지트 사람은 못 부른다
	if status, _ := e.send(t, proto.SendMessageRequest{TempID: "s-2", Body: "@남", ShakeIDs: []int64{e.outsider}}); status != 400 {
		t.Errorf("아지트 밖 사람: status = %d, want 400", status)
	}

	// @전체
	status, m = e.send(t, proto.SendMessageRequest{TempID: "s-3", Body: "@전체 모여", ShakeAll: true})
	if status != 201 || !m.ShakeAll {
		t.Errorf("@전체: status %d, shake_all %v", status, m.ShakeAll)
	}
}

// 흔들면 흔들린 사람에게 알림이 남는다. @전체 는 나 빼고 전원, 나를 부른 건 알림 안 남김.
func TestSendMessage_ShakeNotifications(t *testing.T) {
	e := newSendEnv(t)
	count := func(accountID int64) int64 {
		var n int64
		e.db.Model(&notification.Notification{}).Where("account_id = ?", accountID).Count(&n)
		return n
	}

	if status, _ := e.send(t, proto.SendMessageRequest{TempID: "n-1", Body: "@백호 점심", ShakeIDs: []int64{e.friend}}); status != 201 {
		t.Fatalf("status = %d", status)
	}
	if count(e.friend) != 1 {
		t.Errorf("@백호: 백호 알림 %d, want 1", count(e.friend))
	}

	if status, _ := e.send(t, proto.SendMessageRequest{TempID: "n-2", Body: "@전체 모여", ShakeAll: true}); status != 201 {
		t.Fatalf("status = %d", status)
	}
	if count(e.friend) != 2 {
		t.Errorf("@전체: 백호 알림 %d, want 2", count(e.friend))
	}
	if count(e.outsider) != 0 {
		t.Errorf("다른 아지트 사람에게 알림이 감: %d", count(e.outsider))
	}

	var me int64
	e.db.Model(&notification.Notification{}).Where("actor_id = account_id").Count(&me)
	if me != 0 {
		t.Errorf("나에게 간 알림 %d, want 0", me)
	}

	var n notification.Notification
	e.db.Where("account_id = ?", e.friend).First(&n)
	if n.Kind != notification.KindShake || n.RoomName != "아지트" || n.Preview != "@백호 점심" || n.ActorNick != "문경" {
		t.Errorf("알림 = %+v", n)
	}
}
