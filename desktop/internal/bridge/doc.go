// Package bridge 는 React 화면에 공개하는 Go 메서드를 모은다.
//
// Wails 는 여기 구조체를 application.NewService(...) 로 등록하면, 공개 메서드마다
// TypeScript 함수를 자동 생성해 준다 (wails3 generate bindings → frontend/bindings/).
// React 는 그 함수를 import 해서 await 로 부른다.
//
// Spring 으로 치면 Controller 다. 단, HTTP 대신 Wails 가 JS ↔ Go 호출을 이어 준다.
//   React → bridge.ChatService.Send(...)  → client 가 서버로 전송
//   서버 → client 가 수신 → app.Event.Emit("message", ...) → React 가 구독
//
// 넣는 것: 얇은 메서드. 실제 일은 client, login 에 맡긴다.
// 창 흔들기처럼 네이티브 창을 건드리는 일도 여기서 한다.
package bridge
