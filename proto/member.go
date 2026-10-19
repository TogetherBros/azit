package proto

import "time"

// MemberResponse 는 아지트 멤버 한 명. GET /v1/azits/{id}/members 목록과 멤버를 바꾸는 API 응답에 같이 쓴다.
type MemberResponse struct {
	AccountID int64     `json:"account_id"`
	Name      string    `json:"name"`       // 계정 이름 (@ 뒤에 붙는 고유 이름)
	Nickname  string    `json:"nickname"`   // 이 아지트 안 닉네임
	Color     string    `json:"color"`      // 계정 색
	AvatarURL string    `json:"avatar_url"` // R2 보기 링크. 없으면 ""
	BannerURL string    `json:"banner_url"` // 프로필 배너 R2 보기 링크. 없으면 ""
	Bio       string    `json:"bio"`
	Role      string    `json:"role"` // super_admin, admin, member
	JoinedAt  time.Time `json:"joined_at"`
	Online    bool      `json:"online"` // 지금 앱을 켜 두었는지 (실시간 연결이 있는지)
}

// BanResponse 는 막은 사람 한 명. GET /v1/azits/{id}/bans
type BanResponse struct {
	AccountID     int64     `json:"account_id"`
	Name          string    `json:"name"`     // 계정 이름 (@)
	Nickname      string    `json:"nickname"` // 막을 때의 아지트 닉네임
	BannedByNick  string    `json:"banned_by_nickname"`
	BannedAt      time.Time `json:"banned_at"`
}

// KickRequest 는 DELETE /v1/azits/{id}/members/{accountId} 의 요청이다. 본문이 없으면 ban=false.
type KickRequest struct {
	Ban bool `json:"ban"` // true 면 막기: 새 초대 코드로도 다시 못 들어온다
}

// UpdateMyMemberRequest 는 PATCH /v1/azits/{id}/members/me 의 요청이다. 이 아지트 안 내 닉네임을 바꾼다.
type UpdateMyMemberRequest struct {
	Nickname string `json:"nickname"`
}

// UpdateMemberRoleRequest 는 PATCH /v1/azits/{id}/members/{accountId}/role 의 요청이다.
// 줄 수 있는 역할은 admin, member 둘뿐이다 (만든 사람 super_admin 은 넘기지 못한다).
type UpdateMemberRoleRequest struct {
	Role string `json:"role"`
}
