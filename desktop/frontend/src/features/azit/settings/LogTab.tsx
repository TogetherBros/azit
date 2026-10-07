import type { LogEntry, LogType } from '../../../api/types'
import { Avatar } from '../../../components/Avatar'

type Props = { entries: LogEntry[]; filter?: LogType | '전체' }

const FILTERS: (LogType | '전체')[] = ['전체', '방', '짤', '맞장구', '코드', '멤버', '설정']
const TYPE_CLASS: Record<LogType, string> = { 방: 'log-room', 짤: 'log-sticker', 맞장구: 'log-sticker', 코드: 'log-code', 멤버: 'log-member', 설정: 'log-setting' }

// 변경 기록: 모두 주인이라 누가 뭘 바꿨는지 전부 남고, 누구나 되돌릴 수 있다.
export function LogTab({ entries, filter = '전체' }: Props) {
  return (
    <>
      <div className="row wrap">
        <h1 className="page-title sm grow">변경 기록</h1>
        <div className="row wrap" style={{ gap: 6 }}>
          {FILTERS.map((f) => <button key={f} type="button" className={`chip${f === filter ? ' on' : ''}`}>{f}</button>)}
        </div>
      </div>
      <span className="hint">모두가 주인이라 누가 뭘 바꿨는지 전부 남아요. 되돌리기도 기록돼요.</span>
      <div className="col" style={{ gap: 6 }}>
        {entries.length
          ? entries.map((l) => (
            <div key={l.id} className={`lrow logrow${l.undoneBy ? ' undone' : ''}`}>
              <span className="logrow-ago">{l.ago}</span>
              <Avatar name={l.by.nick} color={l.by.color} size={28} />
              <span className={`tag log ${TYPE_CLASS[l.type]}`}>{l.type}</span>
              <span className="logrow-text"><b>{l.by.nick}</b> {l.text}</span>
              {l.undoneBy
                ? <span className="logrow-undone">{l.undoneBy} 되돌림</span>
                : l.undoable && <button type="button" className="btn s sec" style={{ color: 'var(--lime)' }}>되돌리기</button>}
            </div>
          ))
          : <div className="card mu">기록이 없어요</div>}
      </div>
    </>
  )
}
