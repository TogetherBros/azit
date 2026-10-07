import type { Member, Stats, Title } from '../../api/types'
import { Avatar } from '../../components/Avatar'
import { Stat, Toggle } from '../../components/Controls'
import { Swatches } from '../../components/Swatches'

type Props = {
  azitShort: string
  me: Member
  email: string // 구글 계정
  stats: Stats
  titles: Title[]
  dnd?: boolean
  nickError?: string
}

const RARE_CLASS = { 일반: 'rare-common', 희귀: 'rare-rare', 전설: 'rare-legend' } as const

// 내 정보: 이 아지트에서 보이는 내 모습 · 닉네임/색 · 계정 | 칭호 도감
export function MeView({ azitShort, me, email, stats, titles, dnd, nickError }: Props) {
  const unlocked = titles.filter((t) => t.unlocked).length
  return (
    <div className="main">
      <div className="scroll page me-page">
        <section className="me-left">
          <span className="mu b" style={{ fontSize: 12 }}>{azitShort} 멤버에게 이렇게 보여요</span>
          <div className="profile">
            <div className="profile-cover" style={{ background: me.color }}>
              <span className="profile-av"><Avatar name={me.nick} color="var(--s3)" size={72} /></span>
            </div>
            <div className="profile-body">
              <span className="profile-name">{me.nick}</span>
              <span className="profile-id">{email}</span>
              {me.title
                ? <span className="tag title" style={{ background: me.color }}>{me.title}</span>
                : <span className="hint">칭호 없음 · 아래에서 골라요</span>}
              <div className="grid3" style={{ gap: 8 }}>
                <Stat value={stats.games} label="판" />
                <Stat value={stats.catches} label="라이어 검거" />
                <Stat value={stats.shakes} label="흔들기" />
              </div>
            </div>
          </div>

          <form className="card col" noValidate>
            <div className="field">
              <label htmlFor="me-nick">이 아지트 닉네임</label>
              <input id="me-nick" className={`inp${nickError ? ' err' : ''}`} maxLength={12} defaultValue={me.nick} />
              {nickError && <span className="ferr">{nickError}</span>}
            </div>
            <div className="field">
              <span className="lbl">내 색</span>
              <Swatches selected={me.color} label="내 색" />
            </div>
            <button type="submit" className="btn p" disabled>닉네임 저장</button>
          </form>

          <div className="card col">
            <span className="b">계정 · {email}</span>
            <div className="row">
              <div className="grow col" style={{ gap: 2 }}>
                <span className="b" style={{ fontSize: 13 }}>방해 금지</span>
                <span className="hint">흔들려도 창이 안 흔들리고 알림만 쌓여요</span>
              </div>
              <Toggle on={!!dnd} label="방해 금지" />
            </div>
            <button type="button" className="btn dg">로그아웃</button>
          </div>
        </section>

        <section className="me-right">
          <div className="row">
            <h1 className="page-title grow" style={{ fontSize: 36 }}>도감</h1>
            <span className="mu" style={{ fontSize: 13 }}>{unlocked}/{titles.length} · {azitShort} 기준</span>
          </div>
          <div className="grid3">
            {titles.map((t) => {
              const equipped = me.title === t.name
              const cls = !t.unlocked ? ' locked' : equipped ? ' equipped' : ''
              return (
                <button
                  key={t.name}
                  type="button"
                  className={`title-card${cls}`}
                  style={equipped ? { background: me.color } : undefined}
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
        </section>
      </div>
    </div>
  )
}
