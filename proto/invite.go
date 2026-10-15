package proto

import "time"

// CreateInviteRequest 는 POST /v1/azits/{id}/invites 의 요청이다.
type CreateInviteRequest struct {
	ExpiresIn string `json:"expires_in"` // "30m" "1h" "1d" "7d" "never". 비면 "7d"
	MaxUses   int    `json:"max_uses"`   // 0(무제한) 1 5 10
}

// InviteResponse 는 초대 코드 한 개. 만들기와 목록 응답에 쓴다.
type InviteResponse struct {
	Code      string     `json:"code"` // 보여주기용 XXXX-XXXX
	AzitID    int64      `json:"azit_id"`
	CreatedBy int64      `json:"created_by"`
	ExpiresAt *time.Time `json:"expires_at"` // null 이면 무제한
	MaxUses   int        `json:"max_uses"`   // 0 이면 무제한
	Uses      int        `json:"uses"`
	CreatedAt time.Time  `json:"created_at"`
}

// InvitePreviewResponse 는 GET /v1/invites/{code} 의 응답이다. 들어가기 전에 보여줄 정보.
type InvitePreviewResponse struct {
	Code            string       `json:"code"`
	Azit            AzitResponse `json:"azit"`
	InviterNickname string       `json:"inviter_nickname"` // 초대 만든 사람의 아지트 안 닉네임
	AlreadyMember   bool         `json:"already_member"`
}
