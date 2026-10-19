import type { ReactNode } from 'react'
import { Avatar } from '../../components/Avatar'
import { Seg, Toggle } from '../../components/Controls'
import type { Profile } from '../../components/ProfileCard'
import { DEFAULT_PREFS, type AppPrefs } from '../../lib/prefs'

export { DEFAULT_PREFS, type AppPrefs, type FontSize } from '../../lib/prefs'

export type AppTab = 'account' | 'notify' | 'look' | 'app'

const TABS: [AppTab, string][] = [
  ['account', '내 계정'],
  ['notify', '알림'],
  ['look', '화면'],
  ['app', '앱'],
]

type Props = {
  tab: AppTab
  me: Profile
  prefs?: AppPrefs
  version?: string
  onTab?: (t: AppTab) => void
  onPrefs?: (p: AppPrefs) => void
  autostartSupported?: boolean // 이 운영체제에서 자동 실행을 켤 수 있는지 (지금은 Windows 만)
  appError?: string // 자동 실행을 못 바꿨을 때
  onEditProfile?: () => void
  onLogout?: () => void
}

// 앱 설정: 아지트와 상관없이 이 앱 자체의 설정 (아지트 설정과 같은 틀). 이 컴퓨터에만 저장된다.
export function AppSettingsView({ tab, me, prefs = DEFAULT_PREFS, version = '0.1.0', onTab, onPrefs, autostartSupported = false, appError, onEditProfile, onLogout }: Props) {
  const set = <K extends keyof AppPrefs>(k: K) => (v: AppPrefs[K]) => onPrefs?.({ ...prefs, [k]: v })

  return (
    <div className="main settings">
      <nav className="snav" aria-label="앱 설정">
        <span className="snav-title">앱 설정</span>
        <span className="hint">이 컴퓨터에서만 바뀌어요</span>
        {TABS.map(([t, label]) => (
          <button
            key={t}
            type="button"
            className={`btn snav-btn${t === tab ? ' on' : ''}`}
            aria-current={t === tab ? 'page' : undefined}
            onClick={onTab && (() => onTab(t))}
          >
            {label}
          </button>
        ))}
        {/* 이 컴퓨터에서만 나간다. 아지트와 대화는 그대로 */}
        <button type="button" className="btn snav-btn leave" onClick={onLogout}>로그아웃</button>
      </nav>

      <div className="scroll page settings-body app-set">
        <div className="row wrap show-sm" style={{ gap: 6 }}>
          {TABS.map(([t, label]) => <button key={t} type="button" className={`chip${t === tab ? ' on' : ''}`} onClick={onTab && (() => onTab(t))}>{label}</button>)}
        </div>

        {tab === 'account' && <>
          <h1 className="page-title sm">내 계정</h1>
          <div className="card row" style={{ gap: 14 }}>
            <Avatar name={me.nickname || '?'} color={me.color} size={56} src={me.avatarSrc} />
            <div className="grow col" style={{ gap: 2 }}>
              <span className="disp" style={{ fontSize: 20 }}>{me.nickname}</span>
              <span className="mono b" style={{ fontSize: 12, color: me.color }}>@{me.name}</span>
            </div>
            <button type="button" className="btn s sec" onClick={onEditProfile}>프로필 고치기</button>
          </div>
          <Row title="구글 계정" hint="구글로 로그인해 있어요">
            <span className="tag lg">연결됨</span>
          </Row>
        </>}

        {tab === 'notify' && <>
          <h1 className="page-title sm">알림</h1>
          <Row title="흔들기 받으면 창 흔들기" hint="끄면 알림만 쌓여요 (종 숫자는 그대로 올라가요)">
            <Toggle on={prefs.shakeWindow} label="흔들기 받으면 창 흔들기" onChange={set('shakeWindow')} />
          </Row>
        </>}

        {tab === 'look' && <>
          <h1 className="page-title sm">화면</h1>
          <Row title="글자 크기" hint="채팅 말풍선 글자가 바뀌어요">
            <Seg label="글자 크기" value={prefs.fontSize} options={[['s', '작게'], ['m', '보통'], ['l', '크게']]} onChange={set('fontSize')} />
          </Row>
          <div className={`card2 app-set-sample fs-${prefs.fontSize}`} aria-label="글자 크기 미리보기">
            <span className="app-set-bubble">오늘 마크 서버 열림?</span>
            <span className="app-set-bubble mine">열림 ㅋㅋ 근데 엔드는 아직 가지 마라</span>
          </div>
          <Row title="시간 표시">
            <Seg label="시간 표시" value={prefs.clock} options={[['12', '오후 2:01'], ['24', '14:01']]} onChange={set('clock')} />
          </Row>
          <Row title="움직임 줄이기" hint="창 흔들림, 반짝임 같은 움직임을 꺼요">
            <Toggle on={prefs.reduceMotion} label="움직임 줄이기" onChange={set('reduceMotion')} />
          </Row>
        </>}

        {tab === 'app' && <>
          <h1 className="page-title sm">앱</h1>
          <Row title="컴퓨터 켜면 같이 켜기" hint={autostartSupported ? '켜지면 창 없이 트레이에서 기다려요' : '지금은 Windows 에서만 돼요'}>
            <Toggle on={prefs.autostart} label="컴퓨터 켜면 같이 켜기" disabled={!autostartSupported} onChange={set('autostart')} />
          </Row>
          <Row title="창 닫으면 숨기기" hint="X 를 눌러도 꺼지지 않고 트레이로 가요. 끄려면 트레이 아이콘 메뉴의 끝내기">
            <Toggle on={prefs.closeToTray} label="창 닫으면 숨기기" onChange={set('closeToTray')} />
          </Row>
          {appError && <span role="alert" className="ferr">{appError}</span>}
          <Row title="버전" hint="새 버전이 나오면 알려 줘요">
            <span className="mono b mu" style={{ fontSize: 13 }}>v{version}</span>
          </Row>
        </>}
      </div>
    </div>
  )
}

// 설정 한 줄: 왼쪽 이름·설명, 오른쪽 조작
function Row({ title, hint, children }: { title: string; hint?: string; children: ReactNode }) {
  return (
    <div className="lrow">
      <div className="grow col" style={{ gap: 2 }}>
        <span className="b">{title}</span>
        {hint && <span className="hint">{hint}</span>}
      </div>
      {children}
    </div>
  )
}
