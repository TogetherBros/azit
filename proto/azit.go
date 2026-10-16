package proto

// CreateAzitRequest 는 POST /v1/azits 의 요청이다.
// MasterID 는 인증된 계정에서 정하므로 클라이언트가 보내지 않는다.
type CreateAzitRequest struct {
	Name      string `json:"name"`
	Color     string `json:"color"`
	BannerKey string `json:"banner_key"`
}

// AzitResponse 는 아지트 한 개. 만들기와 목록 응답에 같이 쓴다.
type AzitResponse struct {
	AzitID      int64  `json:"azit_id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	BannerURL   string `json:"banner_url"` // 없으면 ""
	MemberCount int64  `json:"member_count"`
	MyRole      string `json:"my_role"`     // 이 아지트에서 내 역할: super_admin, admin, member. 초대 미리보기처럼 아직 멤버가 아니면 ""
	MyNickname  string `json:"my_nickname"` // 이 아지트 안 내 닉네임. 멤버가 아니면 ""
}
