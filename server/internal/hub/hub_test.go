package hub

import (
	"sync"
	"testing"
)

type testEvent struct {
	Type string `json:"type"`
}

// 받을 사람으로 고른 계정의 연결만 받는다. 한 사람의 연결이 둘이면 둘 다 받는다.
func TestPublish_OnlyTargets(t *testing.T) {
	h := New()
	a1 := h.Register(1)
	a2 := h.Register(1) // 1번이 창을 두 개 띄움
	b := h.Register(2)

	if err := h.Publish([]int64{1}, testEvent{Type: "hi"}); err != nil {
		t.Fatal(err)
	}

	for i, c := range []*Client{a1, a2} {
		select {
		case got := <-c.Send:
			if string(got) != `{"type":"hi"}` {
				t.Errorf("연결 %d 받은 것 = %s", i, got)
			}
		default:
			t.Errorf("연결 %d 가 못 받음", i)
		}
	}
	select {
	case got := <-b.Send:
		t.Errorf("2번이 받으면 안 되는데 받음: %s", got)
	default:
	}
}

// 빼면 Send 가 닫히고(쓰기 고루틴이 끝날 수 있게), 두 번 빼도 터지지 않는다.
func TestUnregister_ClosesOnce(t *testing.T) {
	h := New()
	c := h.Register(1)

	h.Unregister(c)
	h.Unregister(c) // 두 번째는 아무 일도 없어야 한다 (close 두 번이면 panic)

	if _, ok := <-c.Send; ok {
		t.Error("Send 가 닫히지 않음")
	}
	if h.Online(1) {
		t.Error("빠졌는데 Online 이 true")
	}
}

// 꽉 찬 연결이 있어도 Publish 는 멈추지 않고, 그 연결만 끊긴다.
func TestPublish_SlowClientDropped(t *testing.T) {
	h := New()
	slow := h.Register(1)
	fast := h.Register(2)

	for range SendBuffer + 5 { // slow 는 안 읽는다
		if err := h.Publish([]int64{1, 2}, testEvent{Type: "x"}); err != nil {
			t.Fatal(err)
		}
		<-fast.Send // fast 는 매번 읽는다
	}

	n := 0
	for range slow.Send { // 닫혔으면 버퍼에 남은 것만 읽고 끝난다
		n++
	}
	if n != SendBuffer {
		t.Errorf("slow 가 받은 수 = %d, want %d (꽉 찬 뒤 끊김)", n, SendBuffer)
	}
	if !h.Online(2) {
		t.Error("fast 까지 끊김")
	}
}

// 여러 고루틴이 동시에 등록·발행·해제해도 터지지 않는다.
func TestConcurrent(t *testing.T) {
	h := New()
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := h.Register(int64(i % 5))
			_ = h.Publish([]int64{0, 1, 2, 3, 4}, testEvent{Type: "x"})
			h.Unregister(c)
		}()
	}
	wg.Wait()
}
