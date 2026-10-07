import type { ReactNode } from 'react'
import { Icon } from './Icon'

type Props = {
  title?: string // 없으면 children 이 머리를 직접 그린다
  children: ReactNode
}

// 화면 가운데 뜨는 창. 바깥(어두운 바탕)을 누르거나 Esc 로 닫는 건 JS 단계에서.
export function Modal({ title, children }: Props) {
  return (
    <div className="modalbg">
      <div className="modal" role="dialog" aria-modal="true" aria-label={title}>
        {title && (
          <div className="modal-head">
            <h2 className="disp modal-title">{title}</h2>
            <CloseButton />
          </div>
        )}
        {children}
      </div>
    </div>
  )
}

export function CloseButton() {
  return (
    <button type="button" className="btn s icon sec" aria-label="닫기">
      <Icon name="x" />
    </button>
  )
}
