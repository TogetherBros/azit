package proto

// 실시간 이벤트 종류
const (
	EventMessageCreated      = "message.created"      // data: MessageResponse
	EventMessageUpdated      = "message.updated"      // data: MessageRef (고침, 지움, 맞장구)
	EventNotificationCreated = "notification.created" // data: NotificationResponse
	EventAzitChanged         = "azit.changed"         // data: AzitChange (방·정보·멤버·짤·맞장구 목록이 바뀜 → 그 목록을 다시 받는다)
	EventPresence            = "presence"             // data: Presence (누가 앱을 켰다/껐다)
)

type Event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type MessageRef struct {
	AzitID    int64 `json:"azit_id"`
	RoomID    int64 `json:"room_id"`
	MessageID int64 `json:"message_id"`
}

// AzitChange 의 What 값. 앱은 이걸 보고 어느 목록을 다시 받을지 고른다.
const (
	ChangedRooms   = "rooms"   // 방 만들기·이름·순서
	ChangedInfo    = "info"    // 아지트 이름·색·소개·배너
	ChangedMembers = "members" // 들어옴·나감·내보냄·닉네임·역할
	ChangedMemes   = "memes"   // 짤 올리기·빼기
	ChangedReacts  = "reacts"  // 맞장구 목록
)

// AzitChange 는 "이 아지트의 무엇이 바뀌었다" 신호. 내용은 안 보내고, 앱이 HTTP 로 다시 받는다.
type AzitChange struct {
	AzitID int64  `json:"azit_id"`
	What   string `json:"what"`
}

// Presence 는 누가 앱을 켰거나(첫 연결) 껐다(마지막 연결이 끊김). 같은 아지트 사람들에게만 간다.
type Presence struct {
	AccountID int64 `json:"account_id"`
	Online    bool  `json:"online"`
}
