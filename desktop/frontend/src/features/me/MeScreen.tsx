import { useCallback, useEffect, useRef, useState } from 'react'
import {
  getMe, profileErrorText, toProfile, updateMe, uploadImage,
  type ImageKind, type Me, type ProfileFields,
} from '../../api/profile'
import {
  createAzit, errorText, joinInvite, listAzits, listNotifications, previewInvite, readAllNotifications,
  type AzitRes, type InvitePreview, type NotificationRes,
} from '../../api/server'
import type { Notification } from '../../api/types'
import { ago } from '../../lib/chat'
import { ActivityView } from '../activity/ActivityView'
import { Rail } from '../../app/Rail'
import { Shell } from '../../app/Shell'
import { shrinkImage } from '../../lib/image'
import { AzitScreen, shortName } from '../azit/AzitScreen'
import { CreateAzitScreen, type NewAzit } from '../azit/CreateAzitScreen'
import { HomeView, type HomeAzit } from '../home/HomeView'
import { ProfileSetupScreen } from '../auth/ProfileSetupScreen'
import { MeView } from './MeView'

type Props = { onLogout: () => void }

const NOTIF_POLL_MS = 3000 // 웹소켓 전까지는 3초마다 알림을 물어본다

// 창 전체를 흔든다 (base.css 의 .shake). 어느 화면에 있든 같은 자리.
function shakeWindow() {
  const el = document.body
  el.classList.remove('shake')
  void el.offsetWidth // 이어서 흔들려도 애니메이션이 처음부터 다시 돌게
  el.classList.add('shake')
  window.setTimeout(() => el.classList.remove('shake'), 700)
}

// 서버 알림 → 알림 화면 한 줄. 문구는 여기서 만든다.
function toNotification(n: NotificationRes): Notification {
  const who = n.actor_nickname || '누군가'
  const text = n.kind === 'shake'
    ? `${who}가 ${n.room_name || '방'}에서 불렀어요${n.preview ? ` · ${n.preview}` : ''}`
    : n.kind === 'join'
      ? `${who}가 들어왔어요`
      : `${who}가 나를 내보냈어요`
  return { id: String(n.notification_id), kind: n.kind, text, azit: n.azit_name, ago: ago(n.created_at), read: n.read, gone: n.kind === 'kick' }
}

// 로그인한 뒤의 진짜 화면.
// 서버에서 내 정보를 받아, 프로필이 없으면 프로필 만들기, 있으면 레일 + 홈(처음) · 내 정보 · 아지트 안을 보여 준다.
// 저장·사진 올리기·아지트 만들기·코드로 들어가기는 여기서 처리하고, 화면 파일들은 그리기만 한다.
export function MeScreen({ onLogout }: Props) {
  const [me, setMe] = useState<Me>()
  const [loadError, setLoadError] = useState<string>()
  const [busy, setBusy] = useState<ImageKind[]>([]) // 배너·프사는 따로 올릴 수 있다
  const [uploadError, setUploadError] = useState<string>()
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState<string>()
  const [view, setView] = useState<'home' | 'me' | 'activity' | 'create' | number>('home') // number = 그 아지트 안
  const [azits, setAzits] = useState<AzitRes[]>([])
  const [code, setCode] = useState('')
  const [codeError, setCodeError] = useState<string>()
  const [invite, setInvite] = useState<InvitePreview>()
  const [homeBusy, setHomeBusy] = useState(false)
  const [createError, setCreateError] = useState<string>()
  const [notifs, setNotifs] = useState<NotificationRes[]>([])
  const [unread, setUnread] = useState(0)
  const [notifError, setNotifError] = useState<string>()
  const [openRoom, setOpenRoom] = useState<number>() // 알림에서 눌러 들어갈 방
  const seenShake = useRef<number | null>(null) // 이미 흔든 마지막 흔들기 알림 id. null = 아직 처음 목록을 안 받음

  const load = useCallback(() => {
    getMe()
      .then((m) => { setMe(m); setLoadError(undefined) })
      .catch((e) => setLoadError(profileErrorText(e)))
  }, [])

  useEffect(load, [load])

  const loadAzits = useCallback(() => listAzits().then(setAzits).catch((e) => setCodeError(errorText(e))), [])
  // 아지트 목록 (인원 수가 바뀌니 30초마다 다시)
  const ready = !!me && !me.needs_profile
  useEffect(() => {
    if (!ready) return
    void loadAzits()
    const t = window.setInterval(() => void loadAzits(), 30000)
    return () => window.clearInterval(t)
  }, [ready, loadAzits])

  // 알림: 3초마다 받아서 종 숫자를 갱신하고, 새 흔들기 알림이 오면 어느 화면에 있든 창을 흔든다.
  // 처음 받은 목록으로는 흔들지 않는다 (켜자마자 예전 알림으로 흔들리지 않게).
  const loadNotifs = useCallback(async () => {
    const res = await listNotifications()
    setNotifs(res.notifications)
    setUnread(res.unread)
    const newest = res.notifications.find((n) => n.kind === 'shake')?.notification_id ?? 0
    if (seenShake.current !== null && res.notifications.some((n) => n.kind === 'shake' && !n.read && n.notification_id > (seenShake.current ?? 0))) {
      shakeWindow()
    }
    seenShake.current = Math.max(seenShake.current ?? 0, newest)
  }, [])

  useEffect(() => {
    if (!ready) return
    const tick = () => void loadNotifs().then(() => setNotifError(undefined)).catch((e) => setNotifError(errorText(e)))
    tick()
    const t = window.setInterval(tick, NOTIF_POLL_MS)
    return () => window.clearInterval(t)
  }, [ready, loadNotifs])

  // 알림 화면을 열면 다 읽은 걸로
  const openActivity = () => {
    setView('activity')
    if (unread > 0) void readAllNotifications().then(loadNotifs).catch((e) => setNotifError(errorText(e)))
  }

  const openNotif = (id: string) => {
    const n = notifs.find((x) => String(x.notification_id) === id)
    if (!n || n.kind === 'kick') return // 내보내진 아지트는 못 들어간다
    enter(n.azit_id, n.room_id || undefined)
  }

  // 홈: 코드 확인 → 미리보기 → 들어가기
  const checkCode = async () => {
    setHomeBusy(true)
    try {
      setInvite(await previewInvite(code))
      setCodeError(undefined)
    } catch (e) {
      setCodeError(errorText(e))
    } finally {
      setHomeBusy(false)
    }
  }

  const join = async () => {
    if (!invite) return
    if (invite.already_member) { enter(invite.azit.azit_id); return }
    setHomeBusy(true)
    try {
      const a = await joinInvite(invite.code)
      await loadAzits()
      enter(a.azit_id)
    } catch (e) {
      setCodeError(errorText(e))
    } finally {
      setHomeBusy(false)
    }
  }

  // roomId: 알림에서 들어오면 그 방부터 연다
  const enter = (id: number, roomId?: number) => {
    setInvite(undefined); setCode(''); setCodeError(undefined)
    setOpenRoom(roomId)
    setView(id)
  }

  // 새 아지트: 만들고 그 아지트로 (방은 서버가 기본 방 하나만 만든다)
  const create = async (a: NewAzit) => {
    setHomeBusy(true)
    try {
      const made = await createAzit(a.name, a.color)
      await loadAzits()
      setCreateError(undefined)
      setView(made.azit_id)
    } catch (e) {
      setCreateError(errorText(e))
    } finally {
      setHomeBusy(false)
    }
  }

  const pick = async (kind: ImageKind, file: File) => {
    setBusy((b) => [...b, kind])
    setUploadError(undefined)
    try {
      const img = await shrinkImage(file, kind)
      await uploadImage(kind, img.contentType, img.base64)
      load()
    } catch (e) {
      setUploadError(profileErrorText(e))
    } finally {
      setBusy((b) => b.filter((k) => k !== kind))
    }
  }

  const save = async (f: ProfileFields) => {
    setSaving(true)
    setSaveError(undefined)
    try {
      setMe(await updateMe(f))
    } catch (e) {
      setSaveError(profileErrorText(e))
    } finally {
      setSaving(false)
    }
  }

  if (!me) {
    return (
      <div className="main">
        <div className="scroll page col" style={{ alignItems: 'center', justifyContent: 'center', gap: 12 }}>
          {loadError
            ? <>
                <span role="alert" className="ferr">{loadError}</span>
                <div className="row" style={{ gap: 8 }}>
                  <button type="button" className="btn sec" onClick={load}>다시 불러오기</button>
                  <button type="button" className="btn sec" onClick={onLogout}>로그아웃</button>
                </div>
              </>
            : <span className="mu">내 정보를 불러오는 중…</span>}
        </div>
      </div>
    )
  }

  if (me.needs_profile) {
    return (
      <ProfileSetupScreen
        profile={toProfile(me)}
        saving={saving}
        avatarBusy={busy.includes('avatar')}
        error={saveError ?? uploadError}
        onSubmit={save}
        onPickAvatar={(file) => pick('avatar', file)}
        onLeave={onLogout}
      />
    )
  }

  const profile = toProfile(me)

  if (view === 'create') {
    return <CreateAzitScreen busy={homeBusy} error={createError} onSubmit={create} onCancel={() => { setCreateError(undefined); setView('home') }} />
  }

  const inside = typeof view === 'number' ? azits.find((a) => a.azit_id === view) : undefined
  if (inside) {
    return (
      <AzitScreen
        key={inside.azit_id}
        azit={inside}
        initialRoomId={openRoom}
        me={{ id: me.account_id, nick: inside.my_nickname || me.nickname, color: me.color, src: profile.avatarSrc }}
        onBack={() => { setView('home'); void loadAzits() }}
        onChanged={() => void loadAzits()}
        onLeft={() => { setAzits((as) => as.filter((x) => x.azit_id !== inside.azit_id)); setView('home'); void loadAzits() }}
      />
    )
  }

  const homeAzits: HomeAzit[] = azits.map((a) => ({
    id: String(a.azit_id), name: a.name, short: shortName(a.name), color: a.color, banner: a.banner_url || undefined,
    memberCount: a.member_count, unread: 0, online: 0,
  }))

  return (
    <Shell>
      <Rail
        azits={homeAzits}
        me={{ id: String(me.account_id), nick: me.nickname, color: me.color }}
        meSrc={profile.avatarSrc}
        notifCount={unread}
        view={view === 'me' ? 'me' : view === 'activity' ? 'activity' : 'home'}
        onHome={() => setView('home')}
        onMe={() => setView('me')}
        onBell={openActivity}
        onAzit={(id) => enter(Number(id))}
        onAdd={() => setView('home')}
      />
      {view === 'activity' && (
        <ActivityView items={notifs.map(toNotification)} error={notifError} onReadAll={() => void readAllNotifications().then(loadNotifs).catch((e) => setNotifError(errorText(e)))} onOpen={openNotif} />
      )}
      {view === 'activity' ? null : view !== 'me'
        ? (
          <HomeView
            nickname={me.nickname}
            azits={homeAzits}
            code={code}
            codeError={codeError}
            busy={homeBusy}
            preview={invite && {
              name: invite.azit.name, short: shortName(invite.azit.name), color: invite.azit.color,
              memberCount: invite.azit.member_count, inviter: invite.inviter_nickname, already: invite.already_member,
            }}
            onCode={(c) => { setCode(c); setCodeError(undefined) }}
            onCheck={checkCode}
            onJoin={join}
            onCancelJoin={() => { setInvite(undefined); setCode(''); setCodeError(undefined) }}
            onOpen={(id) => enter(Number(id))}
            onCreate={() => setView('create')}
          />
        )
        : <MeView
            azitShort=""
            profile={profile}
            titles={[]}
            busy={busy}
            uploadError={uploadError}
            onPickImage={pick}
            saving={saving}
            saveError={saveError}
            onSave={save}
          />}
    </Shell>
  )
}
