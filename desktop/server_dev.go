//go:build !production

package main

// 개발 빌드(wails3 dev, wails3 build DEV=true): 내 PC 에서 띄운 서버에 붙는다.
// Wails 는 운영 빌드에만 -tags production 을 붙이기 때문에, 이 파일과 server_prod.go 중 하나만 빌드된다.
// (Spring 의 application-dev / application-prod 프로필과 같은 역할)
const serverURL = "http://127.0.0.1:8081"

// 개발 빌드에서는 업데이트 확인을 하지 않는다.
const updateURL = ""
