//go:build production

package main

// 운영 빌드(wails3 package, 릴리스 워크플로): 배포된 서버에 붙는다.
const serverURL = "https://api.togetherbros.uk"

// 최신 버전 정보. 릴리스 워크플로가 GitHub Release 에 latest.json 을 같이 올린다.
// releases/latest/download/<파일> 은 항상 가장 최근 릴리스의 파일로 연결된다.
const updateURL = "https://github.com/TogetherBros/azit/releases/latest/download/latest.json"
