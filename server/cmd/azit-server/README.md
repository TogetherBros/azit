# cmd/azit-server

실제 서버 실행 파일. `go run ./cmd/azit-server` → 빌드하면 `azit-server.exe`.

`main.go` 는 얇게: 설정 읽기 → `app.New(cfg)` 로 조립 → 서버 시작 → 종료 신호(Ctrl+C) 받으면 정리.
로직은 전부 `internal/` 에 둔다. (Spring 의 `@SpringBootApplication` 클래스 역할)

> 여기에 `.go` 파일이 아니라 README 를 둔 이유: `package main` 인데 `func main()` 이 없으면
> 빌드가 깨진다. `main.go` 는 1단계 과제로 직접 만든다.
