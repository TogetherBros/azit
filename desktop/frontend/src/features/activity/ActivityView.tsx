import { useState } from 'react'
import type { NotifKind, Notification } from '../../api/types'
import { Icon, type IconName } from '../../components/Icon'

type Filter = '전체' | '흔들기' | '아지트'

type Props = {
  items: Notification[]
  filter?: Filter // 미리보기용 처음 값
  error?: string
  onReadAll?: () => void
  onOpen?: (id: string) => void // 알림 누르기 → 그 방·아지트로
}

const FILTERS: Filter[] = ['전체', '흔들기', '아지트']
const ICONS: Record<NotifKind, IconName> = { shake: 'shake', game: 'game', join: 'check', kick: 'x', title: 'star' }
const MATCH: Record<Filter, (k: NotifKind) => boolean> = {
  전체: () => true,
  흔들기: (k) => k === 'shake',
  아지트: (k) => k === 'join' || k === 'kick',
}

// 레일 아래 종 → 모든 아지트의 알림을 한곳에
export function ActivityView({ items, filter: first = '전체', error, onReadAll, onOpen }: Props) {
  const [filter, setFilter] = useState<Filter>(first)
  const shown = items.filter((n) => MATCH[filter](n.kind))
  const unread = items.some((n) => !n.read)

  return (
    <div className="main">
      <div className="scroll page">
        <div className="narrow-page">
          <div className="row">
            <h1 className="page-title grow">알림</h1>
            <button type="button" className="btn s sec" disabled={!unread} onClick={onReadAll}>모두 읽음</button>
          </div>
          <div className="row wrap" style={{ gap: 6 }}>
            {FILTERS.map((f) => (
              <button key={f} type="button" className={`chip${f === filter ? ' on' : ''}`} aria-pressed={f === filter} onClick={() => setFilter(f)}>{f}</button>
            ))}
          </div>
          {error && <span role="alert" className="ferr">{error}</span>}
          {shown.length
            ? shown.map((n) => (
              <button
                key={n.id}
                type="button"
                className={`lrow notif${n.read ? ' read' : ''}`}
                disabled={n.gone}
                onClick={onOpen && !n.gone ? () => onOpen(n.id) : undefined}
              >
                <span className={`notif-ic ${n.kind}`}><Icon name={ICONS[n.kind]} size={16} /></span>
                <span className="grow col" style={{ gap: 2 }}>
                  <span className="notif-text">{n.text}</span>
                  <span className="notif-meta">{n.azit} · {n.ago}{n.gone ? ' · 이제 없는 아지트' : ''}</span>
                </span>
                {!n.read && <span className="dot" />}
              </button>
            ))
            : <div className="card empty-card">새 알림이 없어요</div>}
        </div>
      </div>
    </div>
  )
}
