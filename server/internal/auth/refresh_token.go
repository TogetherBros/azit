package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// NewRefreshToken 은 앱에 줄 원문과 DB 에 넣을 해시를 같이 만든다.
func NewRefreshToken() (plain string, hash string) {
	b := make([]byte, 32)
	rand.Read(b)

	plain = base64.RawURLEncoding.EncodeToString(b)
	return plain, HashRefreshToken(plain)
}

// HashRefreshToken 은 RT 원문을 DB 에 넣을 모양(sha256 → 16진수 글자)으로 바꾼다.
func HashRefreshToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
