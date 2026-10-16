// 서버 메시지(proto.MessageResponse) → 채팅 화면 줄(ChatRow).
// 날짜 구분선, 같은 사람 이어 보내기(cont), 묶음 끝 시간(showTime), "여기까지 읽었어요" 선을 여기서 정한다.
import type { MessageRes, ReactionRes } from '../api/server'
import type { ChatRow, MessagePart, Reaction, UserMessage } from '../api/types'

const GROUP_GAP = 5 * 60 * 1000 // 같은 사람이 5분 안에 이어 보내면 한 묶음

// 아직 서버에 안 간 내 메시지 (보내는 중 · 못 보냄)
// 사진·짤은 kind 와 미리 보여 줄 src 를 같이 들고 있다 (사진은 올린 뒤 photoKey 가 생긴다)
export type Outgoing = {
  tempId: string
  body: string
  at: number
  failed?: boolean
  shakeAll?: boolean
  kind?: 'text' | 'photo' | 'meme'
  src?: string // 보내는 동안 보여 줄 그림 (사진은 로컬 미리보기, 짤은 짤 링크)
  name?: string // 짤 이름
  memeId?: number
  photoKey?: string
  file?: File // 아직 못 올린 사진 (다시 보내기 때 다시 올린다)
}

type Content = { kind: 'text'; parts: MessagePart[] } | { kind: 'image'; src: string } | { kind: 'sticker'; src: string; name: string }

function content(kind: string, body: string, shakeAll: boolean, src: string, name: string): Content {
  if (kind === 'photo') return { kind: 'image', src }
  if (kind === 'meme') return { kind: 'sticker', src, name }
  return { kind: 'text', parts: toParts(body, shakeAll) }
}

export type Me = { id: number; nick: string; color: string; src?: string }

export function toRows(messages: MessageRes[], outgoing: Outgoing[], me: Me, lastReadId: number): ChatRow[] {
  const items: { at: Date; msg: UserMessage; senderId: number; id: number }[] = []

  for (const m of messages) {
    if (m.kind === 'system' || !m.sender) continue
    const mine = m.sender.account_id === me.id
    items.push({
      at: new Date(m.created_at),
      senderId: m.sender.account_id,
      id: m.message_id,
      msg: {
        ...content(m.kind, m.body, m.shake_all, m.media_url, '짤'),
        id: String(m.message_id),
        author: { id: String(m.sender.account_id), nick: m.sender.nickname, color: m.sender.color, src: m.sender.avatar_url || undefined },
        time: '',
        mine,
        deleted: m.deleted,
        edited: !!m.edited_at && !m.deleted,
        shookMe: !mine && (m.shake_all || m.shake_ids.includes(me.id)),
        reactions: m.reactions.map(toReaction),
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
        ...content(o.kind ?? 'text', o.body, !!o.shakeAll, o.src ?? '', o.name ?? '짤'),
        id: o.tempId,
        author: { id: String(me.id), nick: me.nick, color: me.color, src: me.src },
        time: '',
        mine: true,
        pending: !o.failed,
        failed: o.failed,
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

function toReaction(r: ReactionRes): Reaction {
  return r.meme_id
    ? { label: `짤 ${r.meme_id}`, img: r.meme_url, count: r.count, on: r.mine, memeId: r.meme_id }
    : { label: r.text, count: r.count, on: r.mine, text: r.text }
}

// @전체 는 흔들기 칩으로 보여 준다
function toParts(body: string, shakeAll: boolean): MessagePart[] {
  if (!shakeAll || !body.includes('@전체')) return [{ text: body }]
  return body.split(/(@전체)/).filter(Boolean).map((t) => (t === '@전체' ? { mention: '전체', shook: true } : { text: t }))
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

// "오후 2:05"
export function clock(d: Date): string {
  const h = d.getHours()
  const m = String(d.getMinutes()).padStart(2, '0')
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
