import { useState } from 'react'
import { NAME_RULE, type ProfileFields } from '../../api/profile'
import { AuthLayout } from '../../app/AuthLayout'
import logoSmall from '../../assets/logo-small.webp'
import { Avatar } from '../../components/Avatar'
import { ProfileCard, type Profile } from '../../components/ProfileCard'
import { Swatches } from '../../components/Swatches'
import { NAME_MSG, useNameCheck, type NameState } from '../../lib/useNameCheck'

export type { NameState }

type Props = {
  account?: string // 오른쪽 위 "나가기 (구글 계정)"
  profile: Profile // 색과 사진. 고치면 오른쪽 출입증이 바로 따라 바뀐다
  nameState?: NameState // 미리보기용: 이름 확인 결과를 정해서 보여 주기
  saving?: boolean
  avatarBusy?: boolean
  error?: string // 저장·사진 실패
  onSubmit?: (f: ProfileFields) => void // 없으면 미리보기용
  onPickAvatar?: (file: File) => void
  onLeave?: () => void
}

type Draft = Pick<Profile, 'nickname' | 'name' | 'color'>

const ACCEPT = 'image/png,image/jpeg,image/webp'

// 구글로 처음 들어온 사람: 계정 프로필 만들기 (회원가입).
// 여기서 만든 프로필이 모든 아지트에서 그대로 보인다. 아지트는 그다음 홈에서 코드로 들어간다.
export function ProfileSetupScreen({ account, profile, nameState, saving, avatarBusy, error, onSubmit, onPickAvatar, onLeave }: Props) {
  // 닉네임·이름은 빈칸에서 시작한다 (구글 이름을 미리 넣지 않는다)
  const [d, setD] = useState<Draft>({ nickname: '', name: '', color: profile.color })
  const set = <K extends keyof Draft>(k: K, v: Draft[K]) => setD((cur) => ({ ...cur, [k]: v }))
  const p: Profile = { ...profile, ...d }

  const checked = useNameCheck(onSubmit ? d.name : '')
  const state = nameState ?? checked

  const nameBad = state === 'taken' || state === 'invalid'
  const nickname = d.nickname.trim()
  const nameOk = NAME_RULE.test(d.name) && !nameBad && state !== 'checking'
  const canSubmit = !!nickname && nameOk && !saving

  const submit = () => {
    if (canSubmit) onSubmit?.({ name: d.name, nickname, bio: '', color: d.color })
  }

  return (
    <AuthLayout blobs="lime">
      <div className="auth-corner left">
        <img src={logoSmall} alt="아지트" />
      </div>
      <div className="auth-corner right">
        <button type="button" className="btn s sec" onClick={onLeave}>나가기{account ? ` (${account})` : ''}</button>
      </div>

      <div className="auth-body setup">
        <h1 className="auth-title sm">프로필부터 만들어요</h1>

        <div className="setup-grid">
          <form className="card col setup-form" noValidate onSubmit={(e) => { e.preventDefault(); submit() }}>
            <div className="row" style={{ gap: 14 }}>
              <Avatar name={p.nickname || '?'} color={p.color} size={56} src={p.avatarSrc} />
              <div className="col" style={{ gap: 4 }}>
                <span className="lbl">프로필 사진 <span className="lbl-sub">· 안 넣어도 돼요</span></span>
                <AvatarButton has={!!p.avatarSrc} busy={avatarBusy} onPick={onPickAvatar} />
              </div>
            </div>

            <div className="field">
              <label htmlFor="setup-nick">닉네임 <span className="lbl-sub">· 크게 보여요, 겹쳐도 돼요</span></label>
              <input
                id="setup-nick"
                className="inp"
                maxLength={16}
                value={d.nickname}
                onChange={(e) => set('nickname', e.target.value)}
              />
            </div>

            <div className="field">
              <label htmlFor="setup-name">이름 <span className="lbl-sub">· 나만 쓰는 이름, @ 뒤에 붙어요</span></label>
              <div className={`inp name-inp${nameBad ? ' err' : ''}`}>
                <span className="mu">@</span>
                <input
                  id="setup-name"
                  maxLength={20}
                  autoComplete="off"
                  spellCheck={false}
                  value={d.name}
                  onChange={(e) => set('name', e.target.value.toLowerCase())}
                />
              </div>
              <span
                role={nameBad ? 'alert' : undefined}
                className={nameBad ? 'ferr' : state === 'ok' ? 'hint ok' : 'hint'}
              >
                {NAME_MSG[state]}
              </span>
            </div>

            <div className="field">
              <span className="lbl">내 색 <span className="lbl-sub">· 사진이 없을 때 이 색으로 보여요</span></span>
              <Swatches selected={d.color} label="내 색" onPick={(c) => set('color', c)} />
            </div>

            {error && <div role="alert" className="ferr">{error}</div>}
            <button type="submit" className="btn p" disabled={!canSubmit}>{saving ? '만드는 중...' : '시작하기'}</button>
          </form>

          <aside className="col setup-preview" style={{ gap: 10 }}>
            <span className="me-cap">친구들에게 이렇게 보여요</span>
            <ProfileCard p={p} />
            <span className="hint">배너와 소개는 들어간 뒤 내 정보에서 꾸밀 수 있어요</span>
          </aside>
        </div>
      </div>
    </AuthLayout>
  )
}

function AvatarButton({ has, busy, onPick }: { has: boolean; busy?: boolean; onPick?: (file: File) => void }) {
  const text = busy ? '올리는 중…' : has ? '바꾸기' : '올리기'
  if (!onPick) return <button type="button" className="btn s sec" style={{ alignSelf: 'flex-start' }}>{text}</button>
  return (
    <label className={`btn s sec${busy ? ' disabled' : ''}`} style={{ alignSelf: 'flex-start' }} aria-disabled={!!busy}>
      {text}
      <input
        type="file"
        accept={ACCEPT}
        hidden
        disabled={!!busy}
        onChange={(e) => {
          const f = e.target.files?.[0]
          e.target.value = ''
          if (f) onPick(f)
        }}
      />
    </label>
  )
}
