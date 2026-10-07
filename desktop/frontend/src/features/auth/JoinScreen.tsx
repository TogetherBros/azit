import { AuthLayout } from '../../app/AuthLayout'
import logoSmall from '../../assets/logo-small.webp'
import { Mark } from '../../components/Mark'
import { Swatches } from '../../components/Swatches'

export type JoinStep = 'code' | 'preview' | 'nick'

type Preview = {
  azitName: string
  short: string
  color: string
  desc?: string
  memberCount: number
  rooms: string[]
  from: string // "강백호의 코드"
  expires: string // "6일 남음"
}

type Props = {
  step: JoinStep
  account: string // 오른쪽 위 "나가기 (계정)"
  hasAzits?: boolean // 이미 아지트가 있으면 "돌아가기"
  code?: string
  error?: string
  triesLeft?: number
  preview?: Preview
  nick?: string
  color?: string
}

const STEPS: [JoinStep, string][] = [
  ['code', '코드'],
  ['preview', '미리보기'],
  ['nick', '닉네임'],
]

// 초대 코드로 아지트에 들어가기: 코드 → 미리보기 → 닉네임
// 오류 6종(없음·폐기·만료·다 씀·막힘·꽉 참)은 error 글자로만 구분해 보여 준다.
export function JoinScreen(p: Props) {
  const idx = STEPS.findIndex(([s]) => s === p.step)

  return (
    <AuthLayout blobs="lime">
      <div className="auth-corner left">
        <img src={logoSmall} alt="아지트" />
      </div>
      <div className="auth-corner right">
        {p.hasAzits && <button type="button" className="btn s sec">돌아가기</button>}
        <button type="button" className="btn s sec">나가기 ({p.account})</button>
      </div>

      <div className="auth-body wide">
        <ol className="steps" aria-label="입장 단계">
          {STEPS.map(([s, label], i) => (
            <li key={s} className={`step${i <= idx ? ' on' : ''}`}>{label}</li>
          ))}
        </ol>

        {p.step === 'code' && (
          <form className="col auth-center" style={{ gap: 18 }} noValidate>
            <h1 className="auth-title">초대 코드를<br />받았나요?</h1>
            <p className="mu">아지트는 안에 있는 친구가 준 코드로만 들어가요.</p>
            <label htmlFor="join-code" className="sr">초대 코드 6자리</label>
            <input
              id="join-code"
              className={`inp code-input${p.error ? ' err' : ''}`}
              maxLength={6}
              autoComplete="off"
              defaultValue={p.code}
            />
            {p.error && (
              <span role="alert" className="ferr">
                {p.error}
                {p.triesLeft !== undefined && ` · 남은 시도 ${p.triesLeft}번`}
              </span>
            )}
            <button type="submit" className="btn l p" disabled={(p.code ?? '').length !== 6}>확인</button>
            <button type="button" className="plain lime" style={{ fontSize: 13 }}>코드가 없어요 · 새 아지트 만들기</button>
          </form>
        )}

        {p.step === 'preview' && p.preview && (
          <div className="col" style={{ gap: 16 }}>
            <span className="join-from">{p.preview.from} · {p.preview.expires}</span>
            <div className="card join-card">
              <div className="join-azit">
                <Mark label={p.preview.short} color={p.preview.color} size="lg" />
                <div className="col" style={{ gap: 4, minWidth: 0 }}>
                  <span className="join-azit-name">{p.preview.azitName}</span>
                  <span className="join-azit-desc">{p.preview.desc || '소개 없음'} · {p.preview.memberCount}명</span>
                </div>
              </div>
              <div className="card2 join-rooms">
                <b>방 {p.preview.rooms.length}개</b>
                <div className="row wrap" style={{ gap: 6 }}>
                  {p.preview.rooms.map((r) => <span key={r} className="tag lg">{r}</span>)}
                </div>
              </div>
              <button type="button" className="btn l p">이 아지트에 들어가기</button>
            </div>
            <button type="button" className="plain mu" style={{ alignSelf: 'center', fontSize: 13 }}>다른 코드 넣기</button>
          </div>
        )}

        {p.step === 'nick' && (
          <form className="col" style={{ gap: 18 }} noValidate>
            <h1 className="auth-title sm">여기서 뭐라고<br />불릴래요?</h1>
            <div className="field">
              <label htmlFor="join-nick">닉네임</label>
              <input id="join-nick" className={`inp${p.error ? ' err' : ''}`} maxLength={12} defaultValue={p.nick} />
              {p.error
                ? <span role="alert" className="ferr">{p.error}</span>
                : <span className="hint">아지트마다 다르게 정할 수 있어요</span>}
            </div>
            <div className="field">
              <span className="lbl">내 색</span>
              <Swatches selected={p.color ?? '#D4FF3A'} label="내 색" />
            </div>
            <button type="submit" className="btn l p">들어가기</button>
          </form>
        )}
      </div>
    </AuthLayout>
  )
}
