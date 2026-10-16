package meme

import (
	"errors"
	"strings"
	"unicode/utf8"
)

const (
	MaxNameLen   = 20 // 짤 이름 글자 수
	MaxReactText = 4  // 글자 맞장구 글자 수 (1~4자)
	MaxReacts    = 10 // 아지트 맞장구 목록 칸 수
)

var (
	ErrNameEmpty      = errors.New("짤 이름을 입력해 주세요")
	ErrNameLong       = errors.New("짤 이름은 20자까지예요")
	ErrReactEmpty     = errors.New("글자나 짤 중 하나를 골라 주세요")
	ErrReactBoth      = errors.New("글자와 짤은 하나만 고를 수 있어요")
	ErrReactTextLong  = errors.New("맞장구 글자는 4자까지예요")
	ErrTooManyReacts  = errors.New("맞장구 목록은 10개까지예요")
	ErrDuplicateReact = errors.New("같은 맞장구가 두 번 들어 있어요")
)

// NormalizeName 은 짤 이름 앞뒤 공백을 지운다.
func NormalizeName(name string) string {
	return strings.TrimSpace(name)
}

// ValidateName 은 짤 이름을 검사한다. NormalizeName 한 값을 넣는다.
func ValidateName(name string) error {
	if name == "" {
		return ErrNameEmpty
	}
	if utf8.RuneCountInString(name) > MaxNameLen {
		return ErrNameLong
	}
	return nil
}

// NormalizeReactText 는 맞장구 글자 앞뒤 공백을 지운다.
func NormalizeReactText(text string) string {
	return strings.TrimSpace(text)
}

// ValidateReact 는 맞장구 하나를 검사한다 (글자 또는 짤 중 정확히 하나).
// 메시지에 다는 맞장구와 아지트 맞장구 목록 칸에 같이 쓴다. text 는 NormalizeReactText 한 값.
// 짤이 이 아지트 것인지는 DB 를 봐야 해서 여기서는 안 본다.
func ValidateReact(text string, memeID int64) error {
	switch {
	case text == "" && memeID <= 0:
		return ErrReactEmpty
	case text != "" && memeID > 0:
		return ErrReactBoth
	case utf8.RuneCountInString(text) > MaxReactText:
		return ErrReactTextLong
	}
	return nil
}
