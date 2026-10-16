import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  addReaction, deleteMessage, editMessage, errorText, listMessages, markRead, removeReaction, sendMessage, uploadFile,
  type MemeRes, type MessageRes, type SendInput,
} from '../../api/server'
import type { ReactOption, Reaction, Room, Sticker, UserMessage } from '../../api/types'
import { ago, newTempId, toRows, type Me, type Outgoing } from '../../lib/chat'
import { shrinkImage } from '../../lib/image'
import { DeleteMessageModal, ImageModal } from './ChatModals'
import { ChatView } from './ChatView'
import type { ComposerPopup } from './Composer'

const POLL_MS = 3000 // 웹소켓 전까지는 3초마다 새 메시지를 물어본다

type Props = {
  azitId: number
  room: Room
  memberCount: number
  me: Me
  reactOptions: ReactOption[]
  memes: MemeRes[] // + → 짤 에서 고를 아지트 짤
  canDeleteOthers: boolean // 관리자
  onRead?: () => void // 읽음 처리했으니 방 목록 안 읽은 수를 다시 받으라고
  onShaken?: () => void // 나를 흔든 새 메시지가 왔다
}

// 방 하나의 채팅. 서버에서 메시지를 받아 ChatView 에 넘기고, 보내기·고치기·지우기·맞장구를 처리한다.
export function ChatScreen({ azitId, room, memberCount, me, reactOptions, memes, canDeleteOthers, onRead, onShaken }: Props) {
  const roomId = Number(room.id)
  const [messages, setMessages] = useState<MessageRes[]>([])
  const [hasMore, setHasMore] = useState(false)
  const [lastReadId, setLastReadId] = useState(0)
  const [loadingMore, setLoadingMore] = useState(false)
  const [outgoing, setOutgoing] = useState<Outgoing[]>([])
  const [draft, setDraft] = useState('')
  const [editing, setEditing] = useState<UserMessage>()
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
    setMessages([]); setOutgoing([]); setHasMore(false); setEditing(undefined); setDraft(''); setError(undefined)
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
        input = { temp_id: o.tempId, body: o.body, shake_all: o.shakeAll }
      }
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
    setDraft('')
    await deliver({ tempId: newTempId(), body, at: Date.now(), shakeAll: body.includes('@전체') })
  }

  // + 메뉴: 짤 고르기 · 사진 고르기
  const [popup, setPopup] = useState<ComposerPopup>(null)
  const [viewing, setViewing] = useState<string>() // 크게 보는 사진

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
    void deliver({ tempId: newTempId(), body: '', at: Date.now(), kind: 'meme', memeId: m.meme_id, src: m.url, name: m.name })
  }

  const sendPhoto = (file: File) => {
    setPopup(null)
    if (file.size > 30 << 20) { setError('사진이 너무 커요'); return } // 줄이기 전 원본. 줄인 뒤 10MB 는 서버가 본다
    void deliver({ tempId: newTempId(), body: '', at: Date.now(), kind: 'photo', file, src: URL.createObjectURL(file) })
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

  const onEdit = (msg: UserMessage) => {
    setEditing(msg)
    setDraft(msg.kind === 'text' ? msg.parts.map((p) => ('text' in p ? p.text : `@${p.mention}`)).join('') : '')
  }

  const rows = useMemo(() => toRows(messages, outgoing, me, lastReadId), [messages, outgoing, me, lastReadId])

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
      memeOptions={memeOptions}
      canDelete={canDeleteOthers}
      composer={{
        value: draft,
        editing: !!editing,
        popup,
        stickers,
        onChange: setDraft,
        onSend: () => void send(),
        onCancelEdit: () => { setEditing(undefined); setDraft('') },
        onPlus: () => setPopup(popup ? null : 'plus'),
        onPopup: setPopup,
        onSticker: sendMeme,
        onPhoto: sendPhoto,
      }}
    />
    {deleting && <DeleteMessageModal busy={deleteBusy} onCancel={() => setDeleting(undefined)} onConfirm={() => void confirmDelete()} />}
    {viewing && <ImageModal src={viewing} onClose={() => setViewing(undefined)} />}
    </>
  )
}
