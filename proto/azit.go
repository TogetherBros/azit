package proto

// CreateAzitRequest 는 POST /v1/azits 의 요청이다.
// MasterID 는 인증된 계정에서 정하므로 클라이언트가 보내지 않는다.
type CreateAzitRequest struct {
	Name      string `json:"name"`
	Color     string `json:"color"`
	BannerKey string `json:"banner_key"`
}

// UpdateAzitRequest 는 PATCH /v1/azits/{id} 의 요청이다. 바꿀 칸만 보낸다 (nil = 안 바꿈).
type UpdateAzitRequest struct {
	Name        *string `json:"name"`
	Color       *string `json:"color"`
	Description *string `json:"description"` // "" 이면 소개를 지운다
	BannerKey   *string `json:"banner_key"`  // POST /v1/uploads (kind=azit_banner) 로 올린 키. "" 이면 배너를 지운다
}

// AzitResponse 는 아지트 한 개. 만들기와 목록 응답에 같이 쓴다.
type AzitResponse struct {
	AzitID      int64  `json:"azit_id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"` // 한 줄 소개. 없으면 ""
	BannerURL   string `json:"banner_url"`  // 없으면 ""
	MemberCount int64  `json:"member_count"`
	MyRole      string `json:"my_role"`     // 이 아지트에서 내 역할: super_admin, admin, member. 초대 미리보기처럼 아직 멤버가 아니면 ""
	MyNickname  string `json:"my_nickname"` // 이 아지트 안 내 닉네임. 멤버가 아니면 ""
}
