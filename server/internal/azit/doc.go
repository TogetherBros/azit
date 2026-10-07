// Package azit 은 아지트(친구 무리)와 그 멤버, 초대 코드를 다룬다.
//
// 테이블: azits, azit_members, invite_codes (MVP 2: azit_bans, audit_log)
//
// 넣는 것: 아지트 만들기, 아지트별 닉네임, 초대 코드 발급·폐기·사용,
// 코드 입장 오류 6종(없음, 폐기, 만료, 다 씀, 막힘, 꽉 참)을 구분하는 에러 값.
//
// 패키지 이름이 azit 이라 밖에서는 azit.Service, azit.Member 처럼 읽힌다.
// 그래서 타입 이름에 Azit 을 또 붙이지 않는다 (azit.AzitService ✗).
package azit
