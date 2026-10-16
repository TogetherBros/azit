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
	ErrDescLong  = errors.New("소개는 40자까지예요")
)

const MaxDescLen = 40

// UpdateInput 은 아지트 기본 정보 바꾸기. nil 인 칸은 안 바꾼다.
type UpdateInput struct {
	Name        *string
	Color       *string
	Description *string
	BannerKey   *string
}

// Normalize 는 보낸 칸만 앞뒤 공백을 자르고 색을 대문자로 맞춘다.
func (in *UpdateInput) Normalize() {
	trim := func(p *string) {
		if p != nil {
			*p = strings.TrimSpace(*p)
		}
	}
	trim(in.Name)
	trim(in.Description)
	trim(in.BannerKey)
	if in.Color != nil {
		*in.Color = strings.ToUpper(strings.TrimSpace(*in.Color))
	}
}

// Validate 는 Normalize 뒤에 부른다. 배너 키 주인 확인은 handler 가 한다 (R2 를 봐야 해서).
func (in UpdateInput) Validate() error {
	if in.Name != nil {
		if err := (CreateInput{Name: *in.Name, Color: account.PickColors[0]}).Validate(); err != nil {
			return err
		}
	}
	if in.Color != nil && !slices.Contains(account.PickColors, *in.Color) {
		return account.ErrBadColor
	}
	if in.Description != nil && utf8.RuneCountInString(*in.Description) > MaxDescLen {
		return ErrDescLong
	}
	return nil
}

// Changes 는 바꿀 칸을 DB 칸 이름 → 값 으로 만든다. 비면 바꿀 게 없다.
func (in UpdateInput) Changes() map[string]any {
	out := map[string]any{}
	if in.Name != nil {
		out["name"] = *in.Name
	}
	if in.Color != nil {
		out["color"] = *in.Color
	}
	if in.Description != nil {
		out["description"] = *in.Description
	}
	if in.BannerKey != nil {
		out["banner_key"] = *in.BannerKey
	}
	return out
}

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
