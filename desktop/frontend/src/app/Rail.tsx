import type { Azit, Member } from '../api/types'
import logoSmall from '../assets/logo-small.webp'
import { Avatar } from '../components/Avatar'
import { Icon } from '../components/Icon'

type Props = {
  azits: Azit[]
  currentId: string
  me: Member
  notifCount: number
  view?: 'activity' | 'me'
}

// 왼쪽 세로 줄: 로고 · 내 아지트들 · [+] · (아래) 알림 · 내 정보
export function Rail({ azits, currentId, me, notifCount, view }: Props) {
  return (
    <nav className="rail" aria-label="아지트 목록">
      <button type="button" className="plain" aria-label="방 목록">
        <img src={logoSmall} alt="아지트" className="rail-logo" />
      </button>

      {azits.map((az) => {
        const on = az.id === currentId
        return (
          <button
            key={az.id}
            type="button"
            className={`rbtn${on ? ' on' : ''}`}
            style={on ? { background: az.color } : undefined}
            aria-label={az.name + (az.unread ? ` 새 메시지 ${az.unread}개` : '')}
            aria-current={on}
          >
            {az.short}
            {az.unread > 0 && !on && <span className="badge">{az.unread > 99 ? '99+' : az.unread}</span>}
          </button>
        )
      })}

      <button type="button" className="rbtn add" aria-label="코드로 들어가기 또는 새 아지트">
        <Icon name="plus" />
      </button>

      <div className="rail-bottom">
        <button
          type="button"
          className={`rbtn sm${view === 'activity' ? ' on' : ''}`}
          aria-label={`알림${notifCount ? ` ${notifCount}개` : ''}`}
        >
          <Icon name="bell" />
          {notifCount > 0 && <span className="badge">{notifCount}</span>}
        </button>
        <button type="button" className="plain" aria-label="내 정보">
          <Avatar name={me.nick} color={me.color} size={44} />
        </button>
      </div>
    </nav>
  )
}
