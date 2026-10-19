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
// fakePublisher 는 진짜 허브 대신 넣는 가짜. 받은 이벤트를 적어 두기만 한다.
type fakePublisher struct {
	sent []published
}

type published struct {
	to []int64
	ev any
}

func (f *fakePublisher) Publish(accountIDs []int64, ev any) error {
	f.sent = append(f.sent, published{to: accountIDs, ev: ev})
	return nil
}

type sendEnv struct {
	db       *gorm.DB
	pub      *fakePublisher // 핸들러가 허브로 보낸 이벤트들
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
	pub := &fakePublisher{}
	h := NewHandler(NewRepository(conn), roomRepo, memberRepo, account.NewRepository(conn), meme.NewRepository(conn), r2, pub)

	p := auth.NewTokenProvider("test-secret", time.Minute, time.Hour)
	token, err := p.IssueAccessToken(acc.ID)
	must(t, err)

	app := fiber.New(fiber.Config{ErrorHandler: httperr.Handler}) // 실제 서버처럼 httperr 를 상태 코드로
	h.Routes(app.Group("/v1/azits", auth.RequireAuth(p)), member.RequireMember(memberRepo))

	return sendEnv{db: conn, pub: pub, app: app, token: token, roomID: rm.RoomID, memeID: mine.MemeID, other: other.MemeID, friend: friend.ID, outsider: outsider.ID}
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

// 메시지를 보내면 저장 뒤 message.created 가 아지트 지금 멤버 전원(나 포함)에게 한 번 나간다.
// 같은 번호표로 다시 보내면(재전송) 또 나가지 않는다.
func TestSendMessage_PublishesCreated(t *testing.T) {
	e := newSendEnv(t)

	status, m := e.send(t, proto.SendMessageRequest{TempID: "p-1", Body: "점심 누가 쏨"})
	if status != 201 {
		t.Fatalf("status = %d", status)
	}
	if len(e.pub.sent) != 1 {
		t.Fatalf("나간 이벤트 수 = %d, want 1", len(e.pub.sent))
	}
	got := e.pub.sent[0]

	ev, ok := got.ev.(proto.Event)
	if !ok || ev.Type != proto.EventMessageCreated {
		t.Fatalf("이벤트 = %#v, want message.created", got.ev)
	}
	if data, ok := ev.Data.(proto.MessageResponse); !ok || data.MessageID != m.MessageID {
		t.Errorf("data = %#v, want 방금 보낸 메시지 %d", ev.Data, m.MessageID)
	}

	to := map[int64]bool{}
	for _, id := range got.to {
		to[id] = true
	}
	if !to[e.friend] || to[e.outsider] || len(got.to) != 2 {
		t.Errorf("받는 사람 = %v, want 나 + 백호 (다른 아지트 사람 빼고)", got.to)
	}

	// 재전송: 같은 번호표 → 새로 안 만들고, 이벤트도 안 나간다
	if status, _ := e.send(t, proto.SendMessageRequest{TempID: "p-1", Body: "점심 누가 쏨"}); status != 201 {
		t.Fatalf("재전송 status = %d", status)
	}
	if len(e.pub.sent) != 1 {
		t.Errorf("재전송 뒤 이벤트 수 = %d, want 1 그대로", len(e.pub.sent))
	}
}

// 고치기·지우기·맞장구는 message.updated 로 "어느 메시지가 바뀌었는지"만 알린다.
func TestUpdates_PublishMessageRef(t *testing.T) {
	e := newSendEnv(t)
	_, m := e.send(t, proto.SendMessageRequest{TempID: "u-1", Body: "처음"})
	e.pub.sent = nil // 보내기 이벤트는 빼고 본다

	do := func(method, path, body string) int {
		r := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		r.Header.Set("Authorization", "Bearer "+e.token)
		r.Header.Set("Content-Type", "application/json")
		res, err := e.app.Test(r)
		must(t, err)
		return res.StatusCode
	}
	base := fmt.Sprintf("/v1/azits/1/messages/%d", m.MessageID)

	steps := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"고치기", "PATCH", base, `{"body":"고침"}`},
		{"짤 맞장구 달기", "PUT", base + "/reactions", fmt.Sprintf(`{"meme_id":%d}`, e.memeID)},
		{"짤 맞장구 떼기", "DELETE", base + "/reactions", fmt.Sprintf(`{"meme_id":%d}`, e.memeID)},
		{"지우기", "DELETE", base, ""},
	}
	for i, s := range steps {
		if status := do(s.method, s.path, s.body); status >= 300 {
			t.Fatalf("%s: status %d", s.name, status)
		}
		if len(e.pub.sent) != i+1 {
			t.Fatalf("%s 뒤 이벤트 수 = %d, want %d", s.name, len(e.pub.sent), i+1)
		}
		ev := e.pub.sent[i].ev.(proto.Event)
		ref, ok := ev.Data.(proto.MessageRef)
		if ev.Type != proto.EventMessageUpdated || !ok || ref.MessageID != m.MessageID || ref.RoomID != e.roomID {
			t.Errorf("%s: 이벤트 = %#v", s.name, ev)
		}
	}

	// 같은 내용으로 고치면 바뀐 게 없으니 안 나간다
	before := len(e.pub.sent)
	_, m2 := e.send(t, proto.SendMessageRequest{TempID: "u-2", Body: "그대로"})
	e.pub.sent = e.pub.sent[:before]
	if status := do("PATCH", fmt.Sprintf("/v1/azits/1/messages/%d", m2.MessageID), `{"body":"그대로"}`); status != 200 {
		t.Fatalf("같은 내용 고치기 status %d", status)
	}
	if len(e.pub.sent) != before {
		t.Errorf("바뀐 게 없는데 이벤트가 나감")
	}
}

// @로 부르면 message.created(아지트 전원)와 별개로, 불린 사람에게만 notification.created 가 간다.
func TestSendMessage_PublishesShakeNotification(t *testing.T) {
	e := newSendEnv(t)

	_, m := e.send(t, proto.SendMessageRequest{TempID: "n-p", Body: "@백호 점심", ShakeIDs: []int64{e.friend}})

	var notes []published
	for _, p := range e.pub.sent {
		if ev := p.ev.(proto.Event); ev.Type == proto.EventNotificationCreated {
			notes = append(notes, p)
		}
	}
	if len(notes) != 1 {
		t.Fatalf("알림 이벤트 수 = %d, want 1", len(notes))
	}
	if len(notes[0].to) != 1 || notes[0].to[0] != e.friend {
		t.Errorf("받는 사람 = %v, want 백호만", notes[0].to)
	}
	n, ok := notes[0].ev.(proto.Event).Data.(proto.NotificationResponse)
	if !ok || n.NotificationID == 0 || n.Kind != "shake" || n.MessageID != m.MessageID || n.RoomName != "아지트" || n.ActorNickname != "문경" {
		t.Errorf("알림 = %#v (저장된 id·이름이 채워져 있어야 함)", n)
	}
}