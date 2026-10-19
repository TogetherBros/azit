//go:build !windows

package bridge

import "errors"

// AutostartFlag 는 자동 실행으로 켜졌다는 표시 (Windows 만 쓴다).
const AutostartFlag = "--autostart"

const autostartSupported = false

// macOS(로그인 항목)·Linux(autostart .desktop)는 아직 없다.
func autostartEnabled() (bool, error) { return false, nil }

func setAutostart(bool) error {
	return errors.New("이 운영체제에서는 아직 자동 실행을 못 켜요")
}
