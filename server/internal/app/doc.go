// Package app 은 서버의 모든 부품을 조립한다.
//
// Spring 의 ApplicationContext 가 자동으로 해 주던 일(빈 생성, 의존성 주입)을
// 여기서 손으로 한다. DB 를 열고 → 각 도메인의 Service 를 만들고 →
// Handler 를 라우터에 등록하고 → http.Server 를 돌려준다.
//
// 넣는 것: 설정(Config) 구조체, New(cfg) 같은 조립 함수, 라우트 등록 목록.
// 넣지 않는 것: 비즈니스 로직. 여기는 "누가 누구를 쓰는지"만 보인다.
//
// 의존 방향: cmd/azit-server → app → (auth, azit, room, hub, ...) → store, httpx
// app 은 모두를 import 하지만, 아무도 app 을 import 하지 않는다.
package app
