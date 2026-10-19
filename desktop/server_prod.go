//go:build production

package main

// 운영 빌드(wails3 package, 릴리스 워크플로): 배포된 서버에 붙는다.
const serverURL = "https://api.togetherbros.uk"

// 최신 버전 정보. 릴리스 워크플로가 EC2 에 올리고 Caddy 가 /download/ 로 내려준다.
const updateURL = "https://api.togetherbros.uk/download/latest.json"
