// Package ws 는 실시간 연결(GET /v1/ws)을 받는다.
// 연결 하나마다 허브에 우편함(hub.Client)을 하나 등록하고, 고루틴 두 개를 돌린다:
//   - 쓰기: 우편함에서 꺼내 소켓으로 보낸다 + 30초마다 ping
//   - 읽기: 앱은 아무것도 안 보내지만, 끊김(과 pong)을 알아채려고 계속 읽는다
//
// 서버 → 앱 한 방향이다. 앱이 보내는 건 전부 HTTP.
package ws

import (
	"context"
	"time"

	"github.com/TogetherBros/azit/server/internal/auth"
	"github.com/TogetherBros/azit/server/internal/hub"
	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
)

const (
	pingEvery = 30 * time.Second // 이 간격으로 살아 있냐고 묻는다
	readWait  = 70 * time.Second // 이만큼 아무것도(pong 포함) 안 오면 끊긴 걸로 본다 (ping 두 번 놓침)
	writeWait = 10 * time.Second // 소켓 한 번 쓰는 데 이보다 오래 걸리면 끊는다
	readLimit = 512              // 앱이 보낼 일은 없지만, 큰 걸 보내면 끊는다

	// localsAccount 는 업그레이드 전에 계정 id 를 넘겨 두는 칸. websocket.Conn 은 문자열 키 locals 만 복사해 간다.
	localsAccount = "ws_account_id"
)

// PresenceFunc 는 누가 앱을 켰다(첫 연결)/껐다(마지막 연결이 끊김) 를 알리는 함수 (live.Live.Presence).
type PresenceFunc func(ctx context.Context, accountID int64, online bool)

type Handler struct {
	hub      *hub.Hub
	presence PresenceFunc // 없으면 접속 알림을 안 보낸다
}

func NewHandler(h *hub.Hub, presence PresenceFunc) *Handler {
	return &Handler{hub: h, presence: presence}
}

// Routes 는 /v1/ws 를 단다. r 에는 인증 미들웨어가 걸려 있어야 한다 (앱 Go 가 Authorization 헤더를 붙여 접속한다).
func (h *Handler) Routes(r fiber.Router) {
	r.Get("/", h.prepare, websocket.New(h.serve))
}

// prepare 는 웹소켓 요청인지 보고, 인증된 계정 id 를 연결 쪽으로 넘겨 둔다.
func (h *Handler) prepare(c fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}
	c.Locals(localsAccount, auth.AccountID(c))
	return c.Next()
}

// serve 는 연결 하나의 일생. 이 함수가 끝나면 연결도 끝난다.
func (h *Handler) serve(conn *websocket.Conn) {
	accountID, _ := conn.Locals(localsAccount).(int64)
	if accountID == 0 {
		_ = conn.Close()
		return
	}

	wasOnline := h.hub.Online(accountID) // 다른 창이 이미 붙어 있었나
	client := h.hub.Register(accountID)
	if !wasOnline {
		h.notifyPresence(accountID, true) // 첫 연결 → "접속 중"
	}
	defer conn.Close() // (2) 소켓 닫기 — 끝난 함수는 소켓을 열어 둔 채로 두면 안 된다
	defer func() {     // (1) 우편함 해지 → client.Send 가 닫힌다. 마지막 연결이었으면 "나감"
		h.hub.Unregister(client)
		if !h.hub.Online(accountID) {
			h.notifyPresence(accountID, false)
		}
	}()
	done := make(chan struct{})    // 읽기가 끝났다는 신호. 쓰기 고루틴이 이걸 보고 퇴근한다
	go writeLoop(conn, client.Send, done)

	readLoop(conn) // 끊길 때까지 여기서 기다린다
	close(done)
}

func (h *Handler) notifyPresence(accountID int64, online bool) {
	if h.presence != nil {
		h.presence(context.Background(), accountID, online)
	}
}

// writeLoop 은 우편함에서 꺼내 소켓에 쓰고, 30초마다 ping 을 보낸다.
// 우편함이 닫혔거나(허브가 너무 밀려서 끊음) 쓰기가 실패하면 소켓을 닫아 읽기 쪽도 끝나게 한다.
func writeLoop(conn *websocket.Conn, send <-chan []byte, done <-chan struct{}) {
	ticker := time.NewTicker(pingEvery)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-send:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "slow"))
				_ = conn.Close()
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				_ = conn.Close()
				return
			}
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				_ = conn.Close()
				return
			}
		case <-done:
			return
		}
	}
}

// readLoop 은 끊길 때까지 읽는다. pong 이 오면 기다리는 시간을 다시 늘린다.
// 70초 동안 아무것도 안 오면(앱이 죽었거나 인터넷이 끊김) ReadMessage 가 에러를 내고 끝난다.
func readLoop(conn *websocket.Conn) {
	conn.SetReadLimit(readLimit)
	_ = conn.SetReadDeadline(time.Now().Add(readWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(readWait))
	})
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}
