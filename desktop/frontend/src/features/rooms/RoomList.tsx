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
}

// 가운데 칸: 아지트 이름 · 놀이터 · 방 찾기 · 방 목록 · [+ 방 만들기][초대 코드]
export function RoomList({ azit, rooms, currentRoomId, mini, query, active }: Props) {
  return (
    <aside className="side" aria-label="방 목록">
      <div className="side-top">
        <div className="side-azit">
          <Mark label={azit.short} color={azit.color} size="sm" />
          <div className="grow col" style={{ gap: 3 }}>
            <span className="side-azit-name">{azit.name}</span>
            <span className="side-azit-meta">아지트 · {azit.memberCount}명 · 모두 주인</span>
          </div>
          <button
            type="button"
            className="btn s icon sec"
            aria-label={mini ? '크게 보기' : '작게 보기'}
            title={mini ? '크게 보기' : '작게 보기'}
          >
            <Icon name={mini ? 'wide' : 'narrow'} size={16} />
          </button>
          <button type="button" className={`btn s icon ${active === 'settings' ? 'p' : 'sec'}`} aria-label="아지트 설정">
            <Icon name="gear" size={16} />
          </button>
        </div>

        <button type="button" className={`btn s ${active === 'hub' ? 'p' : 'sec'}`}>
          <Icon name="game" size={14} />
          놀이터
        </button>

        <div className="search">
          <Icon name="search" size={15} />
          <label htmlFor="room-q" className="sr">방 찾기</label>
          <input id="room-q" placeholder="방 이름 찾기" defaultValue={query} />
        </div>
      </div>

      <nav className="scroll room-nav" aria-label="방">
        {rooms.length === 0 && <p className="hint room-empty">맞는 방이 없어요</p>}
        {rooms.map((r) => {
          const on = r.id === currentRoomId
          return (
            <button key={r.id} type="button" className={`roomrow${on ? ' on' : ''}`} aria-current={on ? 'page' : undefined}>
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
        <button type="button" className="btn s gh grow">+ 방 만들기</button>
        <button type="button" className="btn s p grow">초대 코드</button>
      </div>
    </aside>
  )
}
