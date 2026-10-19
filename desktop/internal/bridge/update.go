package bridge

import (
	"context"

	"github.com/TogetherBros/azit/desktop/internal/update"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// UpdateService 는 새 앱 버전이 있는지 React 에 알려 주고, 받기 페이지를 브라우저로 연다.
type UpdateService struct {
	current string
	url     string
}

func NewUpdateService(current, url string) *UpdateService {
	return &UpdateService{current: current, url: url}
}

// Check 는 최신 버전을 확인한다. 새 버전이 없거나 개발 빌드면 Available=false.
func (s *UpdateService) Check(ctx context.Context) (update.Info, error) {
	return update.Check(ctx, s.current, s.url)
}

// Open 은 설치 파일 주소를 기본 브라우저로 연다. 받아서 설치하는 건 사용자가 한다.
func (s *UpdateService) Open(url string) error {
	return application.Get().Browser.OpenURL(url)
}
