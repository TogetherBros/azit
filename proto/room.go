package proto

// CreateRoomRequest 는 POST /v1/azits/{id}/rooms 의 요청이다.
// 아지트와 만든 사람은 경로와 토큰에서 정하므로 보내지 않는다.
type CreateRoomRequest struct {
	Name string `json:"name"`
}

// UpdateRoomRequest 는 PATCH /v1/azits/{id}/rooms/{roomId} 의 요청이다.
type UpdateRoomRequest struct {
	Name string `json:"name"`
}

// RoomResponse 는 방 한 개. 목록, 만들기, 이름 바꾸기 응답에 같이 쓴다.
type RoomResponse struct {
	RoomID   int64  `json:"room_id"`
	AzitID   int64  `json:"azit_id"`
	Name     string `json:"name"`
	Position int    `json:"position"` // 작을수록 위
	Unread   int64  `json:"unread"`   // 내가 안 읽은 메시지 수 (내가 보낸 것, 지운 것은 빼고)
	// IsDefault 는 아지트를 만들 때 같이 생긴 기본 방 (처음 이름 "아지트"). 이름은 바꿀 수 있어도 지울 수는 없다.
	IsDefault bool `json:"is_default"`
}

// ReorderRoomsRequest 는 PUT /v1/azits/{id}/rooms/order 의 요청이다.
// 그 아지트 방 전부를 새 순서대로 담는다.
type ReorderRoomsRequest struct {
	RoomIDs []int64 `json:"room_ids"`
}
