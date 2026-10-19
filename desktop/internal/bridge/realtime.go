package bridge

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/TogetherBros/azit/desktop/internal/client"
	"github.com/coder/websocket"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// React 로 보내는 Wails 이벤트 이름.
const (
	EventRealtime = "rt"      // data: 서버가 보낸 JSON 한 덩어리 그대로 ({"type": ..., "data": ...})
	EventRTOpen   = "rt:open" // 연결됐다 (처음이든 다시든). React 는 이때 목록을 새로 받아 끊겨 있던 사이를 메운다
)

const (
	retryMin  = time.Second      // 끊기면 1초 뒤 다시
	retryMax  = 30 * time.Second // 계속 실패하면 1→2→4…초, 최대 30초 간격
	waitLogin = 2 * time.Second  // 로그인 전이면 이 간격으로 다시 본다
)

// Realtime 은 서버 실시간 연결을 앱이 켜져 있는 내내 유지하고, 받은 이벤트를 React 로 넘긴다.
// React 는 이 서비스를 직접 부르지 않는다. Wails 이벤트(rt, rt:open)만 듣는다.
type Realtime struct {
	api  *client.Client
	auth *AuthService

	mu     sync.Mutex
	cancel context.CancelFunc // 지금 연결을 끊는 스위치 (로그인·로그아웃 때 다시 붙으려고)
}

// NewRealtime 은 실시간 연결을 만들고, 로그인·로그아웃 때 다시 붙도록 AuthService 에 연결해 둔다.
func NewRealtime(api *client.Client, auth *AuthService) *Realtime {
	r := &Realtime{api: api, auth: auth}
	auth.onChange = r.Reconnect
	return r
}

// Run 은 앱이 끝날 때까지 연결을 유지한다. main 에서 고루틴으로 한 번 띄운다.
func (r *Realtime) Run(ctx context.Context) {
	wait := retryMin
	for ctx.Err() == nil {
		if !r.auth.IsLoggedIn() {
			sleep(ctx, waitLogin)
			continue
		}

		connCtx, cancel := context.WithCancel(ctx)
		r.mu.Lock()
		r.cancel = cancel
		r.mu.Unlock()

		err := r.serve(connCtx)
		cancel()
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			log.Printf("실시간 연결 끊김: %v (%s 뒤 다시)", err, wait)
			sleep(ctx, wait)
			wait = min(wait*2, retryMax)
			continue
		}
		wait = retryMin // 잘 붙어 있다가 끊긴 거면 바로 다시
	}
}

// Reconnect 는 지금 연결을 끊고 새로 붙게 한다. 로그인·로그아웃 때 부른다 (다른 계정의 이벤트가 이어서 오지 않게).
func (r *Realtime) Reconnect() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel != nil {
		r.cancel()
	}
}

// serve 는 한 번 붙어서 끊길 때까지 받은 이벤트를 React 로 넘긴다.
// 붙자마자 rt:open 을 보내서, React 가 끊겨 있던 동안 놓친 걸 HTTP 로 다시 받게 한다.
func (r *Realtime) serve(ctx context.Context) error {
	var conn *websocket.Conn
	err := r.auth.withToken(ctx, func(token string) error { // 토큰이 만료돼 401 이면 갱신하고 한 번 더
		c, err := r.api.DialEvents(ctx, token)
		conn = c
		return err
	})
	if err != nil {
		return err
	}
	defer conn.CloseNow()

	emit(EventRTOpen, "")
	for {
		_, data, err := conn.Read(ctx) // 서버의 ping 에 대한 pong 은 라이브러리가 읽는 중에 알아서 보낸다
		if err != nil {
			if ctx.Err() != nil {
				return nil // 우리가 끊은 것 (Reconnect, 앱 종료)
			}
			return err
		}
		emit(EventRealtime, string(data))
	}
}

func emit(name, data string) {
	if app := application.Get(); app != nil {
		app.Event.Emit(name, data)
	}
}

// sleep 은 d 만큼 기다리되, ctx 가 끝나면 바로 돌아온다.
func sleep(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}
