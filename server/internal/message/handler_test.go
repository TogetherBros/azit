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
	"github.com/TogetherBros/azit/server/internal/db"
	"github.com/TogetherBros/azit/server/internal/httperr"
	"github.com/TogetherBros/azit/server/internal/member"
	"github.com/TogetherBros/azit/server/internal/meme"
	"github.com/TogetherBros/azit/server/internal/room"
	"github.com/TogetherBros/azit/server/internal/storage"
	"github.com/gofiber/fiber/v3"
)

// sendEnv 는 메시지 보내기 테스트용 서버. 아지트 1 에 멤버 한 명, 방 하나, 짤 하나가 있다.
type sendEnv struct {
	app    *fiber.App
	token  string
	roomID int64
	memeID int64 // 아지트 1 짤
	other  int64 // 아지트 2 짤
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

	if err := conn.AutoMigrate(&account.Account{}, &member.Member{}, &room.Room{}, &Message{}, &Read{}, &Shake{}, &Reaction{}, &meme.Meme{}); err != nil {
		t.Fatalf("AutoMigrate 실패: %v", err)
	}

	acc := account.Account{Provider: account.ProviderGoogle, Subject: "sub-1", Email: "a@test.com", Nickname: "문경"}
	must(t, conn.Create(&acc).Error)
	must(t, conn.Create(&member.Member{AzitID: 1, AccountID: acc.ID, NickName: "문경", Role: member.RoleMember}).Error)

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

	return sendEnv{app: app, token: token, roomID: rm.RoomID, memeID: mine.MemeID, other: other.MemeID}
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
