import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  createInvite, createMeme, createRoom, deleteMeme, errorText, isAdmin, kickMember, leaveAzit, listInvites, listMembers, listMemes,
  listReacts, listRooms, renameRoom, reorderRooms, replaceReacts, revokeInvite, updateAzit, updateMemberRole, updateMyNickname, uploadFile,
  type AzitPatch, type AzitRes, type InviteRes, type MemberRes, type MemeRes, type ReactItem, type ReactRes, type RoomRes, type Role,
} from '../../api/server'
import type { Azit, InviteCode, Member, ReactOption, Room, Sticker } from '../../api/types'
import { Shell } from '../../app/Shell'
import { ago, left, type Me } from '../../lib/chat'
import { prepareMeme, shrinkImage } from '../../lib/image'
import { ChatScreen } from '../chat/ChatScreen'
import { NewRoomModal } from '../rooms/NewRoomModal'
import { RoomList } from '../rooms/RoomList'
import { MemberProfileModal } from './MemberProfileModal'
import { CodesTab, type CodeExpiry } from './settings/CodesTab'
import { GeneralTab, type GeneralValues } from './settings/GeneralTab'
import { LeaveTab } from './settings/LeaveTab'
import { MembersTab } from './settings/MembersTab'
import { ReactsTab } from './settings/ReactsTab'
import { RoomsTab } from './settings/RoomsTab'
import { SettingsView, type SettingsTab } from './settings/SettingsView'
import { StickersTab } from './settings/StickersTab'

const ROOMS_POLL_MS = 5000
// 서버 API 가 있는 설정 탭만 (기본 정보·변경 기록은 아직 없음)
const TABS: SettingsTab[] = ['general', 'codes', 'members', 'rooms', 'reacts', 'stickers', 'leave']
const ROLE_LABEL = { super_admin: '만든 사람', admin: '관리자', member: '멤버', '': '' } as const

type Props = {
  azit: AzitRes
  me: Me
  onBack: () => void
  onChanged?: () => void // 내 닉네임·인원이 바뀌었다 → 아지트 목록 다시 받기
  onLeft?: () => void // 아지트를 나갔다
  initialRoomId?: number // 처음 열 방 (알림에서 들어올 때). 없거나 이 아지트 방이 아니면 맨 위 방
}

export const shortName = (name: string) => Array.from(name.trim()).slice(0, 2).join('')

// 아지트 안: 방 목록 + (채팅 | 설정). 방·맞장구 목록·짤은 여기서 받아 아래로 나눠 준다.
export function AzitScreen({ azit, me, onBack, onChanged, onLeft, initialRoomId }: Props) {
  const azitId = azit.azit_id
  const admin = isAdmin(azit.my_role)
  const [rooms, setRooms] = useState<RoomRes[]>([])
  const [roomId, setRoomId] = useState<number>()
  const [view, setView] = useState<'chat' | SettingsTab>('chat')
  const [reacts, setReacts] = useState<ReactRes[]>([])
  const [memes, setMemes] = useState<MemeRes[]>([])
  const [members, setMembers] = useState<MemberRes[]>([]) // @ 후보, 메시지 속 @이름 강조
  const [query, setQuery] = useState('')
  const [newRoom, setNewRoom] = useState(false)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string>()

  const loadRooms = useCallback(() => {
    listRooms(azitId)
      .then((rs) => {
        setRooms(rs)
        const first = rs.some((r) => r.room_id === initialRoomId) ? initialRoomId : rs[0]?.room_id
        setRoomId((cur) => cur ?? first)
      })
      .catch((e) => setError(errorText(e)))
  }, [azitId, initialRoomId])

  useEffect(() => {
    setRooms([]); setRoomId(undefined); setView('chat'); setError(undefined)
    loadRooms()
    listReacts(azitId).then(setReacts).catch(() => {})
    listMemes(azitId).then(setMemes).catch(() => {})
    listMembers(azitId).then(setMembers).catch(() => {})
    const t = window.setInterval(loadRooms, ROOMS_POLL_MS)
    return () => window.clearInterval(t)
  }, [azitId, loadRooms])


  // 화면 모양으로 바꾸기
  const az: Azit = { id: String(azitId), name: azit.name, short: shortName(azit.name), color: azit.color, memberCount: azit.member_count, unread: 0 }
  const roomViews: Room[] = rooms.map((r) => ({ id: String(r.room_id), name: r.name, preview: '', unread: r.unread }))
  const shown = query.trim() ? roomViews.filter((r) => r.name.includes(query.trim())) : roomViews
  const room = roomViews.find((r) => r.id === String(roomId))

  const reactOptions: ReactOption[] = useMemo(
    () => reacts.map((r) => r.meme_id
      ? { id: `r${r.react_id}`, img: r.meme_url, name: memes.find((m) => m.meme_id === r.meme_id)?.name ?? '짤', memeId: r.meme_id }
      : { id: `r${r.react_id}`, text: r.text }),
    [reacts, memes],
  )

  const openRoom = (id: string) => { setRoomId(Number(id)); setView('chat') }

  const addRoom = async (name: string) => {
    setBusy(true)
    try {
      const r = await createRoom(azitId, name)
      setNewRoom(false)
      setError(undefined)
      loadRooms()
      openRoom(String(r.room_id))
    } catch (e) {
      setError(errorText(e))
    } finally {
      setBusy(false)
    }
  }

  let body
  if (view === 'chat') {
    body = room
      ? (
        <ChatScreen
          key={room.id}
          azitId={azitId}
          room={room}
          memberCount={azit.member_count}
          me={me}
          reactOptions={reactOptions}
          memes={memes}
          members={members}
          canDeleteOthers={admin}
          onRead={loadRooms}
        />
      )
      : <div className="main"><div className="page mu" style={{ margin: 'auto' }}>{error ?? '방을 불러오는 중…'}</div></div>
  } else {
    body = (
      <SettingsView tab={view} tabs={TABS} hint={`내 역할: ${ROLE_LABEL[azit.my_role]}`} onTab={(t) => { setView(t); setError(undefined) }}>
        {view === 'codes' && <Codes azitId={azitId} meId={me.id} readOnly={!admin} />}
        {view === 'general' && <General key={azitId} azit={azit} readOnly={!admin} onChanged={onChanged} />}
        {view === 'members' && <Members azitId={azitId} meId={me.id} myRole={azit.my_role} onChanged={onChanged} />}
        {view === 'leave' && <Leave azitId={azitId} owner={azit.my_role === 'super_admin'} onLeft={onLeft} />}
        {view === 'rooms' && (
          <RoomsSettings
            azitId={azitId}
            rooms={rooms}
            onChanged={setRooms}
            onOpen={openRoom}
            onNewRoom={() => setNewRoom(true)}
          />
        )}
        {view === 'reacts' && <Reacts azitId={azitId} reacts={reacts} memes={memes} readOnly={!admin} onChanged={setReacts} />}
        {view === 'stickers' && (
          <Memes
            azitId={azitId}
            memes={memes}
            canDelete={(m) => admin || m.created_by === me.id}
            onChanged={(ms) => {
              setMemes(ms)
              listReacts(azitId).then(setReacts).catch(() => {}) // 짤을 빼면 맞장구 목록에서도 빠진다
            }}
          />
        )}
      </SettingsView>
    )
  }

  return (
    <>
      {/* 흔들기는 MeScreen 이 알림을 보고 창 전체를 흔든다 (다른 방·아지트에 있어도) */}
      <Shell>
        <RoomList
          azit={az}
          rooms={shown}
          currentRoomId={view === 'chat' ? room?.id : undefined}
          query={query}
          active={view === 'chat' ? undefined : 'settings'}
          roleLabel={ROLE_LABEL[azit.my_role]}
          noHub
          onBack={onBack}
          onRoom={openRoom}
          onQuery={setQuery}
          onSettings={() => setView(view === 'chat' ? 'general' : 'chat')}
          onNewRoom={() => { setError(undefined); setNewRoom(true) }}
          onInvite={() => setView('codes')}
        />
        {body}
      </Shell>
      {newRoom && (
        <NewRoomModal memberCount={azit.member_count} busy={busy} error={error} onSubmit={addRoom} onClose={() => { setNewRoom(false); setError(undefined) }} />
      )}
    </>
  )
}

/* ---------- 설정 탭들 (서버와 주고받기) ---------- */

// 기본 정보: 고친 값은 여기 들고 있다가 저장 때 바뀐 칸만 보낸다. 배너는 고르는 즉시 올려 두고 키만 기억한다
function General({ azit, readOnly, onChanged }: { azit: AzitRes; readOnly: boolean; onChanged?: () => void }) {
  const saved: GeneralValues = { name: azit.name, color: azit.color, desc: azit.description ?? '', banner: azit.banner_url || undefined }
  const [value, setValue] = useState<GeneralValues>(saved)
  const [bannerKey, setBannerKey] = useState<string>() // 새로 올린 배너 키
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string>()
  const [done, setDone] = useState(false)

  const patch: AzitPatch = {}
  if (value.name.trim() !== azit.name) patch.name = value.name
  if (value.color !== azit.color) patch.color = value.color
  if (value.desc.trim() !== (azit.description ?? '')) patch.description = value.desc
  if (bannerKey) patch.banner_key = bannerKey
  else if (!value.banner && azit.banner_url) patch.banner_key = '' // 배너 빼기
  const dirty = Object.keys(patch).length > 0

  const pickBanner = async (file: File) => {
    setBusy(true)
    try {
      const img = await shrinkImage(file, 'azit_banner')
      const key = await uploadFile('azit_banner', img.contentType, img.base64)
      setBannerKey(key)
      setValue((v) => ({ ...v, banner: `data:${img.contentType};base64,${img.base64}` }))
      setError(undefined)
    } catch (e) {
      setError(errorText(e))
    } finally {
      setBusy(false)
    }
  }

  const save = async () => {
    setBusy(true)
    try {
      const a = await updateAzit(azit.azit_id, patch)
      setValue({ name: a.name, color: a.color, desc: a.description, banner: a.banner_url || undefined })
      setBannerKey(undefined)
      setError(undefined)
      setDone(true)
      onChanged?.() // 아지트 목록을 다시 받아 왼쪽 이름·색도 바뀐다
    } catch (e) {
      setError(errorText(e))
    } finally {
      setBusy(false)
    }
  }

  return (
    <GeneralTab
      value={value}
      memberCount={azit.member_count}
      readOnly={readOnly}
      dirty={dirty}
      busy={busy}
      error={error}
      saved={done}
      onChange={(v) => { setValue(v); if (!v.banner) setBannerKey(undefined); setDone(false) }}
      onBanner={(f) => void pickBanner(f)}
      onSave={() => void save()}
      onReset={() => { setValue(saved); setBannerKey(undefined); setError(undefined) }}
    />
  )
}

function Members({ azitId, meId, myRole, onChanged }: { azitId: number; meId: number; myRole: Role | ''; onChanged?: () => void }) {
  const [members, setMembers] = useState<MemberRes[]>([])
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string>()

  const load = useCallback(() => listMembers(azitId).then(setMembers).catch((e) => setError(errorText(e))), [azitId])
  useEffect(() => { void load() }, [load])

  // 바꾸고 나면 목록을 다시 받는다 (역할이 바뀌면 순서도 바뀌니까)
  const run = async (f: () => Promise<unknown>) => {
    setBusy(true)
    try {
      await f()
      setError(undefined)
      await load()
      onChanged?.()
      return true
    } catch (e) {
      setError(errorText(e))
      return false
    } finally {
      setBusy(false)
    }
  }

  const views: Member[] = members.map((m) => ({
    id: String(m.account_id), nick: m.nickname, name: m.name, color: m.color, src: m.avatar_url || undefined, role: m.role, joinedAgo: ago(m.joined_at),
  }))
  const [profileId, setProfileId] = useState<number>()
  const profile = members.find((m) => m.account_id === profileId)

  return (
    <>
      <MembersTab
        members={views}
        meId={String(meId)}
        myRole={myRole || 'member'}
        busy={busy}
        error={error}
        onNick={(nick) => run(() => updateMyNickname(azitId, nick))}
        onRole={(id, role) => void run(() => updateMemberRole(azitId, Number(id), role))}
        onKick={(id) => void run(() => kickMember(azitId, Number(id)))}
        onProfile={(id) => setProfileId(Number(id))}
      />
      {profile && <MemberProfileModal member={profile} self={profile.account_id === meId} onClose={() => setProfileId(undefined)} />}
    </>
  )
}

function Leave({ azitId, owner, onLeft }: { azitId: number; owner: boolean; onLeft?: () => void }) {
  const [presses, setPresses] = useState<0 | 1 | 2>(0)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string>()

  const press = async () => {
    if (presses < 2) { setPresses((p) => (p + 1) as 1 | 2); return }
    setBusy(true)
    try {
      await leaveAzit(azitId)
      onLeft?.()
    } catch (e) {
      setError(errorText(e))
      setPresses(0)
      setBusy(false)
    }
  }

  return <LeaveTab owner={owner} presses={presses} busy={busy} error={error} onPress={() => void press()} />
}

function Codes({ azitId, meId, readOnly }: { azitId: number; meId: number; readOnly: boolean }) {
  const [codes, setCodes] = useState<InviteRes[]>([])
  const [expires, setExpires] = useState<CodeExpiry>('7d')
  const [maxUses, setMaxUses] = useState<0 | 1 | 5 | 10>(0)
  const [made, setMade] = useState<InviteRes>()
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string>()

  const load = useCallback(() => {
    if (readOnly) return
    listInvites(azitId).then(setCodes).catch((e) => setError(errorText(e)))
  }, [azitId, readOnly])
  useEffect(load, [load])

  const create = async () => {
    setBusy(true)
    try {
      const inv = await createInvite(azitId, expires, maxUses)
      setMade(inv)
      setError(undefined)
      load()
    } catch (e) {
      setError(errorText(e))
    } finally {
      setBusy(false)
    }
  }

  const revoke = async (code: string) => {
    if (!window.confirm(`${code} 코드를 폐기할까요? 되돌릴 수 없어요`)) return
    try {
      await revokeInvite(code)
      if (made?.code === code) setMade(undefined)
      load()
    } catch (e) {
      setError(errorText(e))
    }
  }

  const copy = (code: string) => void navigator.clipboard?.writeText(code)

  const views: InviteCode[] = codes.map((c) => ({
    code: c.code,
    madeBy: c.created_by === meId ? '내가' : '멤버가',
    state: 'live',
    left: left(c.expires_at),
    uses: c.uses,
    max: c.max_uses,
    joined: [],
  }))

  return (
    <CodesTab
      codes={views}
      expiresIn={expires}
      maxUses={maxUses}
      made={made && { code: made.code, left: left(made.expires_at), uses: made.max_uses ? `${made.max_uses}번 쓸 수 있어요` : '횟수 제한 없음' }}
      readOnly={readOnly}
      busy={busy}
      error={error}
      onExpires={setExpires}
      onMaxUses={setMaxUses}
      onCreate={create}
      onCopy={copy}
      onRevoke={revoke}
    />
  )
}

function RoomsSettings({ azitId, rooms, onChanged, onOpen, onNewRoom }: {
  azitId: number; rooms: RoomRes[]; onChanged: (rs: RoomRes[]) => void; onOpen: (id: string) => void; onNewRoom: () => void
}) {
  const [selected, setSelected] = useState<string>(String(rooms[0]?.room_id ?? ''))
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string>()

  const run = async (f: () => Promise<void>) => {
    setBusy(true)
    try {
      await f()
      setError(undefined)
    } catch (e) {
      setError(errorText(e))
    } finally {
      setBusy(false)
    }
  }

  const rename = (id: string, name: string) => run(async () => {
    const r = await renameRoom(azitId, Number(id), name)
    onChanged(rooms.map((x) => (x.room_id === r.room_id ? { ...x, name: r.name } : x)))
  })

  const move = (id: string, dir: -1 | 1) => run(async () => {
    const ids = rooms.map((r) => r.room_id)
    const i = ids.indexOf(Number(id))
    const j = i + dir
    if (i < 0 || j < 0 || j >= ids.length) return
    ;[ids[i], ids[j]] = [ids[j], ids[i]]
    const res = await reorderRooms(azitId, ids)
    const unread = new Map(rooms.map((r) => [r.room_id, r.unread]))
    onChanged(res.map((r) => ({ ...r, unread: unread.get(r.room_id) ?? r.unread })))
  })

  if (rooms.length === 0) return <span className="mu">방을 불러오는 중…</span>
  return (
    <RoomsTab
      rooms={rooms.map((r) => ({ id: String(r.room_id), name: r.name, preview: '', unread: r.unread }))}
      selectedId={selected}
      busy={busy}
      error={error}
      onSelect={setSelected}
      onRename={rename}
      onMove={move}
      onOpen={onOpen}
      onNewRoom={onNewRoom}
    />
  )
}

function Reacts({ azitId, reacts, memes, readOnly, onChanged }: {
  azitId: number; reacts: ReactRes[]; memes: MemeRes[]; readOnly: boolean; onChanged: (rs: ReactRes[]) => void
}) {
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string>()

  const items: ReactItem[] = reacts.map((r) => ({ text: r.text, meme_id: r.meme_id }))

  // 목록은 통째로 바꾼다 (PUT)
  const save = async (next: ReactItem[]) => {
    setBusy(true)
    try {
      onChanged(await replaceReacts(azitId, next))
      setError(undefined)
      return true
    } catch (e) {
      setError(errorText(e))
      return false
    } finally {
      setBusy(false)
    }
  }

  const options: ReactOption[] = reacts.map((r) => r.meme_id
    ? { id: `r${r.react_id}`, img: r.meme_url, name: memes.find((m) => m.meme_id === r.meme_id)?.name ?? '짤', memeId: r.meme_id }
    : { id: `r${r.react_id}`, text: r.text })

  return (
    <ReactsTab
      reacts={options}
      draft=""
      error={error}
      readOnly={readOnly}
      busy={busy}
      onAdd={(text) => save([...items, { text, meme_id: 0 }])}
      onMove={(i, dir) => {
        const next = [...items]
        const j = i + dir
        if (j < 0 || j >= next.length) return
        ;[next[i], next[j]] = [next[j], next[i]]
        void save(next)
      }}
      onRemove={(i) => void save(items.filter((_, k) => k !== i))}
    />
  )
}

function Memes({ azitId, memes, canDelete, onChanged }: {
  azitId: number; memes: MemeRes[]; canDelete: (m: MemeRes) => boolean; onChanged: (ms: MemeRes[]) => void
}) {
  const [uploading, setUploading] = useState(false)
  const [error, setError] = useState<string>()

  const upload = async (file: File, name: string) => {
    setUploading(true)
    try {
      const img = await prepareMeme(file) // 움직이는 GIF·WebP 는 그대로
      const key = await uploadFile('meme', img.contentType, img.base64)
      const m = await createMeme(azitId, name, key)
      onChanged([m, ...memes])
      setError(undefined)
      return true
    } catch (e) {
      setError(errorText(e))
      return false
    } finally {
      setUploading(false)
    }
  }

  const remove = async (id: string) => {
    const m = memes.find((x) => String(x.meme_id) === id)
    if (!m || !window.confirm(`"${m.name}" 짤을 뺄까요? 맞장구 목록에서도 빠져요`)) return
    try {
      await deleteMeme(azitId, m.meme_id)
      onChanged(memes.filter((x) => x.meme_id !== m.meme_id))
    } catch (e) {
      setError(errorText(e))
    }
  }

  const stickers: Sticker[] = memes.map((m) => ({ id: String(m.meme_id), name: m.name, src: m.url, by: '', ago: ago(m.created_at) }))

  return (
    <StickersTab
      stickers={stickers}
      uploading={uploading}
      error={error}
      canDelete={(id) => { const m = memes.find((x) => String(x.meme_id) === id); return !!m && canDelete(m) }}
      onUpload={upload}
      onDelete={remove}
    />
  )
}
