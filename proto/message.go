package proto

import "time"

// SendMessageRequest 는 POST /v1/azits/{id}/rooms/{roomId}/messages 의 요청이다.
// 보낸 사람은 토큰에서 정하므로 보내지 않는다.
type SendMessageRequest struct {
	TempID    string  `json:"temp_id"`     // 앱이 메시지마다 한 번 만든 UUID. 재전송해도 같은 값이라 중복 저장을 막는다
	Body      string  `json:"body"`        // 1~2000자
	ReplyToID *int64  `json:"reply_to_id"` // 답장이면 원래 메시지 id. 아니면 생략
	ShakeIDs  []int64 `json:"shake_ids"`   // @로 부른 사람들의 account id. 앱이 @칩을 고를 때 모은다. 없으면 생략
	ShakeAll  bool    `json:"shake_all"`   // @전체 를 썼는지
}

// MarkReadRequest 는 POST /v1/azits/{id}/rooms/{roomId}/read 의 요청이다.
// 이 방에서 여기까지 읽었다고 알린다. 저장된 값보다 작으면 무시한다.
type MarkReadRequest struct {
	MessageID int64 `json:"message_id"`
}

// MessageResponse 는 메시지 한 개.
// 보내기 응답과 MessageListResponse 안에 같이 쓴다.
type MessageResponse struct {
	MessageID int64              `json:"message_id"`
	RoomID    int64              `json:"room_id"`
	TempID    string             `json:"temp_id"`     // 앱이 "보내는 중" 말풍선을 이 값으로 찾아 바꾼다
	Kind      string             `json:"kind"`        // text, photo, meme, system
	Body      string             `json:"body"`        // 지운 메시지면 ""
	ReplyToID *int64             `json:"reply_to_id"` // 답장 아니면 null
	ShakeIDs  []int64            `json:"shake_ids"`   // @로 불린 사람들. 없으면 [] (내가 들어있으면 앱이 흔든다)
	ShakeAll  bool               `json:"shake_all"`   // @전체 였는지
	CreatedAt time.Time          `json:"created_at"`
	EditedAt  *time.Time         `json:"edited_at"` // 고친 적 없으면 null
	Deleted   bool               `json:"deleted"`
	Sender    *SenderResponse    `json:"sender"`    // system 메시지면 null
	Reactions []ReactionResponse `json:"reactions"` // 달린 맞장구를 종류별로 묶은 것. 처음 달린 순서. 없으면 []
}

// ReactionResponse 는 MessageResponse 안의 맞장구 한 종류. 같은 맞장구는 하나로 묶고 개수를 센다 ("ㅋㅋ 2").
// 글자 맞장구면 Text, 짤 맞장구면 MemeID + MemeURL 이 채워진다.
type ReactionResponse struct {
	Text    string `json:"text"`     // 글자면 "ㅋㅋ", 짤이면 ""
	MemeID  int64  `json:"meme_id"`  // 짤이면 짤 id, 글자면 0
	MemeURL string `json:"meme_url"` // 짤 보기 링크. 글자면 ""
	Count   int    `json:"count"`    // 단 사람 수
	Mine    bool   `json:"mine"`     // 내가 달았는지. 앱이 버튼을 눌린 상태로 그린다
}

// UpdateMessageRequest 는 PATCH /v1/azits/{id}/messages/{messageId} 의 요청이다. 본인 글자 메시지만 고칠 수 있다.
type UpdateMessageRequest struct {
	Body string `json:"body"` // 1~2000자
}

// ReactionRequest 는 PUT / DELETE /v1/azits/{id}/messages/{messageId}/reactions 의 요청이다.
// 글자(Text) 또는 짤(MemeID) 중 하나만 채운다. 응답은 그 메시지의 []ReactionResponse.
type ReactionRequest struct {
	Text   string `json:"text"`    // 1~4자. 짤이면 ""
	MemeID int64  `json:"meme_id"` // 아지트 짤 id. 글자면 0
}

// SenderResponse 는 MessageResponse 안의 보낸 사람.
// 나간 사람이어도 채운다.
type SenderResponse struct {
	AccountID int64  `json:"account_id"`
	Nickname  string `json:"nickname"` // 계정 닉네임이 아니라 그 아지트 안 닉네임
	Color     string `json:"color"`
	AvatarURL string `json:"avatar_url"` // R2 보기 링크. 없으면 ""
}

// MessageListResponse 는 GET /v1/azits/{id}/rooms/{roomId}/messages?before=&limit= 의 응답이다.
type MessageListResponse struct {
	Messages   []MessageResponse `json:"messages"`     // 오래된 → 최신 순서
	HasMore    bool              `json:"has_more"`     // 더 오래된 메시지가 남았는지
	LastReadID int64             `json:"last_read_id"` // 내가 어디까지 읽었는지. "여기까지 읽었음" 선 위치. 기록 없으면 0
}
