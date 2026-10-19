package ws

import (
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/auth"
	"github.com/TogetherBros/azit/server/internal/hub"
	fws "github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
)

// 진짜 서버를 빈 포트에 띄워, 진짜 웹소켓 클라이언트로 붙어 본다.
// 토큰 없이 → 거절 / 토큰 있으면 → 허브에 넣은 이벤트가 소켓으로 온다 / 끊으면 → 허브에서 빠진다.
func TestServe_EndToEnd(t *testing.T) {
	p := auth.NewTokenProvider("test-secret", time.Minute, time.Hour)
	token, err := p.IssueAccessToken(7)
	if err != nil {
		t.Fatal(err)
	}

	hb := hub.New()
	app := fiber.New()
	NewHandler(hb, nil).Routes(app.Group("/v1/ws", auth.RequireAuth(p)))

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() { _ = app.Listener(ln, fiber.ListenConfig{DisableStartupMessage: true}) }()
	t.Cleanup(func() { _ = app.Shutdown() })
	url := "ws://" + ln.Addr().String() + "/v1/ws"

	// 1) 토큰 없이 → 401
	if _, res, err := fws.DefaultDialer.Dial(url, nil); err == nil || res == nil || res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("토큰 없이 접속: err=%v res=%v, want 401", err, res)
	}

	// 2) 토큰 있으면 붙는다
	conn, _, err := fws.DefaultDialer.Dial(url, http.Header{"Authorization": {"Bearer " + token}})
	if err != nil {
		t.Fatalf("접속 실패: %v", err)
	}
	defer conn.Close()
	waitFor(t, "허브 등록", func() bool { return hb.Online(7) })

	// 3) 허브에 넣으면 소켓으로 온다
	ev := proto.Event{Type: proto.EventMessageUpdated, Data: proto.MessageRef{AzitID: 1, RoomID: 2, MessageID: 3}}
	if err := hb.Publish([]int64{7}, ev); err != nil {
		t.Fatal(err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("받기 실패: %v", err)
	}
	if got := string(msg); !strings.Contains(got, `"type":"message.updated"`) || !strings.Contains(got, `"message_id":3`) {
		t.Errorf("받은 것 = %s", got)
	}

	// 4) 끊으면 허브에서 빠진다
	_ = conn.Close()
	waitFor(t, "허브 해지", func() bool { return !hb.Online(7) })
}

// waitFor 는 cond 가 참이 될 때까지 최대 2초 기다린다 (연결·해지는 다른 고루틴에서 일어나서).
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("%s 를 2초 안에 못 봄", what)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
