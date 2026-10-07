// Package httpx 는 모든 HTTP 핸들러가 같이 쓰는 도구를 모은다.
//
// 이름이 http 가 아닌 이유: 표준 라이브러리 net/http 와 이름이 겹치면
// import 할 때마다 별칭을 붙여야 해서, 관례적으로 x 를 붙인다.
//
// 넣는 것:
//   - JSON 응답/에러 응답 쓰기 헬퍼          (Spring: ResponseEntity, @ControllerAdvice)
//   - 미들웨어: 로깅, panic 복구, 요청 ID    (Spring: Filter, HandlerInterceptor)
//   - 요청 context 에서 로그인한 계정 꺼내기  (Spring: @AuthenticationPrincipal)
//
// 넣지 않는 것: 특정 도메인(아지트, 방)에 대한 지식. 여기는 어느 도메인도 import 하지 않는다.
package httpx
