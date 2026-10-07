// 화면이 받는 데이터 모양.
// 지금은 가짜 데이터(mock/)용으로 손으로 적었다. 나중에 desktop/internal/bridge 가
// proto 구조체를 돌려주면 Wails 가 생성한 타입(frontend/bindings/)으로 바꾼다.
//
// 시간("3분 전", "오후 2:10")처럼 표시용 글자는 이미 만들어진 상태로 받는다고 가정한다.

export type Member = {
  id: string
  nick: string
  color: string
  title?: string // 장착한 칭호
  online?: boolean
  dnd?: boolean // 방해 금지
  via?: string // "MK7Q2X로 들어옴", "아지트를 만듦"
  joinedAgo?: string
}

export type Azit = {
  id: string
  name: string
  short: string // 레일 네모에 들어가는 3자 이하
  color: string
  desc?: string
  memberCount: number
  unread: number
}

export type Room = {
  id: string
  name: string
  landing?: boolean // 새 멤버가 처음 들어오는 방
  preview: string // "강백호: 오늘 마크 서버 열림?"
  lastAgo?: string // "3분 전"
  unread: number
  liar?: 'lobby' | 'playing' // 열린 라이어 판
}

/* ---------- 메시지 ---------- */

// 텍스트 본문은 글자와 @이름 조각으로 나뉜다.
export type MessagePart =
  | { text: string }
  | { mention: string; mine?: boolean; shook?: boolean }

export type Reaction = { label: string; img?: string; count: number; on?: boolean }

// 이 아지트의 맞장구 하나: 글자(4자까지) 또는 짤
export type ReactOption = { id: string; text?: string; img?: string; name?: string }

export type LiarCard = {
  status: 'lobby' | 'playing' | 'done' | 'cancel'
  players: number
  info: string // "음식 · 일반 라이어 · 설명 2바퀴 · 강백호가 엶"
  result?: string // "시민 승리 · 라이어 류승엽 · 제시어 떡볶이"
  joined?: boolean
}

export type RouletteCard = { title?: string; options: string[]; result: string; spinsLeft: number }

export type PollCard = {
  question: string
  options: { label: string; count: number; mine?: boolean }[]
  total: number
  closed?: boolean
  canClose?: boolean // 내가 올린 투표
}

// 사람이 보낸 메시지의 공통 부분 + 종류별 본문
type UserBase = {
  id: string
  author: Member
  time: string
  mine?: boolean
  deleted?: boolean
  shookMe?: boolean // 이 메시지로 내가 흔들렸는지
  reactions?: Reaction[]
}

export type UserMessage = UserBase & (
  | { kind: 'text'; parts: MessagePart[] }
  | { kind: 'image'; src: string }
  | { kind: 'sticker'; src: string; name: string }
  | { kind: 'game'; game: LiarCard }
  | { kind: 'roulette'; roulette: RouletteCard }
  | { kind: 'poll'; poll: PollCard }
)

export type SystemMessage = {
  kind: 'sys'
  id: string
  text: string
  shake?: boolean // 흔들기 알림이면 핑크 아이콘
  action?: string // "되흔들기", "환영 흔들기", "결과"
}

export type ChatMessage = UserMessage | SystemMessage

// 채팅 목록의 한 줄: 날짜 구분선 또는 메시지.
// cont = 같은 사람이 이어 보낸 것(아바타 생략), showTime = 묶음의 마지막이라 시간 표시.
export type ChatRow =
  | { type: 'date'; id: string; label: string }
  | { type: 'msg'; msg: ChatMessage; cont?: boolean; showTime?: boolean }

/* ---------- 짤 · 알림 · 칭호 ---------- */

export type Sticker = { id: string; name: string; src: string; by: string; ago: string; inReacts?: boolean }

export type NotifKind = 'shake' | 'game' | 'join' | 'kick' | 'title'
export type Notification = { id: string; kind: NotifKind; text: string; azit: string; ago: string; read?: boolean; gone?: boolean }

export type Rarity = '일반' | '희귀' | '전설'
export type Title = { name: string; rarity: Rarity; cond: string; unlocked: boolean; progress: number /* 0~1 */ }

export type Stats = { games: number; catches: number; liarSurvive: number; liarGames: number; snipes: number; citizenWins: number; shakes: number }

/* ---------- 설정 ---------- */

export type CodeState = 'live' | 'revoked' | 'expired' | 'used'
export type InviteCode = { code: string; madeBy: string; state: CodeState; left?: string; uses: number; max: number /* 0 = 제한 없음 */; joined: string[] }

export type LogType = '방' | '짤' | '맞장구' | '코드' | '멤버' | '설정'
export type LogEntry = { id: string; ago: string; by: Member; type: LogType; text: string; undoable?: boolean; undoneBy?: string }

/* ---------- 라이어게임 ---------- */

export type LiarPhase = 'lobby' | 'describe' | 'vote' | 'guess' | 'done'

export type Seat = {
  member: Member
  host?: boolean
  me?: boolean
  desc?: string // 마지막 설명
  talking?: boolean
  voted?: boolean
  myVote?: boolean // 내가 이 사람에게 투표함
  votes?: number // 받은 표 (guess/done 에서 공개)
  liar?: boolean // done 에서 공개
}

export type OpenGame = { id: string; room: string; status: 'lobby' | 'playing'; players: number; topic: string; joined?: boolean }
export type GameRecord = { id: string; winner: 'liar' | 'citizen'; text: string; ago: string }
