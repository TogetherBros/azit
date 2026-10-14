import { useCallback, useEffect, useState } from 'react'
import { getMe, profileErrorText, toProfile, uploadImage, type ImageKind, type Me } from '../../api/profile'
import { shrinkImage } from '../../lib/image'
import { MeView } from './MeView'

// 진짜 내 정보 화면: 서버에서 프로필을 받아 MeView 에 넣고, 사진 올리기를 처리한다.
// 이름·닉네임 저장 API 는 아직 없어서 지금은 사진만 진짜로 바뀐다.
export function MeScreen() {
  const [me, setMe] = useState<Me>()
  const [loadError, setLoadError] = useState<string>()
  const [busy, setBusy] = useState<ImageKind>()
  const [uploadError, setUploadError] = useState<string>()

  const load = useCallback(() => {
    getMe()
      .then((m) => { setMe(m); setLoadError(undefined) })
      .catch((e) => setLoadError(profileErrorText(e)))
  }, [])

  useEffect(load, [load])

  const pick = async (kind: ImageKind, file: File) => {
    setBusy(kind)
    setUploadError(undefined)
    try {
      const img = await shrinkImage(file, kind)
      await uploadImage(kind, img.contentType, img.base64)
      load()
    } catch (e) {
      setUploadError(profileErrorText(e))
    } finally {
      setBusy(undefined)
    }
  }

  if (!me) {
    return (
      <div className="main">
        <div className="scroll page col" style={{ alignItems: 'center', justifyContent: 'center', gap: 12 }}>
          {loadError
            ? <>
                <span role="alert" className="ferr">{loadError}</span>
                <button type="button" className="btn sec" onClick={load}>다시 불러오기</button>
              </>
            : <span className="mu">내 정보를 불러오는 중…</span>}
        </div>
      </div>
    )
  }

  return (
    <MeView
      azitShort=""
      profile={toProfile(me)}
      email=""
      titles={[]}
      busy={busy}
      uploadError={uploadError}
      onPickImage={pick}
    />
  )
}
