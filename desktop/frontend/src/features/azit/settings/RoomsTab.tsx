import { useEffect, useState } from 'react'
import type { Room } from '../../../api/types'
import { Icon } from '../../../components/Icon'
import { roomColor } from '../../../lib/color'

type Props = {
  rooms: Room[]
  archived?: string[] // 보관한 방 이름 (아직 서버에 없음)
  selectedId: string
  selectedInfo?: string // "4n5rud가 만듦 · 3주 전 · 메시지 128개"
  error?: string
  busy?: boolean
  onSelect?: (id: string) => void
  onRename?: (id: string, name: string) => void
  onMove?: (id: string, dir: -1 | 1) => void // 순서 위·아래
  onOpen?: (id: string) => void // 이 방으로 가기
  onNewRoom?: () => void
}

// 방 관리: 왼쪽 목록(순서 바꾸기) | 오른쪽 고른 방 이름 바꾸기
export function RoomsTab({ rooms, archived = [], selectedId, selectedInfo, error, busy, onSelect, onRename, onMove, onOpen, onNewRoom }: Props) {
  const sel = rooms.find((r) => r.id === selectedId) ?? rooms[0]
  const [name, setName] = useState(sel?.name ?? '')
  useEffect(() => setName(sel?.name ?? ''), [sel?.id, sel?.name])

  if (!sel) return null
  const changed = name.trim() !== '' && name.trim() !== sel.name

  return (
    <>
      <div className="row">
        <h1 className="page-title sm grow">방 {rooms.length}개</h1>
        <button type="button" className="btn p" onClick={onNewRoom}>+ 방 만들기</button>
      </div>
      <span className="hint">방은 그냥 단톡방이에요. 아지트 멤버는 모든 방에 들어가 있어요. 지울 수는 없어요. 여기 순서가 모두의 방 목록 순서예요.</span>

      <div className="rooms-manage">
        <div className="rooms-manage-list">
          <div className="card">
            {rooms.map((r, i) => (
              <div key={r.id} className={`roomrow${r.id === sel.id ? ' on' : ''}`} style={{ gap: 4 }}>
                <button type="button" className="plain row grow" style={{ gap: 10, minWidth: 0 }} onClick={onSelect && (() => onSelect(r.id))}>
                  <span className="mark xs" style={{ background: roomColor(r.name) }}>{r.name.slice(0, 1)}</span>
                  <span className="grow ell b" style={{ textAlign: 'left' }}>{r.name}</span>
                  {r.landing && <span className="tag">처음 방</span>}
                </button>
                {onMove && (
                  <>
                    <button type="button" className="btn s icon sec room-move" aria-label={`${r.name} 위로`} disabled={i === 0 || busy} onClick={() => onMove(r.id, -1)}>
                      <Icon name="chev" size={12} />
                    </button>
                    <button type="button" className="btn s icon sec room-move down" aria-label={`${r.name} 아래로`} disabled={i === rooms.length - 1 || busy} onClick={() => onMove(r.id, 1)}>
                      <Icon name="chev" size={12} />
                    </button>
                  </>
                )}
              </div>
            ))}
          </div>
          {archived.length > 0 && (
            <div className="col" style={{ gap: 6 }}>
              <span className="mu b" style={{ fontSize: 12 }}>보관함 · {archived.length}</span>
              {archived.map((n) => (
                <div key={n} className="lrow dashed">
                  <span className="grow mu">{n}</span>
                  <button type="button" className="btn s sec">꺼내기</button>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="card rooms-manage-edit">
          <span className="disp ell" style={{ fontSize: 22 }}>{sel.name}</span>
          {selectedInfo && <span className="hint">{selectedInfo}</span>}
          <form
            className="field"
            noValidate
            onSubmit={(e) => {
              e.preventDefault()
              if (changed && !busy) onRename?.(sel.id, name.trim())
            }}
          >
            <label htmlFor="sr-name">방 이름</label>
            <div className="row">
              <input id="sr-name" className={`inp${error ? ' err' : ''}`} maxLength={20} value={name} onChange={(e) => setName(e.target.value)} />
              <button type="submit" className="btn sec" disabled={!changed || busy}>바꾸기</button>
            </div>
            {error && <span className="ferr">{error}</span>}
          </form>
          <div className="row">
            <button type="button" className="btn sec" onClick={onOpen && (() => onOpen(sel.id))}>이 방으로 가기</button>
          </div>
        </div>
      </div>
    </>
  )
}
