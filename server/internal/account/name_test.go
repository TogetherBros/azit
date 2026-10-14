package account

import (
	"errors"
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name string
		ok   bool
	}{
		{"abcde", true},
		{"4n5rud", true},
		{"wjd_ans_092208", true},
		{"abcdefghij0123456789", true},   // 20자
		{"abcd", false},                  // 4자
		{"abcdefghij01234567890", false}, // 21자
		{"Abcde", false},                 // 대문자
		{"ab.cde", false},                // 마침표
		{"ab cde", false},                // 공백
		{"문경문경문경", false},                // 한글
		{"", false},
	}
	for _, tt := range tests {
		err := ValidateName(tt.name)
		if tt.ok && err != nil {
			t.Errorf("%q: 통과해야 하는데 %v", tt.name, err)
		}
		if !tt.ok && !errors.Is(err, ErrInvalidName) {
			t.Errorf("%q: ErrInvalidName 이어야 하는데 %v", tt.name, err)
		}
	}
}

func TestNameBase(t *testing.T) {
	tests := []struct{ email, want string }{
		{"wjdansrud092208@gmail.com", "wjdansrud092208"},
		{"Kim.Min-su@gmail.com", "kimminsu"},
		{"kim@x.com", "kim_azit"},
		{"..@x.com", "azit_user"},
		{"문경@x.com", "azit_user"},
		{"abcdefghijklmnopqrstu@x.com", "abcdefghijklmnop"}, // 16자로 자름
	}
	for _, tt := range tests {
		got := NameBase(tt.email)
		if got != tt.want {
			t.Errorf("%q: got %q, want %q", tt.email, got, tt.want)
		}
		if err := ValidateName(got); err != nil {
			t.Errorf("%q: 추천값 %q 가 규칙에 안 맞음", tt.email, got)
		}
	}
}
