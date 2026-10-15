package account

import (
	"errors"
	"strings"
	"testing"
)

func TestProfileUpdateValidate(t *testing.T) {
	ok := ProfileUpdate{Name: "jmk_01", Nickname: "민경", Bio: "안녕", Color: "#D4FF3A"}

	tests := []struct {
		name   string
		change func(p *ProfileUpdate)
		want   error
	}{
		{"정상", func(p *ProfileUpdate) {}, nil},
		{"이름 규칙 위반", func(p *ProfileUpdate) { p.Name = "JMK" }, ErrInvalidName},
		{"닉네임 공백만", func(p *ProfileUpdate) { p.Nickname = "   " }, ErrNicknameEmpty},
		{"닉네임 한글 16자는 됨", func(p *ProfileUpdate) { p.Nickname = strings.Repeat("가", 16) }, nil},
		{"닉네임 17자", func(p *ProfileUpdate) { p.Nickname = strings.Repeat("가", 17) }, ErrNicknameLong},
		{"소개 191자", func(p *ProfileUpdate) { p.Bio = strings.Repeat("가", 191) }, ErrBioLong},
		{"색 소문자도 됨", func(p *ProfileUpdate) { p.Color = "#d4ff3a" }, nil},
		{"목록에 없는 색", func(p *ProfileUpdate) { p.Color = "#000000" }, ErrBadColor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ok
			tt.change(&p)
			p.Normalize()
			if err := p.Validate(); !errors.Is(err, tt.want) {
				t.Fatalf("got %v, want %v", err, tt.want)
			}
		})
	}
}
