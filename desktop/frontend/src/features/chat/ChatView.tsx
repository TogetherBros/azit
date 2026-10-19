import { useEffect, useLayoutEffect, useRef } from 'react'
import type { ChatRow, ReactOption, Room } from '../../api/types'
import { Icon } from '../../components/Icon'
import { Composer } from './Composer'
import { MessageItem } from './MessageItem'

type ItemHandlers = Pick<Parameters<typeof MessageItem>[0], 'onReact' | 'onToggle' | 'onEdit' | 'onDelete' | 'onRetry' | 'onImage' | 'onProfile' | 'canDelete' | 'memeOptions' | 'onReply' | 'onSaveMeme' | 'onJump'>

type Props = {
  room: Room
  memberCount: number
  rows: ChatRow[]
  reactOptions: ReactOption[]
  mini?: boolean
  reconnecting?: boolean // 연결 끊김 → 주황 띠
  composer?: Parameters<typeof Composer>[0]
  hasMore?: boolean // 위에 더 오래된 메시지가 있음
  loadingMore?: boolean
  error?: string
  onMore?: () => void
  onRoomInfo?: () => void // 머리의 "N명" 누르기 → 방 정보
  jumpTo?: { id: string; n: number } // 이 메시지로 스크롤하고 잠깐 반짝 (n 은 같은 메시지를 또 눌러도 다시 하려고)
} & ItemHandlers

// 오른쪽 본문: 머리(방 이름 · N명) · 메시지 목록 · 입력창
// 새 메시지가 오면 맨 아래에 있을 때만 따라 내려간다. 위로 더 불러오면 보던 자리를 지킨다.
export function ChatView({ room, memberCount, rows, reactOptions, mini, reconnecting, composer, hasMore, loadingMore, error, onMore, onRoomInfo, jumpTo, ...item }: Props) {
  const body = useRef<HTMLDivElement>(null)
  const last = useRef({ first: '', last: '', height: 0, atBottom: true, room: '' })

  const rowKey = (r?: ChatRow) => (r ? (r.type === 'msg' ? r.msg.id : r.id) : '')

  // 그리기 직전 스크롤 상태를 기억해 둔다
  const onScroll = () => {
    const el = body.current
    if (!el) return
    last.current.atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 80
  }

  useLayoutEffect(() => {
    const el = body.current
    if (!el) return
    const prev = last.current
    const first = rowKey(rows[0])
    const lastKey = rowKey(rows[rows.length - 1])
    if (prev.room !== room.id) {
      el.scrollTop = el.scrollHeight // 방을 바꾸면 맨 아래로
    } else if (prev.first && first !== prev.first && lastKey === prev.last) {
      el.scrollTop += el.scrollHeight - prev.height // 위에 붙었으면 보던 자리 그대로
    } else if (lastKey !== prev.last) {
      const mineNew = rows[rows.length - 1]?.type === 'msg' && (rows[rows.length - 1] as { msg: { mine?: boolean } }).msg.mine
      if (prev.atBottom || mineNew) el.scrollTop = el.scrollHeight
    }
    last.current = { first, last: lastKey, height: el.scrollHeight, atBottom: prev.room !== room.id || prev.atBottom, room: room.id }
  }, [rows, room.id])

  // 답장 인용을 누르면 원래 메시지로 (위로 불러온 뒤 위치 지키기가 끝난 다음에 돈다)
  useEffect(() => {
    if (!jumpTo) return
    const el = body.current?.querySelector<HTMLElement>(`[data-msg-id="${CSS.escape(jumpTo.id)}"]`)
    if (!el) return
    el.scrollIntoView({ block: 'center', behavior: 'smooth' })
    el.classList.remove('flash')
    void el.offsetWidth // 같은 메시지를 또 눌러도 애니메이션이 다시 돌게
    el.classList.add('flash')
    const t = window.setTimeout(() => el.classList.remove('flash'), 1600)
    return () => window.clearTimeout(t)
  }, [jumpTo])

  return (
    <div className="main">
      <header className="head">
        {mini && (
          <button type="button" className="btn s icon sec" aria-label="방 목록으로">
            <Icon name="back" />
          </button>
        )}
        <h1 className="ell head-title">{room.name}</h1>
        <span className="grow">
          <button type="button" className="btn s sec" aria-label={`방 정보 · ${memberCount}명`} onClick={onRoomInfo}>{memberCount}명</button>
        </span>
        {mini && (
          <button type="button" className="btn s icon sec" aria-label="크게 보기" title="크게 보기">
            <Icon name="wide" size={16} />
          </button>
        )}
      </header>

      {reconnecting && (
        <div className="banner warn" role="status">
          <span className="grow">연결이 끊겼어요. 다시 잇는 중이에요 · 쓰던 메시지는 그대로 있어요</span>
        </div>
      )}

      <div className="scroll chat-body" ref={body} onScroll={onScroll}>
        {hasMore && (
          <button type="button" className="btn s sec chat-more" disabled={loadingMore} onClick={onMore}>
            {loadingMore ? '불러오는 중…' : '이전 메시지 더 보기'}
          </button>
        )}
        {rows.length === 0 && (
          <div className="chat-empty">
            <span className="chat-empty-title">아직 조용해요</span>
            <span className="mu">첫 마디를 남기거나 친구를 흔들어 보세요</span>
          </div>
        )}
        {rows.map((row) =>
          row.type === 'date'
            ? <div key={row.id} className="datediv">{row.label}</div>
            : row.type === 'read'
              ? <div key={row.id} className="readline">여기까지 읽었어요</div>
              : (
                <MessageItem
                  key={row.msg.id}
                  msg={row.msg}
                  cont={row.cont}
                  showTime={row.showTime}
                  reactOptions={reactOptions}
                  {...item}
                />
              ),
        )}
        {error && <span role="alert" className="ferr chat-err">{error}</span>}
      </div>

      <Composer {...composer} />
    </div>
  )
}
