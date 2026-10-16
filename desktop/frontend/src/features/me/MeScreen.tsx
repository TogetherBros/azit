import { useCallback, useEffect, useState } from 'react'
import {
  getMe, profileErrorText, toProfile, updateMe, uploadImage,
  type ImageKind, type Me, type ProfileFields,
} from '../../api/profile'
import {
  createAzit, createRoom, errorText, joinInvite, listAzits, previewInvite,
  type AzitRes, type InvitePreview,
} from '../../api/server'
import { Rail } from '../../app/Rail'
import { Shell } from '../../app/Shell'
import { shrinkImage } from '../../lib/image'
import { AzitScreen, shortName } from '../azit/AzitScreen'
import { CreateAzitScreen, type NewAzit } from '../azit/CreateAzitScreen'
import { HomeView, type HomeAzit } from '../home/HomeView'
import { ProfileSetupScreen } from '../auth/ProfileSetupScreen'
import { MeView } from './MeView'

type Props = { onLogout: () => void }

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
  const [view, setView] = useState<'home' | 'me' | 'create' | number>('home') // number = 그 아지트 안
  const [azits, setAzits] = useState<AzitRes[]>([])
  const [code, setCode] = useState('')
  const [codeError, setCodeError] = useState<string>()
  const [invite, setInvite] = useState<InvitePreview>()
  const [homeBusy, setHomeBusy] = useState(false)
  const [createError, setCreateError] = useState<string>()

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

  const enter = (id: number) => {
    setInvite(undefined); setCode(''); setCodeError(undefined)
    setView(id)
  }

  // 새 아지트: 만들고 고른 기본 방도 같이 만든 뒤 그 아지트로
  const create = async (a: NewAzit) => {
    setHomeBusy(true)
    try {
      const made = await createAzit(a.name, a.color)
      for (const p of a.presets) await createRoom(made.azit_id, p).catch(() => {})
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
        azit={inside}
        me={{ id: me.account_id, nick: inside.my_nickname || me.nickname, color: me.color, src: profile.avatarSrc }}
        onBack={() => { setView('home'); void loadAzits() }}
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
        notifCount={0}
        view={view === 'me' ? 'me' : 'home'}
        onHome={() => setView('home')}
        onMe={() => setView('me')}
        onAzit={(id) => enter(Number(id))}
        onAdd={() => setView('home')}
      />
      {view !== 'me'
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
