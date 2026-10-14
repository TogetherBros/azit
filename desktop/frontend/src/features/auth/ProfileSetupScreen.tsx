import { AuthLayout } from '../../app/AuthLayout'
import logoSmall from '../../assets/logo-small.webp'
import { Avatar } from '../../components/Avatar'
import { ProfileCard, type Profile } from '../../components/ProfileCard'
import { Swatches } from '../../components/Swatches'

export type NameState = 'idle' | 'checking' | 'ok' | 'taken' | 'invalid'

type Props = {
  account: string // 오른쪽 위 "나가기 (구글 계정)"
  profile: Profile
  nameState?: NameState
  nicknameError?: string
  saving?: boolean
  error?: string // 저장 실패 등
}

const NAME_MSG: Record<NameState, string> = {
  idle: '영문 소문자, 숫자, 밑줄(_)만 써요 · 5~20자',
  checking: '확인하는 중...',
  ok: '쓸 수 있는 이름이에요',
  taken: '이미 누가 쓰고 있어요',
  invalid: '영문 소문자, 숫자, 밑줄(_)만 쓸 수 있어요. 5~20자 사이여야 해요',
}

// 구글로 처음 들어온 사람: 계정 프로필 만들기.
// 여기서 만든 프로필이 모든 아지트에서 그대로 보인다. 아지트는 그다음에 초대 코드로 들어간다.
export function ProfileSetupScreen({ account, profile: p, nameState = 'idle', nicknameError, saving, error }: Props) {
  const nameBad = nameState === 'taken' || nameState === 'invalid'
  const canSubmit = !!p.nickname && nameState === 'ok' && !nicknameError && !saving

  return (
    <AuthLayout blobs="lime">
      <div className="auth-corner left">
        <img src={logoSmall} alt="아지트" />
      </div>
      <div className="auth-corner right">
        <button type="button" className="btn s sec">나가기 ({account})</button>
      </div>

      <div className="auth-body setup">
        <h1 className="auth-title sm">프로필부터<br />만들어요</h1>

        <div className="setup-grid">
          <form className="card col setup-form" style={{ gap: 18 }} noValidate>
            <div className="field">
              <span className="lbl">프로필 사진 <span className="lbl-sub">· 안 넣어도 돼요</span></span>
              <div className="row" style={{ gap: 14 }}>
                <Avatar name={p.nickname || '?'} color={p.color} size={64} src={p.avatarSrc} />
                <div className="row" style={{ gap: 6 }}>
                  <button type="button" className="btn s sec">{p.avatarSrc ? '바꾸기' : '사진 올리기'}</button>
                  {p.avatarSrc && <button type="button" className="btn s sec">지우기</button>}
                </div>
              </div>
            </div>

            <div className="field">
              <label htmlFor="setup-nick">닉네임</label>
              <input id="setup-nick" className={`inp${nicknameError ? ' err' : ''}`} maxLength={16} defaultValue={p.nickname} />
              {nicknameError
                ? <span role="alert" className="ferr">{nicknameError}</span>
                : <span className="hint">채팅에서 크게 보여요. 한글도 되고, 겹쳐도 괜찮아요</span>}
            </div>

            <div className="field">
              <label htmlFor="setup-name">이름 <span className="lbl-sub">· 나만 쓰는 이름, @ 뒤에 붙어요</span></label>
              <div className={`inp name-inp${nameBad ? ' err' : ''}`}>
                <span className="mu">@</span>
                <input id="setup-name" maxLength={20} autoComplete="off" spellCheck={false} defaultValue={p.name} />
              </div>
              <span
                role={nameBad ? 'alert' : undefined}
                className={nameBad ? 'ferr' : nameState === 'ok' ? 'hint ok' : 'hint'}
              >
                {NAME_MSG[nameState]}
              </span>
            </div>

            <div className="field">
              <span className="lbl">내 색</span>
              <Swatches selected={p.color} label="내 색" />
              <span className="hint">창문과 사진이 없을 때 이 색으로 보여요</span>
            </div>

            {error && <div role="alert" className="ferr">{error}</div>}
            <button type="submit" className="btn l p" disabled={!canSubmit}>{saving ? '만드는 중...' : '시작하기'}</button>
          </form>

          <aside className="col setup-preview" style={{ gap: 10 }}>
            <span className="mu b" style={{ fontSize: 12 }}>내 출입증 · 친구들에게 이렇게 보여요</span>
            <ProfileCard p={p} />
            <span className="hint">창문 사진과 소개는 들어간 뒤 내 정보에서 꾸밀 수 있어요.</span>
          </aside>
        </div>
      </div>
    </AuthLayout>
  )
}
