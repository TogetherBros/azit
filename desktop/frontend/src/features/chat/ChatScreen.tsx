import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  addReaction, deleteMessage, editMessage, errorText, listMessages, markRead, removeReaction, saveMemeFromMessage, sendMessage, uploadFile,
  type MemberRes, type MemeRes, type MessageRes, type SendInput,
} from '../../api/server'
import type { Member, ReactOption, Reaction, Room, Sticker, UserMessage } from '../../api/types'
import { ago, mentionQuery, newTempId, preview, toRows, type Me, type Outgoing } from '../../lib/chat'
import { shrinkImage } from '../../lib/image'
import { onRealtime, onReconnect } from '../../lib/realtime'
import { MemberProfileModal } from '../azit/MemberProfileModal'
import { RoomInfoModal } from '../rooms/RoomInfoModal'
import { SaveMemeModal } from './SaveMemeModal'
import { ShakeModal } from './ShakeModal'
import { DeleteMessageModal, ImageModal } from './ChatModals'
import { ChatView } from './ChatView'
import type { ComposerPopup } from './Composer'

const ROLE_TAG = { super_admin: '만든 사람', admin: '관리자', member: '' } as const

// 사람 고르는 목록의 둘째 줄: "접속 중 · @name"
const sub = (m: MemberRes) => [m.online && '접속 중', m.name && `@${m.name}`].filter(Boolean).join(' · ')

const POLL_MS = 30000 // 새 메시지는 실시간(웹소켓)으로 온다. 이건 혹시 놓친 걸 메우는 느린 확인

type Props = {
  azitId: number
  room: Room
  memberCount: number
  me: Me
  reactOptions: ReactOption[]
  memes: MemeRes[] // + → 짤 에서 고를 아지트 짤
  members: MemberRes[] // @ 후보
  canDeleteOthers: boolean // 관리자
  onRead?: () => void // 읽음 처리했으니 방 목록 안 읽은 수를 다시 받으라고
  onShaken?: () => void // 나를 흔든 새 메시지가 왔다
  onMemeSaved?: () => void // 채팅 사진을 짤로 저장했으니 짤 목록을 다시 받으라고
}

// 방 하나의 채팅. 서버에서 메시지를 받아 ChatView 에 넘기고, 보내기·고치기·지우기·맞장구를 처리한다.
export function ChatScreen({ azitId, room, memberCount, me, reactOptions, memes, members, canDeleteOthers, onRead, onShaken, onMemeSaved }: Props) {
  const roomId = Number(room.id)
  const [messages, setMessages] = useState<MessageRes[]>([])
  const [hasMore, setHasMore] = useState(false)
  const [lastReadId, setLastReadId] = useState(0)
  const [loadingMore, setLoadingMore] = useState(false)
  const [outgoing, setOutgoing] = useState<Outgoing[]>([])
  const [draft, setDraft] = useState('')
  const [editing, setEditing] = useState<UserMessage>()
  const [replyTo, setReplyTo] = useState<number>() // 답장할 메시지 id
  const [error, setError] = useState<string>()
  const readSent = useRef(0)
  const seen = useRef(0) // 흔들기 알림을 이미 확인한 마지막 메시지 id

  // 받은 메시지를 지금 목록에 합친다 (같은 id 는 새 것으로 바꿔 맞장구·고침·지움을 반영)
  const merge = useCallback((incoming: MessageRes[]) => {
    setMessages((cur) => {
      const byId = new Map(cur.map((m) => [m.message_id, m]))
      for (const m of incoming) byId.set(m.message_id, m)
      return [...byId.values()].sort((a, b) => a.message_id - b.message_id)
    })
  }, [])

  // 방을 바꾸면 처음부터
  useEffect(() => {
    let stop = false
    setMessages([]); setOutgoing([]); setHasMore(false); setEditing(undefined); setReplyTo(undefined); setDraft(''); setError(undefined)
    readSent.current = 0
    seen.current = Number.MAX_SAFE_INTEGER // 처음 불러온 메시지로는 흔들지 않는다

    const first = async () => {
      try {
        const res = await listMessages(azitId, roomId)
        if (stop) return
        setMessages(res.messages)
        setHasMore(res.has_more)
        setLastReadId(res.last_read_id)
        seen.current = res.messages[res.messages.length - 1]?.message_id ?? 0
      } catch (e) {
        if (!stop) setError(errorText(e))
      }
    }
    void first()

    const timer = window.setInterval(async () => {
      try {
        const res = await listMessages(azitId, roomId)
        if (stop) return
        merge(res.messages)
        setError(undefined)
      } catch (e) {
        if (!stop) setError(errorText(e))
      }
    }, POLL_MS)
    return () => { stop = true; window.clearInterval(timer) }
  }, [azitId, roomId, merge])

  // 실시간: 이 방의 새 메시지는 바로 붙이고, 고침·지움·맞장구면 최신 목록을 다시 받는다.
  // 다시 연결됐을 때도 다시 받는다 (끊겨 있던 사이 놓친 것 메우기).
  useEffect(() => {
    const refresh = () => {
      listMessages(azitId, roomId).then((res) => merge(res.messages)).catch(() => {})
    }
    const offEvent = onRealtime((e) => {
      if (e.type === 'message.created' && e.data.room_id === roomId) merge([e.data])
      if (e.type === 'message.updated' && e.data.room_id === roomId) refresh()
    })
    const offOpen = onReconnect(refresh)
    return () => { offEvent(); offOpen() }
  }, [azitId, roomId, merge])

  // 새로 온 메시지에 내가 흔들렸으면 알린다
  useEffect(() => {
    const fresh = messages.filter((m) => m.message_id > seen.current)
    if (fresh.length === 0) return
    seen.current = fresh[fresh.length - 1].message_id
    if (fresh.some((m) => m.sender && m.sender.account_id !== me.id && (m.shake_all || m.shake_ids.includes(me.id)))) onShaken?.()
  }, [messages, me.id, onShaken])

  // 화면에 보이는 마지막 메시지까지 읽음 (창이 앞에 있을 때만)
  useEffect(() => {
    const lastId = messages[messages.length - 1]?.message_id ?? 0
    if (!lastId || lastId <= readSent.current || document.hidden) return
    readSent.current = lastId
    markRead(azitId, roomId, lastId).then(onRead).catch(() => { readSent.current = 0 })
  }, [messages, azitId, roomId, onRead])

  const more = async () => {
    const first = messages[0]
    if (!first) return
    setLoadingMore(true)
    try {
      const res = await listMessages(azitId, roomId, first.message_id)
      merge(res.messages)
      setHasMore(res.has_more)
    } catch (e) {
      setError(errorText(e))
    } finally {
      setLoadingMore(false)
    }
  }

  const deliver = async (o: Outgoing) => {
    setOutgoing((os) => [...os.filter((x) => x.tempId !== o.tempId), { ...o, failed: false }])
    try {
      let input: SendInput
      if (o.kind === 'meme') {
        input = { temp_id: o.tempId, kind: 'meme', meme_id: o.memeId }
      } else if (o.kind === 'photo') {
        // 사진은 먼저 R2 에 올린다. 올린 키는 기억해 둬서 다시 보내기 때 또 올리지 않는다
        let key = o.photoKey
        if (!key) {
          if (!o.file) throw new Error('사진을 다시 골라 주세요')
          const img = await shrinkImage(o.file, 'photo')
          key = await uploadFile('photo', img.contentType, img.base64)
          o = { ...o, photoKey: key }
          const done = o
          setOutgoing((os) => os.map((x) => (x.tempId === done.tempId ? { ...x, photoKey: key } : x)))
        }
        input = { temp_id: o.tempId, kind: 'photo', photo_key: key }
      } else {
        input = { temp_id: o.tempId, body: o.body, shake_ids: o.shakeIds, shake_all: o.shakeAll }
      }
      if (o.replyToId) input.reply_to_id = o.replyToId
      const m = await sendMessage(azitId, roomId, input)
      merge([m])
      setOutgoing((os) => os.filter((x) => x.tempId !== o.tempId))
      if (o.kind === 'photo' && o.src) URL.revokeObjectURL(o.src)
    } catch (e) {
      setOutgoing((os) => os.map((x) => (x.tempId === o.tempId ? { ...x, failed: true } : x)))
      setError(errorText(e))
    }
  }

  const send = async () => {
    const body = draft.trim()
    if (!body) return
    if (editing) {
      try {
        merge([await editMessage(azitId, Number(editing.id), body)])
        setEditing(undefined)
        setDraft('')
      } catch (e) {
        setError(errorText(e))
      }
      return
    }
    // 골라서 넣은 @이름 중 본문에 아직 남아 있는 사람만 흔든다 (넣었다가 지운 건 빼고)
    const shakeIds = [...new Set(picked.filter((p) => body.includes(`@${p.nick}`)).map((p) => p.id))]
    setDraft('')
    setPicked([])
    setMention(null)
    await deliver({ tempId: newTempId(), body, at: Date.now(), shakeIds, shakeAll: body.includes('@전체'), replyToId: takeReply() })
  }

  // 답장: 다음에 보내는 메시지 하나에 붙이고 비운다
  const takeReply = () => {
    const id = replyTo
    setReplyTo(undefined)
    return id
  }
  const replyMsg = replyTo ? messages.find((m) => m.message_id === replyTo) : undefined

  // 채팅 사진 → 아지트 짤 (이름을 물어보는 창)
  const [saving, setSaving] = useState<number>()
  const [saveBusy, setSaveBusy] = useState(false)
  const [saveError, setSaveError] = useState<string>()
  const saveMeme = async (name: string) => {
    if (!saving) return
    setSaveBusy(true)
    try {
      await saveMemeFromMessage(azitId, saving, name)
      setSaving(undefined)
      setSaveError(undefined)
      onMemeSaved?.()
    } catch (e) {
      setSaveError(errorText(e))
    } finally {
      setSaveBusy(false)
    }
  }

  // @ 부르기 (= 흔들기). "@검색어" 를 치면 후보가 뜨고, 고르면 "@닉네임 " 으로 바꾸고 그 사람 id 를 기억한다
  const [mention, setMention] = useState<{ query: string; caret: number } | null>(null)
  const [mentionIndex, setMentionIndex] = useState(0)
  const [picked, setPicked] = useState<{ id: number; nick: string }[]>([])

  const nicks = useMemo(() => new Map(members.map((m) => [m.account_id, m.nickname])), [members])

  const candidates: (Member & { accountId: number })[] = useMemo(() => {
    if (!mention) return []
    const q = mention.query.toLowerCase()
    const all = { id: 'all', accountId: 0, nick: '전체', color: 'var(--lime)', via: '다같이 흔들기' }
    const people = members
      .filter((m) => m.account_id !== me.id)
      .filter((m) => !q || m.nickname.toLowerCase().includes(q) || m.name.toLowerCase().includes(q))
      .map((m) => ({ id: String(m.account_id), accountId: m.account_id, nick: m.nickname, color: m.color, src: m.avatar_url || undefined, via: sub(m) }))
    return ('전체'.includes(q) ? [all, ...people] : people).slice(0, 8)
  }, [mention, members, me.id])

  const onDraft = (v: string, caret: number) => {
    setDraft(v)
    if (editing) return
    const q = mentionQuery(v, caret)
    if (q === null) { setMention(null); return }
    if (mention?.query !== q) setMentionIndex(0)
    setMention({ query: q, caret })
  }

  const pickMention = (i: number) => {
    const c = candidates[i]
    if (!c || !mention) return
    const start = mention.caret - mention.query.length - 1 // "@" 자리
    const next = `${draft.slice(0, start)}@${c.nick} ${draft.slice(mention.caret)}`
    setDraft(next)
    if (c.accountId) setPicked((ps) => [...ps, { id: c.accountId, nick: c.nick }])
    setMention(null)
  }

  // + 메뉴: 짤 고르기 · 사진 고르기
  const [popup, setPopup] = useState<ComposerPopup>(null)
  const [viewing, setViewing] = useState<string>() // 크게 보는 사진
  const [profileId, setProfileId] = useState<number>() // 프로필 창을 띄운 멤버
  const [roomInfo, setRoomInfo] = useState(false) // 방 정보 창
  const [shaking, setShaking] = useState(false) // + → 흔들기 (사람 고르기 창)
  const profile = members.find((m) => m.account_id === profileId) // 나간 사람이면 없어서 안 뜬다

  const stickers: Sticker[] = useMemo(
    () => memes.map((m) => ({ id: String(m.meme_id), name: m.name, src: m.url, by: '', ago: ago(m.created_at) })),
    [memes],
  )

  // 맞장구 고르기 창 아래 칸: 아지트 짤 전부 (맞장구 목록에 없어도 달 수 있다)
  const memeOptions: ReactOption[] = useMemo(
    () => memes.map((m) => ({ id: `m${m.meme_id}`, img: m.url, name: m.name, memeId: m.meme_id })),
    [memes],
  )

  const sendMeme = (id: string) => {
    const m = memes.find((x) => String(x.meme_id) === id)
    if (!m) return
    setPopup(null)
    void deliver({ tempId: newTempId(), body: '', at: Date.now(), kind: 'meme', memeId: m.meme_id, src: m.url, name: m.name, replyToId: takeReply() })
  }

  const sendPhoto = (file: File) => {
    setPopup(null)
    if (file.size > 30 << 20) { setError('사진이 너무 커요'); return } // 줄이기 전 원본. 줄인 뒤 10MB 는 서버가 본다
    void deliver({ tempId: newTempId(), body: '', at: Date.now(), kind: 'photo', file, src: URL.createObjectURL(file), replyToId: takeReply() })
  }

  const retry = (msg: UserMessage) => {
    const o = outgoing.find((x) => x.tempId === msg.id)
    if (o) void deliver(o)
  }

  // 맞장구: 서버가 그 메시지의 맞장구 묶음 전체를 돌려줘서 그대로 바꿔 끼운다
  const setReactions = (id: number, reactions: MessageRes['reactions']) =>
    setMessages((cur) => cur.map((m) => (m.message_id === id ? { ...m, reactions } : m)))

  const react = async (msg: UserMessage, item: { text: string; meme_id: number }, on: boolean) => {
    const id = Number(msg.id)
    try {
      setReactions(id, await (on ? removeReaction : addReaction)(azitId, id, item))
      setError(undefined)
    } catch (e) {
      setError(errorText(e))
    }
  }

  const onReact = (msg: UserMessage, r: ReactOption) => {
    const item = { text: r.text ?? '', meme_id: r.memeId ?? 0 }
    const already = msg.reactions?.some((x) => (item.meme_id ? x.memeId === item.meme_id : x.text === item.text) && x.on)
    void react(msg, item, !!already)
  }

  const onToggle = (msg: UserMessage, r: Reaction) => void react(msg, { text: r.text ?? '', meme_id: r.memeId ?? 0 }, !!r.on)

  // 지우기: 먼저 확인 창을 띄우고(confirm 은 데스크톱 웹뷰에서 안 뜰 수 있어서 직접 만든 창), 지우기를 누르면 서버에 지운다
  const [deleting, setDeleting] = useState<UserMessage>()
  const [deleteBusy, setDeleteBusy] = useState(false)
  const onDelete = (msg: UserMessage) => setDeleting(msg)
  const confirmDelete = async () => {
    if (!deleting) return
    const id = Number(deleting.id)
    setDeleteBusy(true)
    try {
      await deleteMessage(azitId, id)
      setMessages((cur) => cur.map((m) => (m.message_id === id ? { ...m, deleted: true, body: '', reactions: [] } : m)))
      setError(undefined)
    } catch (e) {
      setError(errorText(e))
    } finally {
      setDeleteBusy(false)
      setDeleting(undefined)
    }
  }

  // 답장 인용 누르기 → 원래 메시지로. 아직 안 불러왔으면 위로 한 묶음씩 더 불러오며 찾는다 (최대 20번 = 1000개)
  const [jump, setJump] = useState<{ id: string; n: number }>()
  const jumpTo = async (id: string) => {
    const target = Number(id)
    const go = () => setJump((j) => ({ id, n: (j?.n ?? 0) + 1 }))
    if (messages.some((m) => m.message_id === target)) { go(); return }
    let oldest = messages[0]?.message_id
    if (!oldest) return
    setLoadingMore(true)
    try {
      for (let i = 0; i < 20; i++) {
        const res = await listMessages(azitId, roomId, oldest)
        merge(res.messages)
        setHasMore(res.has_more)
        if (res.messages.some((m) => m.message_id === target)) { go(); return }
        if (!res.has_more || res.messages.length === 0) break
        oldest = Math.min(...res.messages.map((m) => m.message_id))
      }
      setError('원래 메시지를 찾지 못했어요')
    } catch (e) {
      setError(errorText(e))
    } finally {
      setLoadingMore(false)
    }
  }

  const onEdit = (msg: UserMessage) => {
    setReplyTo(undefined)
    setEditing(msg)
    setDraft(msg.kind === 'text' ? msg.parts.map((p) => ('text' in p ? p.text : `@${p.mention}`)).join('') : '')
  }

  const rows = useMemo(() => toRows(messages, outgoing, me, lastReadId, nicks), [messages, outgoing, me, lastReadId, nicks])

  return (
    <>
    <ChatView
      room={room}
      memberCount={memberCount}
      rows={rows}
      reactOptions={reactOptions}
      hasMore={hasMore}
      loadingMore={loadingMore}
      error={error}
      onMore={more}
      onReact={onReact}
      onToggle={onToggle}
      onEdit={onEdit}
      onDelete={onDelete}
      onRetry={retry}
      onImage={setViewing}
      onProfile={(id) => setProfileId(Number(id))}
      onRoomInfo={() => setRoomInfo(true)}
      memeOptions={memeOptions}
      canDelete={canDeleteOthers}
      onReply={(msg) => { setEditing(undefined); setReplyTo(Number(msg.id)) }}
      onSaveMeme={(msg) => { setSaveError(undefined); setSaving(Number(msg.id)) }}
      onJump={(id) => void jumpTo(id)}
      jumpTo={jump}
      composer={{
        value: draft,
        editing: !!editing,
        popup: mention && candidates.length > 0 ? 'mention' : popup,
        stickers,
        mentionCandidates: candidates,
        mentionIndex,
        onChange: onDraft,
        onMentionMove: (d) => setMentionIndex((i) => (i + d + candidates.length) % candidates.length),
        onMentionPick: pickMention,
        onSend: () => void send(),
        onCancelEdit: () => { setEditing(undefined); setDraft('') },
        replying: replyMsg && { nick: replyMsg.sender?.nickname ?? '알 수 없음', text: preview(replyMsg) },
        onCancelReply: () => setReplyTo(undefined),
        onPlus: () => { setMention(null); setPopup(popup ? null : 'plus') },
        onPopup: (p) => { if (mention) setMention(null); else setPopup(p) },
        onSticker: sendMeme,
        onPhoto: sendPhoto,
        onShake: () => { setPopup(null); setShaking(true) },
      }}
    />
    {deleting && <DeleteMessageModal busy={deleteBusy} onCancel={() => setDeleting(undefined)} onConfirm={() => void confirmDelete()} />}
    {saving && <SaveMemeModal busy={saveBusy} error={saveError} onCancel={() => setSaving(undefined)} onSave={(name) => void saveMeme(name)} />}
    {viewing && <ImageModal src={viewing} onClose={() => setViewing(undefined)} />}
    {shaking && (
      <ShakeModal
        members={members.filter((m) => m.account_id !== me.id).map((m) => ({ id: String(m.account_id), nick: m.nickname, color: m.color, src: m.avatar_url || undefined, via: sub(m) }))}
        onPick={(id) => {
          const m = members.find((x) => String(x.account_id) === id)
          setShaking(false)
          if (m) void deliver({ tempId: newTempId(), body: `@${m.nickname}`, at: Date.now(), shakeIds: [m.account_id] })
        }}
        onClose={() => setShaking(false)}
      />
    )}
    {roomInfo && (
      <RoomInfoModal
        room={room}
        members={members.map((m) => ({ id: String(m.account_id), nick: m.nickname, color: m.color, src: m.avatar_url || undefined, via: [m.online && '접속 중', ROLE_TAG[m.role]].filter(Boolean).join(' · ') }))}
        meId={String(me.id)}
        onProfile={(id) => { setRoomInfo(false); setProfileId(Number(id)) }}
        onClose={() => setRoomInfo(false)}
      />
    )}
    {profile && <MemberProfileModal member={profile} self={profile.account_id === me.id} onClose={() => setProfileId(undefined)} />}
    </>
  )
}
