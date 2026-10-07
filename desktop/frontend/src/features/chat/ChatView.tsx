import type { ChatRow, ReactOption, Room } from '../../api/types'
import { Icon } from '../../components/Icon'
import { Composer } from './Composer'
import { MessageItem } from './MessageItem'

type Props = {
  room: Room
  memberCount: number
  rows: ChatRow[]
  reactOptions: ReactOption[]
  mini?: boolean
  reconnecting?: boolean // 연결 끊김 → 주황 띠
  composer?: Parameters<typeof Composer>[0]
}

// 오른쪽 본문: 머리(방 이름 · N명) · 메시지 목록 · 입력창
export function ChatView({ room, memberCount, rows, reactOptions, mini, reconnecting, composer }: Props) {
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
          <button type="button" className="btn s sec" aria-label={`방 정보 · ${memberCount}명`}>{memberCount}명</button>
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

      <div className="scroll chat-body">
        {rows.length === 0 && (
          <div className="chat-empty">
            <span className="chat-empty-title">아직 조용해요</span>
            <span className="mu">첫 마디를 남기거나 친구를 흔들어 보세요</span>
          </div>
        )}
        {rows.map((row) =>
          row.type === 'date'
            ? <div key={row.id} className="datediv">{row.label}</div>
            : (
              <MessageItem
                key={row.msg.id}
                msg={row.msg}
                cont={row.cont}
                showTime={row.showTime}
                reactOptions={reactOptions}
              />
            ),
        )}
      </div>

      <Composer {...composer} />
    </div>
  )
}
