import type { Azit } from '../../api/types'
import { Icon } from '../../components/Icon'

// 홈에서 보여 줄 아지트 한 칸. 레일에 있는 것보다 조금 더 안다.
export type HomeAzit = Azit & {
  online: number // 지금 들어와 있는 사람
  preview?: string // 마지막 말 "강백호: 라이어 할 사람"
  lastAgo?: string // "3분 전"
  live?: string // 열려 있는 판 "라이어 판 모이는 중"
}

type Props = {
  nickname: string
  azits: HomeAzit[]
  code?: string // 코드 칸에 쳐 둔 것
  codeError?: string
}

// 로그인하고 처음 보는 화면 (디스코드 홈처럼).
// 들어가 있는 아지트를 카드(배너 · 이름 · 인원)로 2x2 씩 보여 주고 넘치면 스크롤. 오른쪽엔 코드로 들어가기 · 새로 만들기.
// 아지트가 하나도 없으면 가운데에 코드 칸만 크게.
export function HomeView({ nickname, azits, code = '', codeError }: Props) {
  if (azits.length === 0) {
    return (
      <div className="main">
        <div className="scroll page home-empty">
          <div className="home-empty-body">
            <h1 className="page-title" style={{ textAlign: 'center' }}>반가워요, {nickname}</h1>
            <p className="mu" style={{ textAlign: 'center', lineHeight: 1.6 }}>
              아직 들어간 아지트가 없어요.<br />친구한테 받은 초대 코드를 넣거나, 우리 아지트를 새로 만들어요.
            </p>
            <JoinCard code={code} error={codeError} big />
          </div>
        </div>
      </div>
    )
  }

  const unread = azits.reduce((n, a) => n + a.unread, 0)
  return (
    <div className="main">
      <div className="page home-page">
        <section className="home-main">
          <div className="col" style={{ gap: 4 }}>
            <h1 className="page-title sm">안녕, {nickname}</h1>
            <span className="mu">
              아지트 {azits.length}곳{unread > 0 ? ` · 안 읽은 말 ${unread > 99 ? '99+' : unread}개` : ' · 다 읽었어요'}
            </span>
          </div>
          <div className="home-grid">
            {azits.map((a) => <AzitCard key={a.id} a={a} />)}
          </div>
        </section>

        <aside className="home-side">
          <JoinCard code={code} error={codeError} />
        </aside>
      </div>
    </div>
  )
}

function AzitCard({ a }: { a: HomeAzit }) {
  return (
    <button type="button" className="home-azit" aria-label={`${a.name} 들어가기`}>
      <span className="home-azit-banner" style={{ background: a.color }}>
        {a.banner
          ? <img src={a.banner} alt="" onError={(e) => { e.currentTarget.style.display = 'none' }} />
          : <span className="home-azit-short">{a.short}</span>}
        {a.unread > 0 && <span className="badge">{a.unread > 99 ? '99+' : a.unread}</span>}
      </span>
      <span className="home-azit-body">
        <span className="home-azit-name">{a.name}</span>
        <span className="home-azit-meta">
          <span className={`home-dot${a.online ? '' : ' off'}`} aria-hidden="true" />
          {a.memberCount}명{a.online ? ` · ${a.online}명 접속` : ''}
        </span>
      </span>
    </button>
  )
}

// 초대 코드 6자리 칸 + 새 아지트 만들기
function JoinCard({ code, error, big }: { code: string; error?: string; big?: boolean }) {
  return (
    <div className={`card col home-join${big ? ' big' : ''}`}>
      <form className="col" style={{ gap: 10 }} noValidate onSubmit={(e) => e.preventDefault()}>
        <label htmlFor="home-code" className="b">초대 코드로 들어가기</label>
        <input
          id="home-code"
          className={`inp code-input${big ? '' : ' sm'}${error ? ' err' : ''}`}
          maxLength={6}
          autoComplete="off"
          spellCheck={false}
          placeholder="MK7Q2X"
          defaultValue={code}
        />
        {error
          ? <span role="alert" className="ferr">{error}</span>
          : <span className="hint">아지트 안 친구가 준 6자리예요</span>}
        <button type="submit" className="btn p" disabled={code.length !== 6}>확인</button>
      </form>
      <div className="home-or"><span>또는</span></div>
      <button type="button" className="btn gh">
        <Icon name="plus" />
        새 아지트 만들기
      </button>
    </div>
  )
}
