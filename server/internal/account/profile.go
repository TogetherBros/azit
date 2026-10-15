package account

import (
	"errors"
	"slices"
	"strings"
	"unicode/utf8"
)

var (
	ErrNicknameEmpty = errors.New("닉네임을 넣어 주세요")
	ErrNicknameLong  = errors.New("닉네임은 16자까지예요")
	ErrBioLong       = errors.New("소개는 190자까지예요")
	ErrBadColor      = errors.New("고를 수 있는 색이 아니에요")
)

// PickColors 는 앱이 보여 주는 색 목록. desktop/frontend/src/lib/color.ts 의 PICK_COLORS 와 같아야 한다.
var PickColors = []string{"#D4FF3A", "#FF5FA8", "#7BC4FF", "#FF8A3D", "#B79CFF", "#F4F4F6"}

// Normalize 는 앞뒤 공백을 지우고 색을 대문자로 맞춘다. Validate 전에 부른다.
func (p *ProfileUpdate) Normalize() {
	p.Name = strings.TrimSpace(p.Name)
	p.Nickname = strings.TrimSpace(p.Nickname)
	p.Bio = strings.TrimSpace(p.Bio)
	p.Color = strings.ToUpper(strings.TrimSpace(p.Color))
}

// Validate 는 저장해도 되는 값인지 본다. 이름이 겹치는지는 DB 가 본다.
func (p ProfileUpdate) Validate() error {
	if err := ValidateName(p.Name); err != nil {
		return err
	}

	n := utf8.RuneCountInString(p.Nickname)
	if n == 0 {
		return ErrNicknameEmpty
	}
	if n > 16 {
		return ErrNicknameLong
	}

	if utf8.RuneCountInString(p.Bio) > 190 {
		return ErrBioLong
	}

	if !slices.Contains(PickColors, p.Color) {
		return ErrBadColor
	}
	return nil
}
