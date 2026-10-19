// 서버 메시지(proto.MessageResponse) → 채팅 화면 줄(ChatRow).
// 날짜 구분선, 같은 사람 이어 보내기(cont), 묶음 끝 시간(showTime), "여기까지 읽었어요" 선을 여기서 정한다.
import type { MessageRes, ReactionRes } from '../api/server'
import type { ChatRow, MessagePart, Reaction, UserMessage } from '../api/types'
import { getPrefs } from './prefs'

const GROUP_GAP = 5 * 60 * 1000 // 같은 사람이 5분 안에 이어 보내면 한 묶음

// 아직 서버에 안 간 내 메시지 (보내는 중 · 못 보냄)
// 사진·짤은 kind 와 미리 보여 줄 src 를 같이 들고 있다 (사진은 올린 뒤 photoKey 가 생긴다)
export type Outgoing = {
  tempId: string
  body: string
  at: number
  failed?: boolean
  shakeAll?: boolean
  shakeIds?: number[] // @로 부른 사람
  kind?: 'text' | 'photo' | 'meme'
  src?: string // 보내는 동안 보여 줄 그림 (사진은 로컬 미리보기, 짤은 짤 링크)
  name?: string // 짤 이름
  memeId?: number
  photoKey?: string
  file?: File // 아직 못 올린 사진 (다시 보내기 때 다시 올린다)
  replyToId?: number // 답장이면 원래 메시지 id
}

type Content = { kind: 'text'; parts: MessagePart[] } | { kind: 'image'; src: string } | { kind: 'sticker'; src: string; name: string }

// 본문 속 @이름 하나. 서버는 본문을 파싱하지 않아서, 부른 사람(shake_ids)의 지금 닉네임으로 본문에서 찾아 칩으로 바꾼다
type Mention = { text: string; mine: boolean }

function content(kind: string, body: string, src: string, name: string, mentions: Mention[]): Content {
  if (kind === 'photo') return { kind: 'image', src }
  if (kind === 'meme') return { kind: 'sticker', src, name }
  return { kind: 'text', parts: toParts(body, mentions) }
}

export type Me = { id: number; nick: string; color: string; src?: string }

// nicks: 아지트 멤버 account id → 아지트 닉네임 (@이름 칩을 찾을 때 쓴다)
export function toRows(messages: MessageRes[], outgoing: Outgoing[], me: Me, lastReadId: number, nicks: Map<number, string> = new Map()): ChatRow[] {
  const items: { at: Date; msg: UserMessage; senderId: number; id: number }[] = []

  const mentionsOf = (shakeAll: boolean, shakeIds: number[]): Mention[] => {
    const out: Mention[] = shakeAll ? [{ text: '전체', mine: true }] : []
    for (const id of shakeIds) {
      const nick = nicks.get(id)
      if (nick) out.push({ text: nick, mine: id === me.id })
    }
    return out
  }

  // 답장 인용: 불러온 메시지에서 원래 것을 찾는다 (더 위라 아직 안 불러왔으면 안내 문구)
  const byId = new Map(messages.map((m) => [m.message_id, m]))
  const quote = (id: number | null | undefined) => {
    if (!id) return undefined
    const o = byId.get(id)
    if (!o) return { id: String(id), nick: '이전 메시지', text: '눌러서 불러오기' }
    return { id: String(id), nick: o.sender?.nickname ?? '알 수 없음', text: preview(o) }
  }

  for (const m of messages) {
    if (m.kind === 'system' || !m.sender) continue
    const mine = m.sender.account_id === me.id
    items.push({
      at: new Date(m.created_at),
      senderId: m.sender.account_id,
      id: m.message_id,
      msg: {
        ...content(m.kind, m.body, m.media_url, '짤', mentionsOf(m.shake_all, m.shake_ids)),
        id: String(m.message_id),
        author: { id: String(m.sender.account_id), nick: m.sender.nickname, color: m.sender.color, src: m.sender.avatar_url || undefined },
        time: '',
        mine,
        deleted: m.deleted,
        edited: !!m.edited_at && !m.deleted,
        shookMe: !mine && (m.shake_all || m.shake_ids.includes(me.id)),
        reactions: m.reactions.map(toReaction),
        reply: quote(m.reply_to_id),
      },
    })
  }

  // 보낸 메시지 응답이 목록보다 먼저 왔을 수 있어서, 이미 목록에 있는 temp_id 는 뺀다
  const sent = new Set(messages.map((m) => m.temp_id))
  for (const o of outgoing) {
    if (sent.has(o.tempId)) continue
    items.push({
      at: new Date(o.at),
      senderId: me.id,
      id: Number.MAX_SAFE_INTEGER,
      msg: {
        ...content(o.kind ?? 'text', o.body, o.src ?? '', o.name ?? '짤', mentionsOf(!!o.shakeAll, o.shakeIds ?? [])),
        id: o.tempId,
        author: { id: String(me.id), nick: me.nick, color: me.color, src: me.src },
        time: '',
        mine: true,
        pending: !o.failed,
        failed: o.failed,
        reply: quote(o.replyToId),
      },
    })
  }

  const rows: ChatRow[] = []
  let readDrawn = lastReadId === 0 // 기록이 없으면 처음부터 다 안 읽은 거라 선을 안 긋는다
  for (let i = 0; i < items.length; i++) {
    const it = items[i]
    const prev = items[i - 1]
    const next = items[i + 1]

    if (!prev || !sameDay(prev.at, it.at)) {
      rows.push({ type: 'date', id: `d${it.at.toDateString()}`, label: dayLabel(it.at) })
    }
    // 처음으로 남이 보낸 안 읽은 메시지 위에 선
    if (!readDrawn && it.id > lastReadId && !it.msg.mine) {
      rows.push({ type: 'read', id: 'read' })
      readDrawn = true
    }

    const cont = !!prev && prev.senderId === it.senderId && sameDay(prev.at, it.at) && it.at.getTime() - prev.at.getTime() < GROUP_GAP && rows[rows.length - 1].type === 'msg'
    const endsGroup = !next || next.senderId !== it.senderId || next.at.getTime() - it.at.getTime() >= GROUP_GAP || !sameDay(next.at, it.at)
    rows.push({ type: 'msg', msg: { ...it.msg, time: clock(it.at) }, cont, showTime: endsGroup })
  }
  return rows
}

// 메시지 한 줄 요약 (답장 인용, 답장 쓰는 중 표시)
export function preview(m: MessageRes): string {
  if (m.deleted) return '지운 메시지'
  if (m.kind === 'photo') return '사진'
  if (m.kind === 'meme') return '짤'
  return m.body.replace(/\s+/g, ' ').slice(0, 80)
}

function toReaction(r: ReactionRes): Reaction {
  return r.meme_id
    ? { label: `짤 ${r.meme_id}`, img: r.meme_url, count: r.count, on: r.mine, memeId: r.meme_id }
    : { label: r.text, count: r.count, on: r.mine, text: r.text }
}

// 본문의 @이름(부른 사람만)·@전체 를 흔들기 칩으로 바꾼다. 긴 이름부터 맞춰서 "@강백" 이 "@강백호" 를 먼저 먹지 않게.
function toParts(body: string, mentions: Mention[]): MessagePart[] {
  const found = mentions.filter((m) => body.includes(`@${m.text}`)).sort((a, b) => b.text.length - a.text.length)
  if (found.length === 0) return [{ text: body }]
  const byText = new Map(found.map((m) => [`@${m.text}`, m]))
  const re = new RegExp(`(${found.map((m) => `@${m.text}`.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|')})`)
  return body.split(re).filter(Boolean).map((t) => {
    const m = byText.get(t)
    return m ? { mention: m.text, mine: m.mine, shook: true } : { text: t }
  })
}

// 입력 중인 @검색어. 글자 커서 바로 앞이 "@검색어" 면 그 검색어, 아니면 null
export function mentionQuery(text: string, caret: number): string | null {
  const m = /(?:^|\s)@([^\s@]{0,16})$/.exec(text.slice(0, caret))
  return m ? m[1] : null
}

const sameDay = (a: Date, b: Date) => a.toDateString() === b.toDateString()

const DAYS = ['일', '월', '화', '수', '목', '금', '토']

function dayLabel(d: Date): string {
  const today = new Date()
  const yest = new Date(today.getTime() - 86400000)
  if (sameDay(d, today)) return '오늘'
  if (sameDay(d, yest)) return '어제'
  const y = d.getFullYear() === today.getFullYear() ? '' : `${d.getFullYear()}년 `
  return `${y}${d.getMonth() + 1}월 ${d.getDate()}일 ${DAYS[d.getDay()]}요일`
}

// "오후 2:05" (앱 설정에서 24시간을 고르면 "14:05")
export function clock(d: Date): string {
  const h = d.getHours()
  const m = String(d.getMinutes()).padStart(2, '0')
  if (getPrefs().clock === '24') return `${String(h).padStart(2, '0')}:${m}`
  return `${h < 12 ? '오전' : '오후'} ${h % 12 || 12}:${m}`
}

// "방금", "3분 전", "어제", "3일 전"
export function ago(iso: string): string {
  const s = (Date.now() - new Date(iso).getTime()) / 1000
  if (s < 60) return '방금'
  if (s < 3600) return `${Math.floor(s / 60)}분 전`
  if (s < 86400) return `${Math.floor(s / 3600)}시간 전`
  if (s < 172800) return '어제'
  if (s < 86400 * 30) return `${Math.floor(s / 86400)}일 전`
  return `${Math.floor(s / (86400 * 30))}달 전`
}

// 초대 코드 남은 시간 "6일 남음", "40분 남음", "기한 없음"
export function left(expiresAt: string | null): string {
  if (!expiresAt) return '기한 없음'
  const s = (new Date(expiresAt).getTime() - Date.now()) / 1000
  if (s <= 0) return '만료'
  if (s < 3600) return `${Math.max(1, Math.ceil(s / 60))}분 남음`
  if (s < 86400) return `${Math.ceil(s / 3600)}시간 남음`
  return `${Math.ceil(s / 86400)}일 남음`
}

// 앱이 메시지마다 한 번 만드는 id (다시 보내도 같은 값 → 서버가 중복 저장 안 함)
export function newTempId(): string {
  return crypto.randomUUID()
}
