package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	content := "# 주석\n\nAZIT_A=hello\nAZIT_B = \"quoted\"\nAZIT_C=a=b\r\nAZIT_KEEP=fromfile\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	// t.Setenv 는 테스트가 끝나면 원래대로 돌려놓는다
	t.Setenv("AZIT_KEEP", "already")
	for _, k := range []string{"AZIT_A", "AZIT_B", "AZIT_C"} {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}

	if err := loadDotEnv(path); err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"AZIT_A":    "hello",
		"AZIT_B":    "quoted",
		"AZIT_C":    "a=b",
		"AZIT_KEEP": "already", // 이미 있던 값은 안 덮어씀
	}
	for k, v := range want {
		if got := os.Getenv(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
}

func TestLoadDotEnvMissingFile(t *testing.T) {
	if err := loadDotEnv(filepath.Join(t.TempDir(), "없음.env")); err != nil {
		t.Errorf("파일이 없으면 nil 이어야 해요: %v", err)
	}
}
