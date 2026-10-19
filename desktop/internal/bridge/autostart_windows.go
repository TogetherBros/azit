//go:build windows

package bridge

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

// 사용자 로그인 때 실행되는 목록 (관리자 권한 필요 없음)
const (
	runKey   = `Software\Microsoft\Windows\CurrentVersion\Run`
	runValue = "azit"
)

// AutostartFlag 는 자동 실행으로 켜졌다는 표시. main 이 보고 창을 숨긴 채 트레이로 시작한다.
const AutostartFlag = "--autostart"

const autostartSupported = true

func autostartEnabled() (bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		return false, fmt.Errorf("자동 실행 설정을 읽지 못했어요: %w", err)
	}
	defer k.Close()
	_, _, err = k.GetStringValue(runValue)
	if errors.Is(err, registry.ErrNotExist) {
		return false, nil
	}
	return err == nil, err
}

func setAutostart(on bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("자동 실행 설정을 바꾸지 못했어요: %w", err)
	}
	defer k.Close()

	if !on {
		if err := k.DeleteValue(runValue); err != nil && !errors.Is(err, registry.ErrNotExist) {
			return fmt.Errorf("자동 실행을 끄지 못했어요: %w", err)
		}
		return nil
	}
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("실행 파일 위치를 찾지 못했어요: %w", err)
	}
	if err := k.SetStringValue(runValue, fmt.Sprintf(`"%s" %s`, exe, AutostartFlag)); err != nil {
		return fmt.Errorf("자동 실행을 켜지 못했어요: %w", err)
	}
	return nil
}
