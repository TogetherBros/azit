package room

import (
	"errors"
	"strings"
	"unicode/utf8"
)

const maxNameLen = 20

var (
	ErrNameEmpty = errors.New("방 이름을 넣어 주세요")
	ErrNameLong  = errors.New("방 이름은 20자까지예요")
)

// NormalizeName 은 앞뒤 공백을 자른다. 만들기와 이름 바꾸기에 같이 쓴다.
func NormalizeName(name string) string {
	return strings.TrimSpace(name)
}

// ValidateName 은 NormalizeName 뒤에 부른다. 글자 수는 바이트가 아니라 글자(rune)로 센다.
func ValidateName(name string) error {
	n := utf8.RuneCountInString(name)
	if n == 0 {
		return ErrNameEmpty
	}
	if n > maxNameLen {
		return ErrNameLong
	}
	return nil
}
