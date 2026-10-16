package proto

import "time"

// NotificationResponse 는 알림 한 줄. 문구는 앱이 kind 와 이름들로 만든다.
type NotificationResponse struct {
	NotificationID int64     `json:"notification_id"`
	Kind           string    `json:"kind"` // shake, join, kick
	AzitID         int64     `json:"azit_id"`
	AzitName       string    `json:"azit_name"`
	RoomID         int64     `json:"room_id"`   // shake 면 그 방, 아니면 0
	RoomName       string    `json:"room_name"` // shake 면 그 방 이름, 아니면 ""
	MessageID      int64     `json:"message_id"`
	ActorID        int64     `json:"actor_id"`
	ActorNickname  string    `json:"actor_nickname"` // 그때의 아지트 닉네임
	Preview        string    `json:"preview"`        // shake 면 메시지 앞부분 (40자)
	Read           bool      `json:"read"`
	CreatedAt      time.Time `json:"created_at"`
}

// NotificationListResponse 는 GET /v1/notifications 의 응답이다.
type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"` // 최신 순 50개
	Unread        int64                  `json:"unread"`        // 안 읽은 알림 수 (목록 밖 것도 센다)
}
