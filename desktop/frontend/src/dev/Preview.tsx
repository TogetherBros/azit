// 화면 미리보기 (개발용, JS 단계에서 진짜 화면 전환이 생기면 지운다)
// 주소 끝 #이름 으로 화면을 고른다. 예: #join-preview, #chat-plus
import { useEffect, useState, type ReactNode } from 'react'
import { Rail } from '../app/Rail'
import { Shell } from '../app/Shell'
import { Toast } from '../components/Toast'
import { ActivityView } from '../features/activity/ActivityView'
import { JoinScreen } from '../features/auth/JoinScreen'
import { LoginScreen } from '../features/auth/LoginScreen'
import { AzitCreatedScreen } from '../features/azit/AzitCreatedScreen'
import { CreateAzitScreen } from '../features/azit/CreateAzitScreen'
import { KickModal, MemberModal } from '../features/azit/MemberModals'
import { CodesTab } from '../features/azit/settings/CodesTab'
import { GeneralTab } from '../features/azit/settings/GeneralTab'
import { LeaveTab } from '../features/azit/settings/LeaveTab'
import { LogTab } from '../features/azit/settings/LogTab'
import { MembersTab } from '../features/azit/settings/MembersTab'
import { RoomsTab } from '../features/azit/settings/RoomsTab'
import { SettingsView, type SettingsTab } from '../features/azit/settings/SettingsView'
import { StickersTab } from '../features/azit/settings/StickersTab'
import { DeleteMessageModal, ImageModal, PollModal, RouletteModal } from '../features/chat/ChatModals'
import { ChatView } from '../features/chat/ChatView'
import { ShakeModal } from '../features/chat/ShakeModal'
import { MeView } from '../features/me/MeView'
import { HubView } from '../features/playground/HubView'
import { LiarGameView } from '../features/playground/LiarGameView'
import { OpenGameModal } from '../features/playground/OpenGameForm'
import { NewRoomModal } from '../features/rooms/NewRoomModal'
import { RoomInfoModal } from '../features/rooms/RoomInfoModal'
import { RoomList } from '../features/rooms/RoomList'
import * as M from '../mock/data'

const AZ = M.AZITS[0]
const ROOM = M.ROOMS[0]
const OTHERS = M.MEMBERS.filter((m) => m.id !== M.ME.id)

// 레일 + 방 목록 + 본문. 본문을 안 주면 채팅.
function App(props: {
  chat?: Partial<Parameters<typeof ChatView>[0]>
  body?: ReactNode
  active?: 'settings' | 'hub'
  view?: 'activity'
  overlay?: ReactNode
  shaking?: boolean
}) {
  return (
    <>
      <Shell shaking={props.shaking}>
        <Rail azits={M.AZITS} currentId={AZ.id} me={M.ME} notifCount={2} view={props.view} />
        <RoomList azit={AZ} rooms={M.ROOMS} currentRoomId={props.body ? undefined : ROOM.id} active={props.active} />
        {props.body ?? (
          <ChatView
            room={ROOM}
            memberCount={AZ.memberCount}
            rows={M.CHAT_ROWS}
            reactOptions={M.REACTS}
            {...props.chat}
            composer={{ stickers: M.STICKERS, liarLive: '모이는 중 3/8 · 들어가기', ...props.chat?.composer }}
          />
        )}
      </Shell>
      {props.overlay}
    </>
  )
}

const settings = (tab: SettingsTab, body: ReactNode) => () => (
  <App active="settings" body={<SettingsView tab={tab}>{body}</SettingsView>} />
)

const liar = (p: Partial<Parameters<typeof LiarGameView>[0]>) => () => (
  <App
    active="hub"
    body={
      <LiarGameView
        room="게임방"
        phase="lobby"
        seats={M.SEATS.lobby}
        topic="음식"
        mode="일반 라이어"
        player
        chat={M.GAME_CHAT}
        {...p}
      />
    }
  />
)

const SCREENS: Record<string, () => ReactNode> = {
  // 들어오기
  'login': () => <LoginScreen />,
  'login-wait': () => <LoginScreen waiting />,
  'login-error': () => <LoginScreen error="멈춘 계정이에요. 아지트 친구에게 물어봐 주세요" />,
  'join-code': () => <JoinScreen step="code" account="wjdans" code="MK7Q" />,
  'join-code-error': () => (
    <JoinScreen step="code" account="wjdans" code="PL0W4E" error="폐기된 코드예요. 친구에게 새 코드를 받아 주세요" triesLeft={3} />
  ),
  'join-preview': () => (
    <JoinScreen
      step="preview"
      account="wjdans"
      preview={{
        azitName: AZ.name, short: AZ.short, color: AZ.color, desc: AZ.desc,
        memberCount: AZ.memberCount, rooms: M.ROOMS.map((r) => r.name), from: '강백호의 코드', expires: '6일 남음',
      }}
    />
  ),
  'join-nick': () => <JoinScreen step="nick" account="wjdans" nick="문경" color="#7BC4FF" />,
  'create': () => <CreateAzitScreen name="동네 친구들" />,
  'created': () => <AzitCreatedScreen name="동네 친구들" short="동네" color="#FF5FA8" code="H3BR9T" />,

  // 채팅
  'chat': () => <App />,
  'chat-plus': () => <App chat={{ composer: { popup: 'plus' } }} />,
  'chat-mention': () => <App chat={{ composer: { draft: '@강', popup: 'mention', mentionCandidates: OTHERS.slice(0, 3) } }} />,
  'chat-sticker': () => <App chat={{ composer: { popup: 'sticker' } }} />,
  'chat-games': () => <App chat={{ composer: { popup: 'games' } }} />,
  'chat-reconnect': () => <App chat={{ reconnecting: true, composer: { draft: '보내다 끊긴 메시지' } }} />,
  'chat-empty': () => <App chat={{ room: M.ROOMS[4], rows: [] }} />,
  'chat-shake': () => <App shaking />,
  'toast': () => <App overlay={<Toast text="짤로 저장했어요" action="되돌리기" />} />,

  // 모달
  'modal-newroom': () => <App overlay={<NewRoomModal memberCount={AZ.memberCount} />} />,
  'modal-newroom-error': () => <App overlay={<NewRoomModal memberCount={AZ.memberCount} name="짤방" error="같은 이름의 방이 이미 있어요" />} />,
  'modal-roominfo': () => (
    <App overlay={<RoomInfoModal room={ROOM} madeBy="4n5rud가 만듦" messageCount={128} members={M.MEMBERS} meId={M.ME.id} />} />
  ),
  'modal-shake': () => <App overlay={<ShakeModal members={OTHERS} />} />,
  'modal-roulette': () => (
    <App overlay={<RouletteModal title="점심 쏘기" candidates={[...M.MEMBERS.map((m) => m.nick), '떡볶이']} selected={['4n5rud', '김환성', '강백호']} />} />
  ),
  'modal-poll': () => <App overlay={<PollModal question="토요일 몇 시?" options={['오후 2시', '오후 4시', '']} />} />,
  'modal-image': () => <App overlay={<ImageModal src={M.PHOTO_SRC} />} />,
  'modal-delete': () => <App overlay={<DeleteMessageModal />} />,
  'modal-member': () => <App overlay={<MemberModal member={M.KIM} stats={M.KIM_STATS} titleCount={2} roomName="다같이" />} />,
  'modal-member-me': () => <App overlay={<MemberModal member={M.ME} stats={M.MY_STATS} titleCount={5} self />} />,
  'modal-kick': () => <App overlay={<KickModal nick="류승엽" />} />,
  'modal-opengame': () => <App overlay={<OpenGameModal rooms={M.ROOM_OPTIONS} roomId="r3" />} />,

  // 놀이터 · 라이어
  'hub': () => (
    <App
      active="hub"
      body={<HubView openGames={M.OPEN_GAMES} records={M.RECORDS} ranking={M.RANKING} myStats={M.MY_STATS} rooms={M.ROOM_OPTIONS} roomId="r3" />}
    />
  ),
  'liar-lobby': liar({ host: true }),
  'liar-describe': liar({ phase: 'describe', seats: M.SEATS.describe, word: '떡볶이', revealed: true, round: '1/2바퀴', turn: '김환성' }),
  'liar-describe-me': liar({ phase: 'describe', seats: M.SEATS.describe, amLiar: true, revealed: true, myTurn: true, round: '1/2바퀴' }),
  'liar-hidden': liar({ phase: 'describe', seats: M.SEATS.describe, word: '떡볶이', revealed: false, round: '1/2바퀴', turn: '김환성' }),
  'liar-vote': liar({ phase: 'vote', seats: M.SEATS.vote, word: '떡볶이', revealed: true, votedCount: 3, myVote: '류승엽', host: true }),
  'liar-guess': liar({ phase: 'guess', seats: M.SEATS.done.map((s) => ({ ...s, liar: false })), word: '떡볶이', revealed: true, accused: '류승엽' }),
  'liar-done': liar({
    phase: 'done',
    seats: M.SEATS.done,
    result: { winner: 'citizen', liars: '류승엽', word: '떡볶이', note: '류승엽이 지목됨' },
  }),
  'liar-watch': liar({ phase: 'describe', seats: M.SEATS.describe, player: false, round: '1/2바퀴', turn: '김환성' }),

  // 알림 · 내 정보
  'activity': () => <App view="activity" body={<ActivityView items={M.NOTIFS} />} />,
  'activity-empty': () => <App view="activity" body={<ActivityView items={[]} filter="칭호" />} />,
  'me': () => (
    <App body={<MeView azitShort={AZ.short} me={M.ME} email="wjdans@gmail.com" stats={M.MY_STATS} titles={M.TITLES} />} />
  ),

  // 아지트 설정
  'set-general': settings('general', (
    <GeneralTab azit={AZ} rooms={M.ROOMS} landingId="r1" welcome lastChange="류승엽이 바꿈 · 어제" />
  )),
  'set-general-error': settings('general', (
    <GeneralTab azit={AZ} rooms={M.ROOMS} landingId="r1" welcome lastChange="" dirty error="이름은 10분에 한 번 바꿀 수 있어요 · 7분 뒤에" />
  )),
  'set-codes': settings('codes', <CodesTab codes={M.CODES} made={{ code: 'Q9WE2R', left: '24시간 남음', uses: '5번' }} />),
  'set-members': settings('members', <MembersTab members={M.MEMBERS} meId={M.ME.id} bans={[{ id: 'ex01', name: '권재현' }]} />),
  'set-rooms': settings('rooms', (
    <RoomsTab rooms={M.ROOMS} archived={['테스트ㅋㅋ']} selectedId="r3" selectedInfo="류승엽이 만듦 · 2주 전 · 메시지 54개" />
  )),
  'set-rooms-landing': settings('rooms', (
    <RoomsTab rooms={M.ROOMS} archived={[]} selectedId="r1" selectedInfo="4n5rud가 만듦 · 3주 전 · 메시지 128개" />
  )),
  'set-stickers': settings('stickers', <StickersTab reacts={M.REACTS} stickers={M.STICKERS} pickingSticker renamingId="st2" />),
  'set-log': settings('log', <LogTab entries={M.LOGS} />),
  'set-leave': settings('leave', <LeaveTab myMessages={342} myLiveCodes={1} presses={1} />),

  // 작게 보기
  'mini': () => (
    <Shell mini>
      <Rail azits={M.AZITS} currentId={AZ.id} me={M.ME} notifCount={2} />
      <ChatView mini room={ROOM} memberCount={AZ.memberCount} rows={M.CHAT_ROWS} reactOptions={M.REACTS} />
    </Shell>
  ),
  'mini-rooms': () => (
    <Shell mini>
      <Rail azits={M.AZITS} currentId={AZ.id} me={M.ME} notifCount={2} />
      <RoomList mini azit={AZ} rooms={M.ROOMS} />
    </Shell>
  ),
}

function useHash() {
  const [hash, setHash] = useState(() => location.hash.slice(1))
  useEffect(() => {
    const on = () => setHash(location.hash.slice(1))
    window.addEventListener('hashchange', on)
    return () => window.removeEventListener('hashchange', on)
  }, [])
  return hash
}

export function Preview() {
  const hash = useHash()
  const name = hash in SCREENS ? hash : 'chat'
  const isMini = name.startsWith('mini')

  return (
    <>
      {isMini
        ? <div style={{ height: '100%', width: 400, borderRight: '1px solid var(--line2)' }}>{SCREENS[name]()}</div>
        : SCREENS[name]()}
      <nav
        aria-label="미리보기 화면"
        style={{
          position: 'fixed', right: 8, bottom: 8, zIndex: 100, maxWidth: 420, display: 'flex', flexWrap: 'wrap', gap: 4,
          padding: 8, borderRadius: 12, background: 'rgba(0,0,0,.85)', fontSize: 11, opacity: 0.3,
        }}
        onMouseEnter={(e) => (e.currentTarget.style.opacity = '1')}
        onMouseLeave={(e) => (e.currentTarget.style.opacity = '0.3')}
      >
        {Object.keys(SCREENS).map((k) => (
          <a key={k} href={`#${k}`} style={{ color: k === name ? 'var(--lime)' : 'var(--mu)', textDecoration: 'none' }}>{k}</a>
        ))}
      </nav>
    </>
  )
}
