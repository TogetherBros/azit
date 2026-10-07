// 화면 확인용 가짜 데이터 (프로토타입 내용 그대로).
// 서버와 연결되면 이 폴더는 통째로 지운다.
import type {
  Azit, ChatRow, GameRecord, InviteCode, LogEntry, Member, Notification,
  OpenGame, ReactOption, Room, Seat, Stats, Sticker, Title,
} from '../api/types'

/* ---------- 사람 · 아지트 · 방 ---------- */

export const ME: Member = { id: '4n5rud', nick: '4n5rud', color: '#D4FF3A', title: '눈치 백단', online: true, via: '아지트를 만듦', joinedAgo: '3주 전' }

export const MEMBERS: Member[] = [
  ME,
  { id: 'kim', nick: '김환성', color: '#FF5FA8', online: true, via: 'MK7Q2X로 들어옴', joinedAgo: '3주 전' },
  { id: 'kang', nick: '강백호', color: '#7BC4FF', online: true, title: '포커페이스', via: 'MK7Q2X로 들어옴', joinedAgo: '2주 전' },
  { id: 'ryu', nick: '류승엽', color: '#FF8A3D', title: '눈치 백단', via: 'H3BR9T로 들어옴', joinedAgo: '2주 전' },
  { id: 'yoon', nick: '윤재한', color: '#B79CFF', dnd: true, via: 'H3BR9T로 들어옴', joinedAgo: '10일 전' },
  { id: 'chu', nick: '추승주', color: '#F4F4F6', via: 'MK7Q2X로 들어옴', joinedAgo: '방금' },
]
export const [, KIM, KANG, RYU, YOON, CHU] = MEMBERS

export const AZITS: Azit[] = [
  {
    id: 'az1',
    name: '의성읍시회축구단이낚시하다낚은강아지를키워봅세',
    short: '의성',
    color: '#D4FF3A',
    desc: '낚시하다 강아지 낚은 사람들의 모임',
    memberCount: 14,
    unread: 0,
  },
  { id: 'az2', name: '마크 원정대', short: '마크', color: '#7BC4FF', memberCount: 6, unread: 12 },
  { id: 'az3', name: '동네 친구들', short: '동네', color: '#FF8A3D', memberCount: 9, unread: 0 },
]

export const ROOMS: Room[] = [
  { id: 'r1', name: '다같이', landing: true, preview: '김환성: @강백호 @4n5rud 점심 내기 누가 쏨', lastAgo: '방금', unread: 0 },
  { id: 'r2', name: '게임방', preview: '강백호: 라이어 할 사람', lastAgo: '3분 전', unread: 4, liar: 'lobby' },
  { id: 'r3', name: '짤방', preview: '류승엽: 사진', lastAgo: '1시간 전', unread: 0 },
  { id: 'r4', name: '마크 원정대', preview: '추승주: 성 다 지음 보러 와라', lastAgo: '2시간 전', unread: 23 },
  { id: 'r5', name: '건축 자랑', preview: '아직 조용해요', unread: 0 },
]
export const ROOM_OPTIONS = ROOMS.map((r) => ({ id: r.id, name: r.name, busy: r.id === 'r2' }))

/* ---------- 짤 · 맞장구 ---------- */

// 짤 그림 대신 쓰는 단색 그림 (진짜 짤은 서버의 media 에서 온다)
function stickerSvg(text: string, bg: string, fg = '#0E0E12'): string {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 120 120"><rect width="120" height="120" rx="28" fill="${bg}"/><text x="60" y="72" font-size="30" font-family="sans-serif" font-weight="800" text-anchor="middle" fill="${fg}">${text}</text></svg>`
  return `data:image/svg+xml,${encodeURIComponent(svg)}`
}

export const STICKERS: Sticker[] = [
  { id: 'st1', name: '정색', src: stickerSvg('정색', '#7BC4FF'), by: '강백호', ago: '2주 전', inReacts: true },
  { id: 'st2', name: '현타', src: stickerSvg('현타', '#B79CFF'), by: '류승엽', ago: '1주 전' },
  { id: 'st3', name: '급식', src: stickerSvg('급식', '#FF8A3D'), by: '4n5rud', ago: '5일 전' },
  { id: 'st4', name: '엔더각', src: stickerSvg('엔더', '#D4FF3A'), by: '추승주', ago: '어제' },
  { id: 'st5', name: '킹받네', src: stickerSvg('킹받', '#FF5FA8'), by: '김환성', ago: '방금' },
]

export const REACTS: ReactOption[] = [
  { id: 'rx1', text: 'ㅋㅋ' },
  { id: 'rx2', text: '헐' },
  { id: 'rx3', img: STICKERS[0].src, name: '정색' },
  { id: 'rx4', text: '킹받네' },
  { id: 'rx5', text: '최고' },
]

const PHOTO = (() => {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 240 180"><defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#7BC4FF"/><stop offset="1" stop-color="#B79CFF"/></linearGradient></defs><rect width="240" height="180" fill="url(#g)"/><path d="M0 140l60-50 50 40 40-30 90 60v20H0z" fill="#2A2A34" opacity=".6"/><circle cx="185" cy="45" r="18" fill="#F4F4F6" opacity=".8"/></svg>`
  return `data:image/svg+xml,${encodeURIComponent(svg)}`
})()
export const PHOTO_SRC = PHOTO

/* ---------- 채팅 ---------- */

export const CHAT_ROWS: ChatRow[] = [
  { type: 'date', id: 'd1', label: '10월 7일 화요일' },
  { type: 'msg', msg: { kind: 'sys', id: 's1', text: '추승주가 MK7Q2X로 들어왔어요', action: '환영 흔들기' } },
  {
    type: 'msg',
    msg: { kind: 'text', id: 'm1', author: KANG, parts: [{ text: '오늘 마크 서버 열림?' }], time: '오후 1:58' },
    showTime: true,
  },
  {
    type: 'msg',
    msg: {
      kind: 'text',
      id: 'm2',
      author: ME,
      mine: true,
      parts: [{ text: '열림 ㅋㅋ 근데 엔드는 아직 가지 마라' }],
      time: '오후 2:01',
      reactions: [{ label: 'ㅋㅋ', count: 3 }, { label: '정색', img: STICKERS[0].src, count: 1 }],
    },
    showTime: true,
  },
  {
    type: 'msg',
    msg: { kind: 'text', id: 'm3', author: RYU, parts: [{ text: '이거 봐라 ㅋㅋㅋ' }], time: '오후 2:05' },
  },
  {
    type: 'msg',
    msg: { kind: 'image', id: 'm4', author: RYU, src: PHOTO, time: '오후 2:05', reactions: [{ label: '헐', count: 2, on: true }] },
    cont: true,
    showTime: true,
  },
  {
    type: 'msg',
    msg: { kind: 'sticker', id: 'm5', author: CHU, src: STICKERS[4].src, name: '킹받네', time: '오후 2:06' },
    showTime: true,
  },
  {
    type: 'msg',
    msg: {
      kind: 'text',
      id: 'm6',
      author: KIM,
      parts: [{ mention: '강백호', shook: true }, { text: ' ' }, { mention: '4n5rud', mine: true, shook: true }, { text: ' 점심 내기 누가 쏨' }],
      time: '오후 2:10',
      shookMe: true,
    },
  },
  {
    type: 'msg',
    msg: {
      kind: 'roulette',
      id: 'm7',
      author: KIM,
      time: '오후 2:10',
      roulette: { title: '점심 쏘기', options: ['김환성', '강백호', '4n5rud', '류승엽'], result: '강백호', spinsLeft: 2 },
    },
    cont: true,
    showTime: true,
  },
  { type: 'msg', msg: { kind: 'sys', id: 's2', text: '강백호가 흔들었어요 · 4n5rud', shake: true, action: '되흔들기' } },
  {
    type: 'msg',
    msg: {
      kind: 'poll',
      id: 'm8',
      author: ME,
      mine: true,
      time: '오후 2:14',
      poll: {
        question: '토요일 몇 시?',
        options: [{ label: '오후 2시', count: 3, mine: true }, { label: '오후 4시', count: 1 }, { label: '저녁', count: 2 }],
        total: 6,
        canClose: true,
      },
    },
    showTime: true,
  },
  {
    type: 'msg',
    msg: {
      kind: 'game',
      id: 'm9',
      author: KANG,
      time: '오후 2:20',
      game: { status: 'lobby', players: 3, info: '음식 · 일반 라이어 · 설명 2바퀴 · 강백호가 엶' },
    },
    showTime: true,
  },
  {
    type: 'msg',
    msg: {
      kind: 'game',
      id: 'm10',
      author: KANG,
      time: '오후 2:41',
      game: { status: 'done', players: 5, info: '음식 · 일반 라이어 · 설명 2바퀴 · 강백호가 엶', result: '시민 승리 · 라이어 류승엽 · 제시어 떡볶이', joined: true },
    },
    showTime: true,
  },
  {
    type: 'msg',
    msg: { kind: 'text', id: 'm11', author: CHU, parts: [], time: '오후 2:42', deleted: true },
    showTime: true,
  },
]

/* ---------- 놀이터 · 라이어 ---------- */

export const MY_STATS: Stats = { games: 7, catches: 3, liarSurvive: 1, liarGames: 2, snipes: 0, citizenWins: 4, shakes: 12 }
export const KIM_STATS: Stats = { games: 4, catches: 1, liarSurvive: 0, liarGames: 1, snipes: 0, citizenWins: 2, shakes: 5 }

export const OPEN_GAMES: OpenGame[] = [
  { id: 'g1', room: '게임방', status: 'lobby', players: 3, topic: '음식' },
  { id: 'g2', room: '다같이', status: 'playing', players: 5, topic: '동물', joined: true },
]

export const RECORDS: GameRecord[] = [
  { id: 'gr1', winner: 'citizen', text: '게임방 · 제시어 떡볶이 · 라이어 류승엽', ago: '10분 전' },
  { id: 'gr2', winner: 'liar', text: '다같이 · 제시어 펭귄 · 라이어 강백호', ago: '어제' },
]

export const RANKING = [
  { member: ME, wins: 5, me: true },
  { member: KANG, wins: 4 },
  { member: KIM, wins: 2 },
  { member: RYU, wins: 1 },
]

const seat = (m: Member, extra: Partial<Seat> = {}): Seat => ({ member: m, me: m.id === ME.id, ...extra })

export const SEATS = {
  lobby: [seat(KANG, { host: true }), seat(ME), seat(KIM)],
  describe: [
    seat(KANG, { host: true, desc: '빨간색임' }),
    seat(ME, { desc: '학교 앞에서 팜' }),
    seat(KIM, { talking: true }),
    seat(RYU),
    seat(CHU),
  ],
  vote: [
    seat(KANG, { host: true, desc: '빨간색임', voted: true }),
    seat(ME, { desc: '학교 앞에서 팜', voted: true }),
    seat(KIM, { desc: '맵다', voted: true }),
    seat(RYU, { desc: '음... 좋아함', myVote: true }),
    seat(CHU, { desc: '다들 알 거임' }),
  ],
  done: [
    seat(KANG, { host: true, desc: '빨간색임', votes: 1 }),
    seat(ME, { desc: '학교 앞에서 팜' }),
    seat(KIM, { desc: '맵다' }),
    seat(RYU, { desc: '음... 좋아함', votes: 4, liar: true }),
    seat(CHU, { desc: '다들 알 거임' }),
  ],
}

export const GAME_CHAT = [
  { text: '강백호가 라이어 판을 열었어요' },
  { who: '강백호', color: KANG.color, text: '빨리 들어와' },
  { who: '김환성', color: KIM.color, text: '류승엽 수상함 ㅋㅋ' },
  { who: '류승엽', color: RYU.color, text: '아님' },
]

/* ---------- 알림 · 칭호 ---------- */

export const NOTIFS: Notification[] = [
  { id: 'n1', kind: 'shake', text: '김환성이 다같이에서 흔들었어요', azit: '의성', ago: '방금' },
  { id: 'n2', kind: 'game', text: '강백호가 게임방에 라이어 판을 열었어요', azit: '의성', ago: '3분 전' },
  { id: 'n3', kind: 'join', text: '추승주가 MK7Q2X로 들어왔어요', azit: '의성', ago: '1시간 전', read: true },
  { id: 'n4', kind: 'title', text: '새 칭호 · 눈치 백단', azit: '의성', ago: '어제', read: true },
  { id: 'n5', kind: 'kick', text: '옛날 모임에서 내보내졌어요 · 새 코드로 다시 들어올 수 있어요', azit: '옛날', ago: '3일 전', read: true, gone: true },
]

export const TITLES: Title[] = [
  { name: '라이어 입문', rarity: '일반', cond: '라이어게임 1판', unlocked: true, progress: 1 },
  { name: '눈치 백단', rarity: '희귀', cond: '시민으로 라이어 3번 검거', unlocked: true, progress: 1 },
  { name: '포커페이스', rarity: '희귀', cond: '라이어로 1번 살아남기', unlocked: true, progress: 1 },
  { name: '제시어 저격수', rarity: '전설', cond: '잡힌 뒤 제시어 맞히기', unlocked: false, progress: 0 },
  { name: '억울한 사람', rarity: '일반', cond: '시민인데 2번 지목당하기', unlocked: false, progress: 0.5 },
  { name: '판돌이', rarity: '일반', cond: '라이어 판 5번 하기', unlocked: true, progress: 1 },
  { name: '흔들기 장인', rarity: '일반', cond: '10번 흔들기', unlocked: true, progress: 1 },
]

/* ---------- 설정 ---------- */

export const CODES: InviteCode[] = [
  { code: 'MK7Q2X', madeBy: '4n5rud가', state: 'live', left: '6일 남음', uses: 3, max: 0, joined: ['김환성', '강백호', '추승주'] },
  { code: 'H3BR9T', madeBy: '강백호가', state: 'live', left: '20시간 남음', uses: 2, max: 5, joined: ['류승엽', '윤재한'] },
  { code: 'PL0W4E', madeBy: '김환성이', state: 'revoked', uses: 0, max: 1, joined: [] },
  { code: 'ZX81QA', madeBy: '4n5rud가', state: 'expired', uses: 1, max: 5, joined: ['석지원'] },
]

export const LOGS: LogEntry[] = [
  { id: 'l1', ago: '방금', by: KIM, type: '짤', text: '킹받네 짤 올림', undoable: true },
  { id: 'l2', ago: '10분 전', by: KANG, type: '방', text: '테스트ㅋㅋ 방 보관', undoable: true },
  { id: 'l3', ago: '1시간 전', by: ME, type: '코드', text: '초대 코드 PL0W4E 폐기' },
  { id: 'l4', ago: '어제', by: RYU, type: '설정', text: '아지트 이름을 바꿈', undoneBy: '4n5rud가' },
  { id: 'l5', ago: '3일 전', by: YOON, type: '맞장구', text: '맞장구 순서 바꿈', undoable: true },
]
