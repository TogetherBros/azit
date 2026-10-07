import type { ReactNode } from 'react'

type Props = {
  mini?: boolean // 작게 보기
  shaking?: boolean // 흔들기 받는 중
  children: ReactNode // <Rail/> + <RoomList/> + 본문
}

// 앱 뼈대. 레일 | 방 목록 | 본문을 가로로 놓는다. (styles/layout.css 그림 참고)
export function Shell({ mini, shaking, children }: Props) {
  return <div className={`shell${mini ? ' mini' : ''}${shaking ? ' shake' : ''}`}>{children}</div>
}
