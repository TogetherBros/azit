# api
Go(bridge) 와의 연결부. React 는 서버를 직접 부르지 않는다.
- 생성된 바인딩(`frontend/bindings/`, `wails3 generate bindings`)을 감싸는 함수
- `Events.On("message", ...)` 같은 Go → React 이벤트 구독
화면 코드는 이 폴더만 import 하고, 바인딩을 직접 부르지 않는다. (바인딩 경로가 바뀌어도 여기만 고치면 됨)
