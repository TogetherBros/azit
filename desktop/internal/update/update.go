// Package update 는 서버에 올라간 최신 앱 버전을 확인한다.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// latest 는 릴리스 워크플로가 만드는 latest.json 의 모양.
//
//	{"version":"0.2.0","notes":"...","files":{"windows":"https://...","darwin":"...","linux":"..."}}
type latest struct {
	Version string            `json:"version"`
	Notes   string            `json:"notes"`
	Files   map[string]string `json:"files"` // 키는 runtime.GOOS
}

// Info 는 React 에 넘기는 확인 결과.
type Info struct {
	Current   string `json:"current"`
	Latest    string `json:"latest"`
	Available bool   `json:"available"` // 새 버전이 있고, 이 OS 용 파일도 있음
	URL       string `json:"url"`       // 이 OS 용 설치 파일 주소
	Notes     string `json:"notes"`
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

// Check 는 url 의 latest.json 을 받아 current 보다 새 버전이 있는지 본다.
// url 이 비었거나(개발 빌드) current 가 "dev" 면 확인하지 않고 Available=false 를 돌려준다.
func Check(ctx context.Context, current, url string) (Info, error) {
	info := Info{Current: current}
	if url == "" || current == "dev" {
		return info, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return info, err
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return info, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return info, fmt.Errorf("업데이트 정보 응답 %d", res.StatusCode)
	}

	var l latest
	if err := json.NewDecoder(res.Body).Decode(&l); err != nil {
		return info, err
	}

	info.Latest = l.Version
	info.Notes = l.Notes
	info.URL = l.Files[runtime.GOOS]
	info.Available = info.URL != "" && newer(l.Version, current)
	return info, nil
}

// newer 는 a 가 b 보다 높은 버전인지 본다. "v1.2.3" 과 "1.2.3" 둘 다 받는다.
func newer(a, b string) bool {
	pa, pb := parts(a), parts(b)
	for i := range 3 {
		if pa[i] != pb[i] {
			return pa[i] > pb[i]
		}
	}
	return false
}

func parts(v string) [3]int {
	var out [3]int
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	v, _, _ = strings.Cut(v, "-") // 1.2.3-beta 같은 꼬리는 무시
	for i, p := range strings.SplitN(v, ".", 3) {
		out[i], _ = strconv.Atoi(p)
	}
	return out
}
