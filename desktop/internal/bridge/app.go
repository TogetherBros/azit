package bridge

import (
	"sync/atomic"
)

// AppService 는 앱 설정 중 운영체제에 닿는 것: 컴퓨터 켜면 자동 실행, 창 닫으면 트레이로 숨기기.
// 설정 값 자체는 React(localStorage)가 들고 있고, 앱이 켜질 때와 바뀔 때 여기로 알려 준다.
type AppService struct {
	closeToTray atomic.Bool
	quitting    atomic.Bool
}

func NewAppService() *AppService {
	return &AppService{}
}

// SetCloseToTray: 켜면 창의 X 가 앱을 끄지 않고 트레이로 숨긴다 (main.go 의 닫기 훅이 본다).
func (s *AppService) SetCloseToTray(on bool) {
	s.closeToTray.Store(on)
}

// HideOnClose 는 지금 창 닫기를 숨기기로 바꿔야 하는지. 트레이의 "끝내기"로 끄는 중이면 아니다.
// (메서드로 두면 React 쪽 바인딩에도 나가서, main 만 쓰는 건 함수로 둔다)
func HideOnClose(s *AppService) bool {
	return s.closeToTray.Load() && !s.quitting.Load()
}

// MarkQuitting 은 트레이 "끝내기" 직전에 부른다 (닫기 훅이 막지 않게).
func MarkQuitting(s *AppService) {
	s.quitting.Store(true)
}

// AutostartSupported: 이 운영체제에서 자동 실행을 켤 수 있는지 (지금은 Windows 만).
func (s *AppService) AutostartSupported() bool {
	return autostartSupported
}

// Autostart: 지금 자동 실행이 켜져 있는지 (레지스트리에서 읽는다).
func (s *AppService) Autostart() (bool, error) {
	return autostartEnabled()
}

// SetAutostart: 컴퓨터 켜면 자동 실행을 켜고 끈다. 켤 때는 지금 실행 파일 경로를 적는다.
func (s *AppService) SetAutostart(on bool) error {
	return setAutostart(on)
}
