// Package store 는 SQLite 연결과 스키마를 관리한다.
//
// 넣는 것:
//   - DB 열기 (WAL 모드, 외래키 켜기 같은 SQLite 설정)
//   - migrations/ 의 SQL 을 순서대로 적용
//   - 트랜잭션 헬퍼
//
// 넣지 않는 것: 도메인 쿼리. "아지트 찾기" SQL 은 azit 패키지에, "메시지 저장"
// SQL 은 room 패키지에 둔다. Spring 에서 Repository 가 도메인 옆에 있던 것과 같다.
// 한 패키지에 모든 SQL 을 모으면 그 패키지가 모든 도메인 타입을 import 하게 되고,
// Go 는 import 순환을 컴파일 에러로 막기 때문에 금방 막힌다.
//
// migrations/: 0001_init.sql, 0002_xxx.sql 처럼 번호를 붙인 SQL 파일.
// 한 번 적용된 파일은 고치지 않고, 바꿀 게 생기면 새 번호 파일을 추가한다. (Flyway 와 같은 규칙)
package store
