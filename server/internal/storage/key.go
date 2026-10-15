package storage

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
)

// Kind 는 업로드 종류. 종류마다 받는 형식, 최대 크기, 저장 폴더가 다르다.
type Kind string

const (
	KindAvatar Kind = "avatar"
	KindBanner Kind = "banner"
	KindMeme   Kind = "meme"

	KindAzitBanner Kind = "azit_banner" // 아지트 배너. 내 프로필 배너(KindBanner)와 폴더가 다르다
)

var (
	ErrUnknownKind = errors.New("알 수 없는 업로드 종류예요")
	ErrBadType     = errors.New("올릴 수 없는 파일 형식이에요")
	ErrTooLarge    = errors.New("파일이 너무 크거나 비어 있어요")
)

// rule 은 종류 하나의 접수 규칙.
type rule struct {
	prefix  string            // R2 안의 폴더 이름
	maxSize int64             // 최대 바이트
	exts    map[string]string // 받는 Content-Type → 확장자
}

var imageExts = map[string]string{
	"image/png":  "png",
	"image/jpeg": "jpg",
	"image/webp": "webp",
}

var memeExts = map[string]string{
	"image/png":  "png",
	"image/jpeg": "jpg",
	"image/webp": "webp",
	"image/gif":  "gif",
}

var rules = map[Kind]rule{
	KindAvatar: {prefix: "avatars", maxSize: 5 << 20, exts: imageExts},
	KindBanner: {prefix: "banners", maxSize: 5 << 20, exts: imageExts},
	KindMeme:   {prefix: "memes", maxSize: 10 << 20, exts: memeExts},

	KindAzitBanner: {prefix: "azit-banners", maxSize: 5 << 20, exts: imageExts},
}

// NewKey 는 업로드 요청을 검사하고, 통과하면 R2 저장 경로를 만든다.
// 예: avatars/12/ABCDEFGHIJKLMNOPQRSTUVWXYZ.webp
func NewKey(kind Kind, accountID int64, contentType string, size int64) (string, error) {
	r, ok := rules[kind]
	if !ok {
		return "", ErrUnknownKind
	}

	ext, ok := r.exts[contentType]
	if !ok {
		return "", ErrBadType
	}

	if size <= 0 || size > r.maxSize {
		return "", ErrTooLarge
	}

	return fmt.Sprintf("%s/%d/%s.%s", r.prefix, accountID, rand.Text(), ext), nil
}

// OwnedBy 는 key 가 그 계정의 해당 종류 폴더 안에 있는지 본다.
func OwnedBy(key string, kind Kind, accountID int64) bool {
	r, ok := rules[kind]
	if !ok {
		return false
	}
	return strings.HasPrefix(key, fmt.Sprintf("%s/%d/", r.prefix, accountID))
}
