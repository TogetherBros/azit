package auth

import "testing"

// 만들 때마다 다른 값이 나와야 한다
func TestNewRefreshToken_Unique(t *testing.T) {
	a, _ := NewRefreshToken()
	b, _ := NewRefreshToken()

	if a == b {
		t.Errorf("두 번 만들었는데 같은 값이 나왔다: %s", a)
	}
}

// 해시는 원문을 HashRefreshToken 한 값과 같고, 원문과는 달라야 한다
func TestNewRefreshToken_Hash(t *testing.T) {
	plain, hash := NewRefreshToken()

	if hash != HashRefreshToken(plain) {
		t.Errorf("hash 가 HashRefreshToken(plain) 과 다르다")
	}
	if hash == plain {
		t.Errorf("hash 가 원문이랑 같다 (해시가 안 됐다)")
	}
}
