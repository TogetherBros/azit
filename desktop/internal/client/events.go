package client

import (
	"context"
	"net/http"
	"strings"

	"github.com/coder/websocket"
)

// DialEvents 는 서버 실시간 연결(GET /v1/ws)에 붙는다. 토큰은 HTTP 처럼 Authorization 헤더로 보낸다.
// 토큰이 만료돼 거절되면 *APIError{Status: 401} 을 돌려준다 (AuthService.withToken 이 보고 갱신하게).
func (c *Client) DialEvents(ctx context.Context, token string) (*websocket.Conn, error) {
	url := "ws" + strings.TrimPrefix(c.baseURL, "http") + "/v1/ws" // http://… → ws://…, https://… → wss://…

	conn, res, err := websocket.Dial(ctx, url, &websocket.DialOptions{
		HTTPHeader: http.Header{"Authorization": {"Bearer " + token}},
	})
	if err != nil {
		if res != nil && res.StatusCode == http.StatusUnauthorized {
			return nil, &APIError{Status: http.StatusUnauthorized, Code: "unauthorized", Message: "다시 로그인해 주세요"}
		}
		return nil, err
	}
	conn.SetReadLimit(1 << 20) // 이벤트 하나는 작다. 1MB 넘으면 이상한 것
	return conn, nil
}
