package account

import (
	"errors"
	"regexp"
	"strings"
)

var ErrInvalidName = errors.New("영문 소문자, 숫자, 밑줄(_)만 쓸 수 있어요. 5~20자 사이여야 해요")

var namePattern = regexp.MustCompile(`^[a-z0-9_]{5,20}$`)

func ValidateName(h string) error {
	if !namePattern.MatchString(h) {
		return ErrInvalidName
	}
	return nil
}

var notNameChar = regexp.MustCompile(`[^a-z0-9_]`)

// NameBase 는 이메일 앞부분으로 이름 후보를 만든다.
// 이미 누가 쓰는지는 여기서 안 본다 (2-3 에서 DB 를 보고 숫자를 붙인다).
func NameBase(email string) string {
	local, _, _ := strings.Cut(email, "@")
	h := notNameChar.ReplaceAllString(strings.ToLower(local), "")
	if len(h) > 16 {
		h = h[:16]
	}
	if h == "" {
		return "azit_user"
	}
	if len(h) < 5 {
		h += "_azit"
	}
	return h
}
