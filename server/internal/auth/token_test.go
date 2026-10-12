package auth

import (
	"testing"
	"time"
)

func TestParseAT(t *testing.T) {
	// 1. 준비: 토큰을 하나 만든다
	token, err := IssueAT("test-secret", 14, time.Minute)
	if err != nil {
		t.Fatalf("IssueAT 실패: %v", err)
	}

	// 2. 실행: 테스트할 함수를 부른다
	got, err := ParseAT("test-secret", token)

	// 3. 확인: 결과가 기대한 값인지 본다
	if err != nil {
		t.Fatalf("ParseAT 실패: %v", err)
	}
	if got != 14 {
		t.Errorf("got %d, want 14", got)
	}
}

func TestParseAT_WrongSecret(t *testing.T) {
	// 1. 준비: "test-secret" 으로 토큰을 만든다
	token, err := IssueAT("test-secret", 14, time.Minute)
	if err != nil {
		t.Fatalf("IssueAT 실패: %v", err)
	}

	// 2. 실행: 다른 비밀키로 연다. 번호는 필요 없으니 _ 로 버린다
	_, err = ParseAT("other-secret", token)

	// 3. 확인: 에러가 "나야" 정상
	if err == nil {
		t.Errorf("틀린 비밀키인데 에러가 안 났다")
	}
}

func TestParseAT_Expired(t *testing.T) {
	// 1. 준비: 유효 기간을 음수로 줘서 "1분 전에 이미 만료된" 토큰을 만든다
	token, err := IssueAT("test-secret", 14, -time.Minute)
	if err != nil {
		t.Fatalf("IssueAT 실패: %v", err)
	}

	// 2. 실행
	_, err = ParseAT("test-secret", token)

	// 3. 확인: 만료됐으니 에러가 나야 정상
	if err == nil {
		t.Errorf("만료된 토큰인데 에러가 안 났다")
	}
}
