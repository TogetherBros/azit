package client

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// 진짜 서버에 붙어 보는 테스트. 서버를 띄우고 환경 변수를 줄 때만 돈다:
//
//	AZIT_E2E_URL=http://127.0.0.1:18081 AZIT_E2E_TOKEN=<AT> AZIT_E2E_ROOM=/v1/azits/1/rooms/1 go test ./internal/client/ -run E2E
//
// 붙은 다음 그 방에 메시지를 보내고, 웹소켓으로 message.created 가 오는지 본다.
func TestDialEvents_E2E(t *testing.T) {
	base, token, room := os.Getenv("AZIT_E2E_URL"), os.Getenv("AZIT_E2E_TOKEN"), os.Getenv("AZIT_E2E_ROOM")
	if base == "" || token == "" || room == "" {
		t.Skip("AZIT_E2E_URL / AZIT_E2E_TOKEN / AZIT_E2E_ROOM 이 없어서 건너뜀")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c := New(base)

	if _, err := c.DialEvents(ctx, "틀린-토큰"); err == nil {
		t.Fatal("틀린 토큰인데 붙음")
	} else if apiErr, ok := err.(*APIError); !ok || apiErr.Status != 401 {
		t.Fatalf("틀린 토큰 에러 = %v, want 401 APIError (그래야 토큰 갱신이 돈다)", err)
	}

	conn, err := c.DialEvents(ctx, token)
	if err != nil {
		t.Fatalf("접속 실패: %v", err)
	}
	defer conn.CloseNow()

	body := map[string]any{"temp_id": "e2e-" + time.Now().Format("150405.000"), "body": "실시간 확인"}
	if err := c.Call(ctx, "POST", room+"/messages", token, body, nil); err != nil {
		t.Fatalf("메시지 보내기 실패: %v", err)
	}

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("이벤트 못 받음: %v", err)
	}
	if got := string(data); !strings.Contains(got, `"type":"message.created"`) || !strings.Contains(got, "실시간 확인") {
		t.Errorf("받은 것 = %s", got)
	}
}
