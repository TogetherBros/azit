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
}
