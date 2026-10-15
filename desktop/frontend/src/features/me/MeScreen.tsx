import { useCallback, useEffect, useState } from 'react'
import {
  getMe, profileErrorText, toProfile, updateMe, uploadImage,
  type ImageKind, type Me, type ProfileFields,
} from '../../api/profile'
import { Rail } from '../../app/Rail'
import { Shell } from '../../app/Shell'
import { shrinkImage } from '../../lib/image'
import { HomeView } from '../home/HomeView'
import { ProfileSetupScreen } from '../auth/ProfileSetupScreen'
import { MeView } from './MeView'

type Props = { onLogout: () => void }

// 로그인한 뒤의 진짜 화면.
// 서버에서 내 정보를 받아, 프로필이 없으면 프로필 만들기, 있으면 레일 + 홈(처음) 또는 내 정보를 보여 준다.
// 아지트 API 가 아직 없어서 홈은 늘 "아직 들어간 아지트가 없어요" 상태다.
// 저장·사진 올리기는 여기서 처리하고, 화면 파일들은 그리기만 한다.
export function MeScreen({ onLogout }: Props) {
  const [me, setMe] = useState<Me>()
  const [loadError, setLoadError] = useState<string>()
  const [busy, setBusy] = useState<ImageKind[]>([]) // 배너·프사는 따로 올릴 수 있다
  const [uploadError, setUploadError] = useState<string>()
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState<string>()
  const [view, setView] = useState<'home' | 'me'>('home')

  const load = useCallback(() => {
    getMe()
      .then((m) => { setMe(m); setLoadError(undefined) })
      .catch((e) => setLoadError(profileErrorText(e)))
  }, [])

  useEffect(load, [load])

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
  return (
    <Shell>
      <Rail
        azits={[]}
        me={{ id: String(me.account_id), nick: me.nickname, color: me.color }}
        meSrc={profile.avatarSrc}
        notifCount={0}
        view={view}
        onHome={() => setView('home')}
        onMe={() => setView('me')}
      />
      {view === 'home'
        ? <HomeView nickname={me.nickname} azits={[]} />
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
