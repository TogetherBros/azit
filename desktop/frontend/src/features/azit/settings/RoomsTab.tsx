import type { Room } from '../../../api/types'
import { roomColor } from '../../../lib/color'

type Props = {
  rooms: Room[]
  archived: string[] // 보관한 방 이름
  selectedId: string
  selectedInfo: string // "4n5rud가 만듦 · 3주 전 · 메시지 128개"
  error?: string
}

// 방 관리: 왼쪽 목록 + 보관함 | 오른쪽 고른 방 이름 바꾸기 · 처음 방 · 보관
export function RoomsTab({ rooms, archived, selectedId, selectedInfo, error }: Props) {
  const sel = rooms.find((r) => r.id === selectedId) ?? rooms[0]
  return (
    <>
      <div className="row">
        <h1 className="page-title sm grow">방 {rooms.length}개</h1>
        <button type="button" className="btn p">+ 방 만들기</button>
      </div>
      <span className="hint">방은 그냥 단톡방이에요. 아지트 멤버는 모든 방에 들어가 있어요. 지울 수는 없고 보관만 돼요.</span>

      <div className="rooms-manage">
        <div className="rooms-manage-list">
          <div className="card">
            {rooms.map((r) => (
              <button key={r.id} type="button" className={`roomrow${r.id === sel.id ? ' on' : ''}`}>
                <span className="mark xs" style={{ background: roomColor(r.name) }}>{r.name.slice(0, 1)}</span>
                <span className="grow ell b">{r.name}</span>
                {r.landing && <span className="tag">처음 방</span>}
              </button>
            ))}
          </div>
          {archived.length > 0 && (
            <div className="col" style={{ gap: 6 }}>
              <span className="mu b" style={{ fontSize: 12 }}>보관함 · {archived.length}</span>
              {archived.map((name) => (
                <div key={name} className="lrow dashed">
                  <span className="grow mu">{name}</span>
                  <button type="button" className="btn s sec">꺼내기</button>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="card rooms-manage-edit">
          <span className="disp ell" style={{ fontSize: 22 }}>{sel.name}</span>
          <span className="hint">{selectedInfo}</span>
          <div className="field">
            <label htmlFor="sr-name">방 이름</label>
            <div className="row">
              <input id="sr-name" className={`inp${error ? ' err' : ''}`} maxLength={20} defaultValue={sel.name} />
              <button type="button" className="btn sec" disabled>바꾸기</button>
            </div>
            {error && <span className="ferr">{error}</span>}
          </div>
          <div className="row">
            <button type="button" className="btn sec">이 방으로 가기</button>
            {!sel.landing && <button type="button" className="btn sec">처음 방으로 정하기</button>}
          </div>
          {sel.landing
            ? <div className="card2 hint">새 멤버가 처음 들어오는 방이라 보관할 수 없어요. 다른 방을 처음 방으로 정하면 보관할 수 있어요.</div>
            : (
              <div className="row dashed-box">
                <span className="grow hint">보관하면 목록에서 숨겨지고 메시지는 그대로 남아요. 누구나 다시 꺼낼 수 있어요.</span>
                <button type="button" className="btn s sec">보관하기</button>
              </div>
            )}
        </div>
      </div>
    </>
  )
}
