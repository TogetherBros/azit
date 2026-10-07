import type { ReactNode } from 'react'

export type SettingsTab = 'general' | 'codes' | 'members' | 'rooms' | 'stickers' | 'log' | 'leave'

const TABS: [SettingsTab, string][] = [
  ['general', '기본 정보'],
  ['codes', '초대 코드'],
  ['members', '멤버'],
  ['rooms', '방'],
  ['stickers', '짤 · 맞장구'],
  ['log', '변경 기록'],
  ['leave', '나가기'],
]

type Props = { tab: SettingsTab; children: ReactNode }

// 아지트 설정 틀: 왼쪽 탭 목록 + 오른쪽 내용. 창이 좁으면 탭이 칩으로 바뀐다.
export function SettingsView({ tab, children }: Props) {
  return (
    <div className="main settings">
      <nav className="snav" aria-label="아지트 설정">
        <span className="snav-title">아지트 설정</span>
        <span className="hint">모두 주인 · 삭제만 못 해요. 바꾼 건 전부 기록돼요</span>
        {TABS.map(([t, label]) => (
          <button
            key={t}
            type="button"
            className={`btn snav-btn${t === tab ? ' on' : ''}${t === 'leave' ? ' leave' : ''}`}
            aria-current={t === tab ? 'page' : undefined}
          >
            {label}
          </button>
        ))}
      </nav>
      <div className="scroll page settings-body">
        <div className="row wrap show-sm" style={{ gap: 6 }}>
          {TABS.map(([t, label]) => <button key={t} type="button" className={`chip${t === tab ? ' on' : ''}`}>{label}</button>)}
        </div>
        {children}
      </div>
    </div>
  )
}
