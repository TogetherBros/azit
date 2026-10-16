import { useEffect, type ReactNode } from 'react'
import { Icon } from './Icon'

type Props = {
  title?: string // 없으면 children 이 머리를 직접 그린다
  children: ReactNode
  onClose?: () => void // 닫기 버튼 · 바깥 누르기 · Esc
}

// 화면 가운데 뜨는 창. 바깥(어두운 바탕)을 누르거나 Esc 로 닫는 건 JS 단계에서.
export function Modal({ title, children, onClose }: Props) {
  useEffect(() => {
    if (!onClose) return
    const esc = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    window.addEventListener('keydown', esc)
    return () => window.removeEventListener('keydown', esc)
  }, [onClose])

  return (
    <div className="modalbg" onClick={onClose && ((e) => { if (e.target === e.currentTarget) onClose() })}>
      <div className="modal" role="dialog" aria-modal="true" aria-label={title}>
        {title && (
          <div className="modal-head">
            <h2 className="disp modal-title">{title}</h2>
            <CloseButton onClick={onClose} />
          </div>
        )}
        {children}
      </div>
    </div>
  )
}

export function CloseButton({ onClick }: { onClick?: () => void }) {
  return (
    <button type="button" className="btn s icon sec" aria-label="닫기" onClick={onClick}>
      <Icon name="x" />
    </button>
  )
}
