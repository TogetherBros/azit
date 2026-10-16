package proto

import "time"

// MemeResponse 는 아지트 짤 한 개. GET /v1/azits/{id}/memes 목록과 등록 응답에 쓴다.
type MemeResponse struct {
	MemeID    int64     `json:"meme_id"`
	Name      string    `json:"name"`       // "킹받네" 같은 이름
	URL       string    `json:"url"`        // R2 보기 링크
	CreatedBy int64     `json:"created_by"` // 올린 사람 account id
	CreatedAt time.Time `json:"created_at"`
}

// CreateMemeRequest 는 POST /v1/azits/{id}/memes 의 요청이다.
// 먼저 POST /v1/uploads (kind=meme) 로 올리고, 받은 key 를 여기에 넣는다.
type CreateMemeRequest struct {
	Name string `json:"name"` // 1~20자
	Key  string `json:"key"`  // 업로드로 받은 R2 키 (memes/{내 id}/...)
}

// ReactItem 은 아지트 맞장구 목록의 한 칸. 글자(Text) 또는 짤(MemeID) 중 하나만 채운다.
type ReactItem struct {
	Text   string `json:"text"`    // 1~2자. 짤이면 ""
	MemeID int64  `json:"meme_id"` // 아지트 짤 id. 글자면 0
}

// ReplaceReactsRequest 는 PUT /v1/azits/{id}/reacts 의 요청이다. 목록을 통째로 바꾼다 (최대 10개, 순서대로).
type ReplaceReactsRequest struct {
	Reacts []ReactItem `json:"reacts"`
}

// ReactResponse 는 아지트 맞장구 목록의 한 칸. GET / PUT /v1/azits/{id}/reacts 응답은 이것의 배열이다.
type ReactResponse struct {
	ReactID  int64  `json:"react_id"`
	Text     string `json:"text"`     // 글자면 "ㅋㅋ", 짤이면 ""
	MemeID   int64  `json:"meme_id"`  // 짤이면 짤 id, 글자면 0
	MemeURL  string `json:"meme_url"` // 짤 보기 링크. 글자면 ""
	Position int    `json:"position"` // 작을수록 앞
}
