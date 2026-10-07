# cmd/azitctl

운영용 CLI. 서버와 같은 `internal/store`, `internal/auth` 를 재사용한다.

예: `azitctl account disable <id>` (계정 멈추기), `azitctl migrate` (스키마 적용).
Spring 이면 별도 프로젝트를 만들 일을, Go 에서는 `cmd/` 폴더 하나로 끝낸다.
