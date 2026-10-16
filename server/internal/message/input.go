package message

import (
	"errors"
	"slices"
	"strings"
	"unicode/utf8"
)

// 글자 수 제한. 글자 수는 바이트가 아니라 글자(rune) 기준이다 ("가" 는 3바이트지만 1글자).
const (
	MaxBodyLen   = 2000
	MaxTempIDLen = 64
)

var (
	ErrBodyEmpty   = errors.New("메시지를 입력해 주세요")
	ErrBodyLong    = errors.New("메시지는 2000자까지 보낼 수 있어요")
	ErrTempIDEmpty = errors.New("temp_id 가 없어요")
	ErrTempIDLong  = errors.New("temp_id 가 너무 길어요")
)

// ValidateBody 는 본문을 검사한다. 앞뒤 공백만 있는 메시지는 빈 메시지로 본다.
// 본문 자체는 고치지 않는다 (줄바꿈, 들여쓰기도 메시지의 일부라서).
func ValidateBody(body string) error {
	if strings.TrimSpace(body) == "" {
		return ErrBodyEmpty
	}
	if utf8.RuneCountInString(body) > MaxBodyLen {
		return ErrBodyLong
	}
	return nil
}

// ValidateTempID 는 번호표를 검사한다. 앱이 만든 UUID 라 36자지만 넉넉히 64자까지 받는다.
func ValidateTempID(tempID string) error {
	if tempID == "" {
		return ErrTempIDEmpty
	}
	if len(tempID) > MaxTempIDLen {
		return ErrTempIDLong
	}
	return nil
}

// UniqueIDs 는 id 목록을 정렬하고 중복을 뺀 새 슬라이스를 돌려준다.
// 같은 사람을 두 번 @ 해도 Shake 는 한 줄이어야 해서 (두 줄이면 PK 충돌로 저장이 통째로 실패한다).
func UniqueIDs(ids []int64) []int64 {
	out := slices.Clone(ids) // 원본(요청 값)은 건드리지 않으려고 복사
	slices.Sort(out)
	return slices.Compact(out) // 정렬된 상태에서 붙어 있는 같은 값을 하나로 줄인다
}
