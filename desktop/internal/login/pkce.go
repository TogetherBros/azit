package login

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// randomString 은 n 바이트 랜덤값을 URL 에 넣어도 되는 문자열로 만든다 (패딩 없음).
func randomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// newPKCE 는 verifier(앱만 아는 비밀)와 challenge(그 해시)를 만든다.
// 구글에는 challenge 를, 서버에는 나중에 verifier 를 보낸다.
func newPKCE() (verifier, challenge string, err error) {
	verifier, err = randomString(32)
	if err != nil {
		return "", "", err
	}
	return verifier, challengeOf(verifier), nil
}

// challengeOf 는 verifier 를 SHA-256 해서 base64 로 바꾼다 (S256 방식).
func challengeOf(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
