import { useState } from 'react'
import type { LogEntry, LogType } from '../../../api/types'
import { Avatar } from '../../../components/Avatar'

type Props = {
  entries: LogEntry[]
  error?: string
  busy?: boolean
  onUndo?: (id: string) => void // 없으면 되돌리기 버튼을 안 보인다 (관리자만)
}

const FILTERS: (LogType | '전체')[] = ['전체', '방', '짤', '맞장구', '멤버', '설정']
const TYPE_CLASS: Record<LogType, string> = { 방: 'log-room', 짤: 'log-sticker', 맞장구: 'log-sticker', 코드: 'log-code', 멤버: 'log-member', 설정: 'log-setting' }

// 변경 기록: 누가 뭘 바꿨는지 멤버 누구나 본다. 되돌리기는 관리자 이상.
export function LogTab({ entries, error, busy, onUndo }: Props) {
  const [filter, setFilter] = useState<LogType | '전체'>('전체')
  const shown = filter === '전체' ? entries : entries.filter((l) => l.type === filter)

  return (
    <>
      <div className="row wrap">
        <h1 className="page-title sm grow">변경 기록</h1>
        <div className="row wrap" style={{ gap: 6 }}>
          {FILTERS.map((f) => (
            <button key={f} type="button" className={`chip${f === filter ? ' on' : ''}`} aria-pressed={f === filter} onClick={() => setFilter(f)}>{f}</button>
          ))}
        </div>
      </div>
      <span className="hint">누가 뭘 바꿨는지 전부 남아요. {onUndo ? '되돌리기도 기록돼요.' : '되돌리기는 관리자가 할 수 있어요.'}</span>
      {error && <span role="alert" className="ferr">{error}</span>}
      <div className="col" style={{ gap: 6 }}>
        {shown.length
          ? shown.map((l) => (
            <div key={l.id} className={`lrow logrow${l.undoneBy ? ' undone' : ''}`}>
              <span className="logrow-ago">{l.ago}</span>
              <Avatar name={l.by.nick} color={l.by.color} size={28} />
              <span className={`tag log ${TYPE_CLASS[l.type]}`}>{l.type}</span>
              <span className="logrow-text"><b>{l.by.nick}</b> {l.text}</span>
              {l.undoneBy
                ? <span className="logrow-undone">{l.undoneBy} 되돌림</span>
                : l.undoable && onUndo && (
                  <button type="button" className="btn s sec" style={{ color: 'var(--lime)' }} disabled={busy} onClick={() => onUndo(l.id)}>되돌리기</button>
                )}
            </div>
          ))
          : <div className="card mu">기록이 없어요</div>}
      </div>
    </>
  )
}
