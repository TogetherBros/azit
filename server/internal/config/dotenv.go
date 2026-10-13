package config

import (
	"errors"
	"os"
	"strings"
)

// loadDotEnv 는 .env 파일의 KEY=값 을 환경 변수로 올린다.
// 파일이 없으면 그냥 넘어간다 (운영 서버는 환경 변수를 직접 넣는다).
// 이미 들어 있는 환경 변수는 덮어쓰지 않는다.
func loadDotEnv(path string) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = unquote(strings.TrimSpace(value))

		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	return nil
}

// unquote 는 값 양끝의 "..." 나 '...' 를 벗긴다.
func unquote(s string) string {
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') && s[len(s)-1] == s[0] {
		return s[1 : len(s)-1]
	}
	return s
}
