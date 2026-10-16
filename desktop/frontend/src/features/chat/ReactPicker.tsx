import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import type { ReactOption } from '../../api/types'

type Props = {
  anchor: HTMLElement // 이 버튼 옆에 붙어서 뜬다
  reacts: ReactOption[] // 아지트 맞장구 목록 (최대 10개)
  memes: ReactOption[] // 아지트 짤 전부 (목록에 없어도 달 수 있음)
  isOn: (o: ReactOption) => boolean // 내가 이미 단 것
  onPick: (o: ReactOption) => void // 누르면 달기 / 떼기
  onClose: () => void
}

const W = 268 // .rpick 너비와 같게
const GAP = 6

// 맞장구 고르기 창. 메시지의 맞장구 버튼에 붙어서 뜨고, 바깥 누르기·Esc·하나 고르기로 닫힌다.
// 창 밖으로 안 나가게: 아래 자리가 모자라면 위로 뒤집고, 좌우는 창 안으로 민다.
export function ReactPicker({ anchor, reacts, memes, isOn, onPick, onClose }: Props) {
  const box = useRef<HTMLDivElement>(null)
  const [pos, setPos] = useState<{ left: number; top: number }>()

  useLayoutEffect(() => {
    const a = anchor.getBoundingClientRect()
    const h = box.current?.offsetHeight ?? 0
    const below = a.bottom + GAP
    const top = below + h <= window.innerHeight - 8 ? below : Math.max(8, a.top - GAP - h)
    const left = Math.min(Math.max(8, a.right - W), window.innerWidth - W - 8)
    setPos({ left, top })
  }, [anchor])

  useEffect(() => {
    const down = (e: MouseEvent) => {
      const t = e.target as Node
      if (!box.current?.contains(t) && !anchor.contains(t)) onClose()
    }
    const esc = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    document.addEventListener('mousedown', down)
    window.addEventListener('keydown', esc)
    window.addEventListener('resize', onClose)
    return () => {
      document.removeEventListener('mousedown', down)
      window.removeEventListener('keydown', esc)
      window.removeEventListener('resize', onClose)
    }
  }, [anchor, onClose])

  const pick = (o: ReactOption) => { onPick(o); onClose() }

  const tile = (o: ReactOption) => (
    <button
      key={o.id}
      type="button"
      className={`rpick-tile${isOn(o) ? ' on' : ''}`}
      aria-pressed={isOn(o)}
      aria-label={o.text ?? `짤 ${o.name ?? ''}`}
      title={o.name}
      onClick={() => pick(o)}
    >
      {o.img ? <img src={o.img} alt="" /> : <span className="rpick-text">{o.text}</span>}
    </button>
  )

  return (
    <div
      ref={box}
      className="rpick"
      role="dialog"
      aria-label="맞장구 고르기"
      style={pos ? { left: pos.left, top: pos.top } : { visibility: 'hidden' }}
    >
      <span className="rpick-cap">맞장구</span>
      {reacts.length > 0
        ? <div className="rpick-grid">{reacts.map(tile)}</div>
        : <span className="hint rpick-empty">아지트 설정 → 맞장구에서 넣을 수 있어요</span>}
      <span className="rpick-cap">아지트 짤</span>
      {memes.length > 0
        ? <div className="rpick-grid rpick-memes scroll">{memes.map(tile)}</div>
        : <span className="hint rpick-empty">아직 짤이 없어요</span>}
    </div>
  )
}
