import { useEffect, useState } from 'react'
import type { ImageKind, ProfileFields } from '../../api/profile'
import type { Title } from '../../api/types'
import { Avatar } from '../../components/Avatar'
import { ProfileCard, type Profile } from '../../components/ProfileCard'
import { Swatches } from '../../components/Swatches'
import { NAME_MSG, useNameCheck } from '../../lib/useNameCheck'

type Props = {
  azitShort: string
  profile: Profile
  titles: Title[]
  dirty?: boolean // 미리보기용: 고친 상태로 보여 주기
  nameError?: string
  busy?: ImageKind[] // 올리는 중인 사진들
  uploadError?: string
  onPickImage?: (kind: ImageKind, file: File) => void // 없으면 미리보기용 (버튼만 보임)
  saving?: boolean
  saveError?: string
  onSave?: (f: ProfileFields) => void // 없으면 미리보기용
}

// 고칠 수 있는 글자·색. 사진은 고르면 바로 올라가서 여기 없다.
type Draft = Pick<Profile, 'nickname' | 'name' | 'bio' | 'color'>

const RARE_CLASS = { 일반: 'rare-common', 희귀: 'rare-rare', 전설: 'rare-legend' } as const
const BIO_MAX = 190

const draftOf = (p: Profile): Draft => ({ nickname: p.nickname, name: p.name, bio: p.bio ?? '', color: p.color })

// 내 정보: (왼쪽) 다른 사람에게 보이는 모습 | (오른쪽) 프로필 고치기 · 칭호 도감
// 고치는 동안 왼쪽 출입증이 바로 따라 바뀐다. 프로필은 계정에 하나, 모든 아지트에서 똑같다.
export function MeView({ azitShort, profile, titles, dirty, nameError, busy, uploadError, onPickImage, saving, saveError, onSave }: Props) {
  const [d, setD] = useState<Draft>(() => draftOf(profile))

  // 저장해서 서버 값이 바뀌면 고치던 칸도 그 값으로 맞춘다. 사진만 바뀐 건 상관없다.
  const saved = JSON.stringify(draftOf(profile))
  useEffect(() => setD(JSON.parse(saved) as Draft), [saved])

  const nameState = useNameCheck(onSave ? d.name : '', profile.name)
  const nameMsg = nameError ?? (nameState === 'taken' || nameState === 'invalid' ? NAME_MSG[nameState] : undefined)

  const set = <K extends keyof Draft>(k: K, v: Draft[K]) => setD((cur) => ({ ...cur, [k]: v }))
  // 사진은 서버에서 받은 것, 글자·색은 고치는 중인 것 (사진을 올려도 고치던 글자는 그대로)
  const p: Profile = { ...profile, ...d, bio: d.bio || undefined }
  const changed = dirty || JSON.stringify(d) !== JSON.stringify(draftOf(profile))
  const unlocked = titles.filter((t) => t.unlocked).length
  const canSave = changed && !!d.nickname.trim() && !nameMsg && nameState !== 'checking' && !saving

  const save = () => {
    if (!canSave || !onSave) return
    const f = { name: d.name, nickname: d.nickname.trim(), bio: (d.bio ?? '').trim(), color: d.color }
    setD(f)
    onSave(f)
  }

  return (
    <div className="main">
      <div className="scroll page me-page">
        <section className="me-left">
          <span className="me-cap">다른 사람에게는 이렇게 보여요</span>
          <ProfileCard p={p} />
        </section>

        <section className="me-right">
          <form className="card col me-form" noValidate onSubmit={(e) => { e.preventDefault(); save() }}>
            <div className="grid2">
              <div className="field">
                <span className="lbl">배너</span>
                <div className="row">
                  <div className="banner-pick" style={{ background: p.color }}>
                    {p.bannerSrc && <img src={p.bannerSrc} alt="" onError={hideBroken} />}
                  </div>
                  <PickButton kind="banner" has={!!p.bannerSrc} busy={busy?.includes('banner')} onPick={onPickImage} />
                </div>
              </div>
              <div className="field">
                <span className="lbl">프로필 사진</span>
                <div className="row">
                  <Avatar name={p.nickname || '?'} color={p.color} size={56} src={p.avatarSrc} />
                  <PickButton kind="avatar" has={!!p.avatarSrc} busy={busy?.includes('avatar')} onPick={onPickImage} />
                </div>
              </div>
            </div>
            {uploadError && <span role="alert" className="ferr">{uploadError}</span>}

            <div className="grid2">
              <div className="field">
                <label htmlFor="me-nick">닉네임</label>
                <input id="me-nick" className="inp" maxLength={16} value={d.nickname} onChange={(e) => set('nickname', e.target.value)} />
              </div>
              <div className="field">
                <label htmlFor="me-name">이름</label>
                <div className={`inp name-inp${nameMsg ? ' err' : ''}`}>
                  <span className="mu">@</span>
                  <input
                    id="me-name"
                    maxLength={20}
                    autoComplete="off"
                    spellCheck={false}
                    value={d.name}
                    onChange={(e) => set('name', e.target.value.toLowerCase())}
                  />
                </div>
                {nameMsg && <span role="alert" className="ferr">{nameMsg}</span>}
              </div>
            </div>

            <div className="field">
              <div className="row">
                <label htmlFor="me-bio">자기소개</label>
                <span className="count">{(d.bio ?? '').length}/{BIO_MAX}</span>
              </div>
              <textarea
                id="me-bio"
                className="inp"
                rows={2}
                maxLength={BIO_MAX}
                value={d.bio}
                onChange={(e) => set('bio', e.target.value)}
                placeholder="한 줄이면 충분해요"
              />
            </div>

            {saveError && <span role="alert" className="ferr">{saveError}</span>}
            <div className="row me-form-foot">
              <div className="field grow">
                <span className="lbl">내 색</span>
                <Swatches selected={d.color} label="내 색" onPick={(c) => set('color', c)} />
              </div>
              {changed && !saving && <button type="button" className="btn sec" onClick={() => setD(draftOf(profile))}>되돌리기</button>}
              <button type="submit" className="btn p" disabled={onSave ? !canSave : !changed}>{saving ? '저장 중…' : '저장'}</button>
            </div>
          </form>

          {titles.length > 0 && <>
          <div className="row" style={{ marginTop: 10 }}>
            <h2 className="page-title grow" style={{ fontSize: 28 }}>도감</h2>
            <span className="mu" style={{ fontSize: 13 }}>{unlocked}/{titles.length} · {azitShort} 기준</span>
          </div>
          <div className="grid3">
            {titles.map((t) => {
              const equipped = p.title === t.name
              const cls = !t.unlocked ? ' locked' : equipped ? ' equipped' : ''
              return (
              <button
                key={t.name}
                type="button"
                className={`title-card${cls}`}
                style={equipped ? { background: p.color } : undefined}
                disabled={!t.unlocked}
                aria-pressed={equipped}
              >
                <span className={`tag ${equipped ? '' : RARE_CLASS[t.rarity]}`}>{t.rarity}{equipped ? ' · 장착' : ''}</span>
                <span className="title-name">{t.unlocked ? t.name : '???'}</span>
                <span className="title-cond">{t.cond}</span>
                {!t.unlocked && <span className="progress"><span style={{ width: `${Math.round(t.progress * 100)}%` }} /></span>}
              </button>
              )
            })}
          </div>
          </>}
        </section>
      </div>
    </div>
  )
}

const ACCEPT = 'image/png,image/jpeg,image/webp'

// 사진 주소가 만료됐거나 못 불러오면 깨진 아이콘 대신 바탕색만 보이게
const hideBroken = (e: { currentTarget: HTMLImageElement }) => {
  e.currentTarget.style.display = 'none'
}

// 사진 올리기 버튼. onPick 이 있으면 누를 때 파일 고르기 창이 뜬다.
function PickButton({ kind, has, busy, onPick }: {
  kind: ImageKind
  has: boolean
  busy?: boolean // 이 사진을 올리는 중
  onPick?: (kind: ImageKind, file: File) => void
}) {
  const text = busy ? '올리는 중…' : has ? '바꾸기' : '올리기'
  if (!onPick) return <button type="button" className="btn s sec">{text}</button>
  return (
    <label className={`btn s sec${busy ? ' disabled' : ''}`} aria-disabled={!!busy}>
      {text}
      <input
        type="file"
        accept={ACCEPT}
        hidden
        disabled={!!busy}
        onChange={(e) => {
          const f = e.target.files?.[0]
          e.target.value = '' // 같은 파일을 다시 골라도 반응하게
          if (f) onPick(kind, f)
        }}
      />
    </label>
  )
}
