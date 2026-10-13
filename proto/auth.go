package proto

// RefreshRequest 는 POST /v1/auth/refresh 의 요청 본문.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// TokenResponse 는 로그인·갱신이 성공했을 때 돌려주는 토큰 한 쌍.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// GoogleLoginRequest 는 POST /v1/auth/google 의 요청 본문.
// 앱이 구글 로그인 후 받은 code 와, 처음에 만든 PKCE verifier,
// 콜백에 쓴 주소를 그대로 보낸다.
type GoogleLoginRequest struct {
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
	RedirectURI  string `json:"redirect_uri"`
}
