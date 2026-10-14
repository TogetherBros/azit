import type { ReactNode } from 'react'
import { WindowBar } from './WindowBar'

// 창 전체: 위에 창 조작 줄, 아래에 화면
export function WindowFrame({ children }: { children: ReactNode }) {
  return (
    <div className="win">
      <WindowBar />
      <div className="win-body">{children}</div>
    </div>
  )
}
