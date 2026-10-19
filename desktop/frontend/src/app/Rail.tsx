import type { Azit, Member } from '../api/types'
import logoSmall from '../assets/logo-small.webp'
import { Avatar } from '../components/Avatar'
import { Icon } from '../components/Icon'

type Props = {
  azits: Azit[]
  currentId?: string // 지금 보고 있는 아지트. 홈·설정이면 없음
  me: Member
  notifCount: number
  view?: 'home' | 'activity' | 'me' | 'settings'
  meSrc?: string // 내 프사 주소
  onHome?: () => void
  onMe?: () => void
  onBell?: () => void // 알림
  onSettings?: () => void // 앱 설정
  onAzit?: (id: string) => void
  onAdd?: () => void // + : 코드로 들어가기 · 새 아지트 (홈으로 간다)
}

// 왼쪽 세로 줄: 로고(= 홈) · 내 아지트들 · [+] · (아래) 알림 · 앱 설정 · 내 정보
export function Rail({ azits, currentId, me, notifCount, view, meSrc, onHome, onMe, onBell, onSettings, onAzit, onAdd }: Props) {
  return (
    <nav className="rail" aria-label="아지트 목록">
      <button
        type="button"
        className={`plain rail-home${view === 'home' ? ' on' : ''}`}
        aria-label="홈 · 내 아지트 모아 보기"
        aria-current={view === 'home' ? 'page' : undefined}
        onClick={onHome}
      >
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
            title={az.name}
            onClick={onAzit && (() => onAzit(az.id))}
          >
            {az.short}
            {az.unread > 0 && !on && <span className="badge">{az.unread > 99 ? '99+' : az.unread}</span>}
          </button>
        )
      })}

      <button type="button" className="rbtn add" aria-label="코드로 들어가기 또는 새 아지트" onClick={onAdd}>
        <Icon name="plus" />
      </button>

      <div className="rail-bottom">
        <button
          type="button"
          className={`rbtn sm${view === 'activity' ? ' on' : ''}`}
          aria-label={`알림${notifCount ? ` ${notifCount}개` : ''}`}
          aria-current={view === 'activity' ? 'page' : undefined}
          onClick={onBell}
        >
          <Icon name="bell" />
          {notifCount > 0 && <span className="badge">{notifCount > 99 ? '99+' : notifCount}</span>}
        </button>
        <button
          type="button"
          className={`rbtn sm${view === 'settings' ? ' on' : ''}`}
          aria-label="앱 설정"
          aria-current={view === 'settings' ? 'page' : undefined}
          onClick={onSettings}
        >
          <Icon name="gear" />
        </button>
        <button
          type="button"
          className={`plain rail-me${view === 'me' ? ' on' : ''}`}
          aria-label="내 정보"
          aria-current={view === 'me' ? 'page' : undefined}
          onClick={onMe}
        >
          <Avatar name={me.nick} color={me.color} size={44} src={meSrc} />
        </button>
      </div>
    </nav>
  )
}
