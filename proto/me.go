package proto

// MeResponse 는 GET /v1/me 의 응답. 로그인한 내 계정 프로필.
type MeResponse struct {
	AccountID int64   `json:"account_id"`
	Name      *string `json:"name"` // @ 뒤 이름. 프로필을 아직 안 만들었으면 null
	Nickname  string  `json:"nickname"`
	Bio       string  `json:"bio"`
	Color     string  `json:"color"`
	AvatarURL string  `json:"avatar_url"` // 24시간짜리 보기 주소. 사진이 없으면 ""
	BannerURL string  `json:"banner_url"`

	// NeedsProfile 이 true 면 앱은 프로필 만들기 화면부터 보여준다.
	NeedsProfile bool `json:"needs_profile"`
}

// UpdateMeRequest 는 PATCH /v1/me 의 요청. 안 바뀐 칸도 네 칸 다 보낸다.
type UpdateMeRequest struct {
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
	Bio      string `json:"bio"`
	Color    string `json:"color"`
}

// NameCheckResponse 는 GET /v1/me/name-check 의 응답.
type NameCheckResponse struct {
	Available bool   `json:"available"`
	Message   string `json:"message,omitempty"` // 못 쓰는 이유. 쓸 수 있으면 빠진다
}
