package azit

import (
	"errors"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/TogetherBros/azit/server/internal/account"
)

var (
	ErrNameEmpty = errors.New("아지트 이름을 넣어 주세요")
	ErrNameLong  = errors.New("아지트 이름은 20자까지예요")
)

// Normalize 는 앞뒤 공백을 자르고 색을 대문자로 맞춘다. 색이 비면 기본 색.
func (in *CreateInput) Normalize() {
	in.Name = strings.TrimSpace(in.Name)
	in.Color = strings.ToUpper(strings.TrimSpace(in.Color))
	if in.Color == "" {
		in.Color = account.PickColors[0]
	}
}

// Validate 는 Normalize 뒤에 부른다.
func (in CreateInput) Validate() error {
	n := utf8.RuneCountInString(in.Name)
	if n == 0 {
		return ErrNameEmpty
	}
	if n > 20 {
		return ErrNameLong
	}

	if !slices.Contains(account.PickColors, in.Color) {
		return account.ErrBadColor
	}

	return nil
}
