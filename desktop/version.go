package main

// appVersion 은 지금 앱의 버전. 업데이트 확인할 때 서버의 최신 버전과 비교한다.
// 손으로 바꾸지 않는다: 릴리스 워크플로가 git 태그(v0.2.0)를 보고 빌드 직전에 이 값을 바꿔 넣는다.
// 로컬 빌드는 "dev" 라서 업데이트 알림이 뜨지 않는다.
const appVersion = "dev"
