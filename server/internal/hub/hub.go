package hub

import (
	"encoding/json"
	"sync"
)

const SendBuffer = 32

type Client struct {
	AccountID int64
	Send      chan []byte
}

type Hub struct {
	mu      sync.Mutex
	clients map[int64]map[*Client]struct{}
}

func New() *Hub {
	return &Hub{clients: map[int64]map[*Client]struct{}{}}
}

func (h *Hub) Register(accountID int64) *Client {
	c := &Client{
		AccountID: accountID,
		Send:      make(chan []byte, SendBuffer),
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[accountID] == nil {
		h.clients[accountID] = map[*Client]struct{}{}
	}
	h.clients[accountID][c] = struct{}{}
	return c
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.remove(c)
}

func (h *Hub) remove(c *Client) {
	conns, ok := h.clients[c.AccountID]
	if !ok {
		return
	}
	if _, ok := conns[c]; !ok {
		return
	}
	delete(conns, c)
	close(c.Send)
	if len(conns) == 0 {
		delete(h.clients, c.AccountID)
	}
}

// Publish 는 이벤트를 accountIDs 의 모든 연결에 넣는다. 절대 기다리지 않는다.
// 어떤 연결이 꽉 찼으면 그 연결을 끊고 다음으로 넘어간다.
func (h *Hub) Publish(accountIDs []int64, ev any) error {
	msg, err := json.Marshal(ev) // 한 번만 바꿔서 모두에게 같은 바이트를 보낸다
	if err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	for _, id := range accountIDs {
		for c := range h.clients[id] {
			select {
			case c.Send <- msg:
			default:
				h.remove(c) // 꽉 참 → 끊는다. 앱은 다시 붙고 목록을 새로 받는다
			}
		}
	}
	return nil
}

func (h *Hub) Online(accountID int64) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.clients[accountID]) > 0
}
