import type { NotifKind, Notification } from '../../api/types'
import { Icon, type IconName } from '../../components/Icon'

type Filter = '전체' | '흔들기' | '놀이' | '아지트' | '칭호'

type Props = { items: Notification[]; filter?: Filter }

const FILTERS: Filter[] = ['전체', '흔들기', '놀이', '아지트', '칭호']
const ICONS: Record<NotifKind, IconName> = { shake: 'shake', game: 'game', join: 'check', kick: 'x', title: 'star' }

// 레일 아래 종 → 모든 아지트의 알림을 한곳에
export function ActivityView({ items, filter = '전체' }: Props) {
  return (
    <div className="main">
      <div className="scroll page">
        <div className="narrow-page">
          <div className="row">
            <h1 className="page-title grow">알림</h1>
            <button type="button" className="btn s sec">모두 읽음</button>
          </div>
          <div className="row wrap" style={{ gap: 6 }}>
            {FILTERS.map((f) => <button key={f} type="button" className={`chip${f === filter ? ' on' : ''}`}>{f}</button>)}
          </div>
          {items.length
            ? items.map((n) => (
              <button key={n.id} type="button" className={`lrow notif${n.read ? ' read' : ''}`}>
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
