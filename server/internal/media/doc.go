// Package media 는 사진과 짤 파일을 저장하고 꺼내 준다. (MVP 2)
//
// 테이블: stickers
//
// 넣는 것: 업로드 크기 제한(사진 10MB, 짤 5MB), 이미지 줄이기,
// 파일 키 발급, 디스크 저장. 메시지에는 파일 자체가 아니라 file_key 만 들어간다.
package media
