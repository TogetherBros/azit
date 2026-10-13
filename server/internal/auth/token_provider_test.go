package auth

import (
	"testing"
	"time"
)

func TestParseAccessToken(t *testing.T) {
	// 1. 준비: 토큰을 하나 만든다
	p := NewTokenProvider("test-secret", time.Minute, time.Hour)
	token, err := p.IssueAccessToken(14)
	if err != nil {
		t.Fatalf("IssueAccessToken 실패: %v", err)
	}

	// 2. 실행: 테스트할 함수를 부른다
	got, err := p.ParseAccessToken(token)

	// 3. 확인: 결과가 기대한 값인지 본다
	if err != nil {
		t.Fatalf("ParseAccessToken 실패: %v", err)
	}
	if got != 14 {
		t.Errorf("got %d, want 14", got)
	}
}

func TestParseAccessToken_WrongSecret(t *testing.T) {
	// 1. 준비: "test-secret" 으로 토큰을 만든다
	p := NewTokenProvider("test-secret", time.Minute, time.Hour)
	token, err := p.IssueAccessToken(14)
	if err != nil {
		t.Fatalf("IssueAccessToken 실패: %v", err)
	}

	// 2. 실행: 다른 비밀키를 가진 provider 로 연다
	other := NewTokenProvider("other-secret", time.Minute, time.Hour)
	_, err = other.ParseAccessToken(token)

	// 3. 확인: 에러가 "나야" 정상
	if err == nil {
		t.Errorf("틀린 비밀키인데 에러가 안 났다")
	}
}

func TestParseAccessToken_Expired(t *testing.T) {
	// 1. 준비: AT 유효 기간을 음수로 줘서 "1분 전에 이미 만료된" 토큰을 만든다
	p := NewTokenProvider("test-secret", -time.Minute, time.Hour)
	token, err := p.IssueAccessToken(14)
	if err != nil {
		t.Fatalf("IssueAccessToken 실패: %v", err)
	}

	// 2. 실행
	_, err = p.ParseAccessToken(token)

	// 3. 확인: 만료됐으니 에러가 나야 정상
	if err == nil {
		t.Errorf("만료된 토큰인데 에러가 안 났다")
	}
}

// 만들 때마다 다른 값이 나와야 한다
func TestNewRefreshToken_Unique(t *testing.T) {
	p := NewTokenProvider("test-secret", time.Minute, time.Hour)
	a, _, _ := p.NewRefreshToken()
	b, _, _ := p.NewRefreshToken()

	if a == b {
		t.Errorf("두 번 만들었는데 같은 값이 나왔다: %s", a)
	}
}

// 해시는 원문을 HashRefreshToken 한 값과 같고, 원문과는 달라야 한다
func TestNewRefreshToken_Hash(t *testing.T) {
	p := NewTokenProvider("test-secret", time.Minute, time.Hour)
	plain, hash, _ := p.NewRefreshToken()

	if hash != p.HashRefreshToken(plain) {
		t.Errorf("hash 가 HashRefreshToken(plain) 과 다르다")
	}
	if hash == plain {
		t.Errorf("hash 가 원문이랑 같다 (해시가 안 됐다)")
	}
}

// 만료 시각은 지금 + refreshTTL 이어야 한다
func TestNewRefreshToken_ExpiresAt(t *testing.T) {
	p := NewTokenProvider("test-secret", time.Minute, time.Hour)
	_, _, expiresAt := p.NewRefreshToken()

	want := time.Now().Add(time.Hour)
	if diff := want.Sub(expiresAt); diff < 0 || diff > time.Second {
		t.Errorf("expiresAt = %v, want 약 %v", expiresAt, want)
	}
}
