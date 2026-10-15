import type { ReactNode } from 'react'
import { Avatar } from '../../components/Avatar'
import { Seg, Toggle } from '../../components/Controls'
import type { Profile } from '../../components/ProfileCard'

export type AppTab = 'account' | 'notify' | 'look' | 'app'

const TABS: [AppTab, string][] = [
  ['account', '내 계정'],
  ['notify', '알림'],
  ['look', '화면'],
  ['app', '앱'],
]

export type FontSize = 's' | 'm' | 'l'

// 지금 켜져 있는 값. 이 컴퓨터에만 저장되고 서버에는 안 간다.
export type AppPrefs = {
  dnd: boolean // 방해 금지
  notify: 'all' | 'mention' | 'off'
  sound: boolean
  shakeWindow: boolean // 흔들기 받으면 창이 흔들림
  fontSize: FontSize
  reduceMotion: boolean
  clock: '12' | '24'
  autostart: boolean
  closeToTray: boolean
}

export const DEFAULT_PREFS: AppPrefs = {
  dnd: false,
  notify: 'all',
  sound: true,
  shakeWindow: true,
  fontSize: 'm',
  reduceMotion: false,
  clock: '12',
  autostart: false,
  closeToTray: true,
}

type Props = {
  tab: AppTab
  me: Profile
  email: string // 구글 계정
  prefs?: AppPrefs
  version?: string
}

// 앱 설정: 아지트와 상관없이 이 앱 자체의 설정 (아지트 설정과 같은 틀).
export function AppSettingsView({ tab, me, email, prefs = DEFAULT_PREFS, version = '0.1.0' }: Props) {
  return (
    <div className="main settings">
      <nav className="snav" aria-label="앱 설정">
        <span className="snav-title">앱 설정</span>
        <span className="hint">이 컴퓨터에서만 바뀌어요</span>
        {TABS.map(([t, label]) => (
          <button key={t} type="button" className={`btn snav-btn${t === tab ? ' on' : ''}`} aria-current={t === tab ? 'page' : undefined}>
            {label}
          </button>
        ))}
        {/* 이 컴퓨터에서만 나간다. 아지트와 대화는 그대로 */}
        <button type="button" className="btn snav-btn leave">로그아웃</button>
      </nav>

      <div className="scroll page settings-body app-set">
        {tab === 'account' && <>
          <h1 className="page-title sm">내 계정</h1>
          <div className="card row" style={{ gap: 14 }}>
            <Avatar name={me.nickname || '?'} color={me.color} size={56} src={me.avatarSrc} />
            <div className="grow col" style={{ gap: 2 }}>
              <span className="disp" style={{ fontSize: 20 }}>{me.nickname}</span>
              <span className="mono b" style={{ fontSize: 12, color: me.color }}>@{me.name}</span>
            </div>
            <button type="button" className="btn s sec">프로필 고치기</button>
          </div>
          <Row title="구글 계정" hint={`${email} 로 들어와 있어요`}>
            <span className="tag lg">연결됨</span>
          </Row>
        </>}

        {tab === 'notify' && <>
          <h1 className="page-title sm">알림</h1>
          <Row title="방해 금지" hint="흔들려도 창이 안 흔들리고 알림만 쌓여요">
            <Toggle on={prefs.dnd} label="방해 금지" />
          </Row>
          <Row title="새 메시지 알림" hint="멘션만 고르면 나를 부를 때와 흔들기만 알려요">
            <Seg label="새 메시지 알림" value={prefs.notify} options={[['all', '전부'], ['mention', '멘션만'], ['off', '끄기']]} />
          </Row>
          <Row title="알림 소리">
            <Toggle on={prefs.sound} label="알림 소리" />
          </Row>
          <Row title="흔들기 받으면 창 흔들기" hint="끄면 알림만 와요">
            <Toggle on={prefs.shakeWindow} label="흔들기 받으면 창 흔들기" />
          </Row>
        </>}

        {tab === 'look' && <>
          <h1 className="page-title sm">화면</h1>
          <Row title="글자 크기" hint="채팅과 본문 글자가 같이 바뀌어요">
            <Seg label="글자 크기" value={prefs.fontSize} options={[['s', '작게'], ['m', '보통'], ['l', '크게']]} />
          </Row>
          <div className={`card2 app-set-sample fs-${prefs.fontSize}`} aria-label="글자 크기 미리보기">
            <span className="app-set-bubble">오늘 마크 서버 열림?</span>
            <span className="app-set-bubble mine">열림 ㅋㅋ 근데 엔드는 아직 가지 마라</span>
          </div>
          <Row title="시간 표시">
            <Seg label="시간 표시" value={prefs.clock} options={[['12', '오후 2:01'], ['24', '14:01']]} />
          </Row>
          <Row title="움직임 줄이기" hint="창 흔들림, 반짝임 같은 움직임을 꺼요">
            <Toggle on={prefs.reduceMotion} label="움직임 줄이기" />
          </Row>
        </>}

        {tab === 'app' && <>
          <h1 className="page-title sm">앱</h1>
          <Row title="컴퓨터 켜면 같이 켜기">
            <Toggle on={prefs.autostart} label="컴퓨터 켜면 같이 켜기" />
          </Row>
          <Row title="창 닫으면 숨기기" hint="X 를 눌러도 꺼지지 않고 알림을 계속 받아요">
            <Toggle on={prefs.closeToTray} label="창 닫으면 숨기기" />
          </Row>
          <Row title="버전" hint="새 버전이 나오면 여기서 알려 줘요">
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
