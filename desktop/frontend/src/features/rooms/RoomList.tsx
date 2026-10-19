import type { Azit, Room } from '../../api/types'
import { Icon } from '../../components/Icon'
import { Mark } from '../../components/Mark'
import { roomColor } from '../../lib/color'

type Props = {
  azit: Azit
  rooms: Room[] // 최근 메시지 순으로 정렬돼서 온다
  currentRoomId?: string
  mini?: boolean
  query?: string
  active?: 'settings' | 'hub' // 본문에 설정/놀이터가 떠 있으면 그 버튼을 라임으로
  onBack?: () => void // ← 홈으로
  onRoom?: (id: string) => void
  onSettings?: () => void
  onNewRoom?: () => void
  onInvite?: () => void
  onQuery?: (q: string) => void
  roleLabel?: string // "모두 주인" 대신 "관리자" 처럼
  noHub?: boolean // 놀이터가 아직 없으면 숨긴다
  notifCount?: number // 안 읽은 알림 수. 아지트 안엔 레일(종)이 없어서 "홈으로" 옆에 보여 준다
}

// 아지트 안 왼쪽 칸: ← 홈 · 아지트 이름 · 놀이터 · 방 찾기 · 방 목록 · [+ 방 만들기][초대 코드]
// 아지트에 들어오면 레일(다른 아지트들)은 숨기고 이 칸만 보인다. 나갈 때는 ← 로 홈.
export function RoomList({ azit, rooms, currentRoomId, mini, query, active, onBack, onRoom, onSettings, onNewRoom, onInvite, onQuery, roleLabel = '모두 주인', noHub, notifCount = 0 }: Props) {
  return (
    <aside className="side" aria-label="방 목록">
      <div className="side-top">
        <button type="button" className="plain side-back" onClick={onBack} aria-label={`홈으로${notifCount ? ` · 새 알림 ${notifCount}개` : ''}`}>
          <Icon name="back" size={16} />
          홈으로
          {notifCount > 0 && <span className="side-back-badge"><Icon name="bell" size={11} />{notifCount > 99 ? '99+' : notifCount}</span>}
        </button>
        <div className="side-azit">
          <Mark label={azit.short} color={azit.color} size="sm" />
          <div className="grow col" style={{ gap: 3 }}>
            <span className="side-azit-name">{azit.name}</span>
            <span className="side-azit-meta">아지트 · {azit.memberCount}명 · {roleLabel}</span>
          </div>
          {!noHub && (
            <button
              type="button"
              className="btn s icon sec"
              aria-label={mini ? '크게 보기' : '작게 보기'}
              title={mini ? '크게 보기' : '작게 보기'}
            >
              <Icon name={mini ? 'wide' : 'narrow'} size={16} />
            </button>
          )}
          <button type="button" className={`btn s icon ${active === 'settings' ? 'p' : 'sec'}`} aria-label="아지트 설정" title="아지트 설정" onClick={onSettings}>
            <Icon name="gear" size={16} />
          </button>
        </div>

        {!noHub && (
          <button type="button" className={`btn s ${active === 'hub' ? 'p' : 'sec'}`}>
            <Icon name="game" size={14} />
            놀이터
          </button>
        )}

        <div className="search">
          <Icon name="search" size={15} />
          <label htmlFor="room-q" className="sr">방 찾기</label>
          <input id="room-q" placeholder="방 이름 찾기" value={query ?? ''} onChange={onQuery && ((e) => onQuery(e.target.value))} readOnly={!onQuery} />
        </div>
      </div>

      <nav className="scroll room-nav" aria-label="방">
        {rooms.length === 0 && <p className="hint room-empty">맞는 방이 없어요</p>}
        {rooms.map((r) => {
          const on = r.id === currentRoomId
          return (
            <button key={r.id} type="button" className={`roomrow${on ? ' on' : ''}`} aria-current={on ? 'page' : undefined} onClick={onRoom && (() => onRoom(r.id))}>
              <Mark label={r.name.slice(0, 1)} color={roomColor(r.name)} size="md" />
              <span className="roomrow-body">
                <span className="row" style={{ gap: 6 }}>
                  <span className="ell roomrow-name">{r.name}</span>
                  {r.landing && <span className="tag">처음 방</span>}
                </span>
                <span className="ell roomrow-preview">
                  {r.liar && (
                    <span className="roomrow-live">라이어 판 {r.liar === 'lobby' ? '모이는 중' : '한창'} · </span>
                  )}
                  {r.preview}
                </span>
              </span>
              <span className="roomrow-meta">
                <span className="roomrow-time">{r.lastAgo}</span>
                {r.unread > 0 && !on
                  ? <span className="badge">{r.unread > 99 ? '99+' : r.unread}</span>
                  : <span className="spacer" />}
              </span>
            </button>
          )
        })}
      </nav>

      <div className="side-foot">
        {onNewRoom && <button type="button" className="btn s gh grow" onClick={onNewRoom}>+ 방 만들기</button>}
        <button type="button" className="btn s p grow" onClick={onInvite}>초대 코드</button>
      </div>
    </aside>
  )
}
