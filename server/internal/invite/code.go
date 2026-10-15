package invite

import (
	"crypto/rand"
	"strings"
)

// codeAlphabet 은 헷갈리는 글자(0, O, 1, I)를 뺀 32자. 256 이 32 로 나눠떨어져서 치우침이 없다.
const codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

const codeLen = 8

// NewCode 는 8자리 초대 코드를 만든다. 예: XK7P2Q9M
func NewCode() (string, error) {
	random := make([]byte, codeLen)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}

	code := make([]byte, codeLen)
	for i := range code {
		code[i] = codeAlphabet[int(random[i])%len(codeAlphabet)]
	}
	return string(code), nil
}

// NormalizeCode 는 사람이 친 코드를 저장된 모양으로 맞춘다. "xk7p-2q9m " -> "XK7P2Q9M"
func NormalizeCode(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	return strings.NewReplacer("-", "", " ", "").Replace(s)
}

// FormatCode 는 보여줄 때 4-4 로 끊는다. "XK7P2Q9M" -> "XK7P-2Q9M"
func FormatCode(code string) string {
	if len(code) != codeLen {
		return code
	}
	return code[:4] + "-" + code[4:]
}
