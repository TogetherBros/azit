import type { Title } from '../../api/types'
import { Avatar } from '../../components/Avatar'
import { Toggle } from '../../components/Controls'
import { ProfileCard, type Profile } from '../../components/ProfileCard'
import { Swatches } from '../../components/Swatches'

type Props = {
  azitShort: string
  profile: Profile
  email: string // 구글 계정
  titles: Title[]
  dnd?: boolean
  dirty?: boolean // 고친 게 있으면 저장 버튼이 켜진다
  nameError?: string
}

const RARE_CLASS = { 일반: 'rare-common', 희귀: 'rare-rare', 전설: 'rare-legend' } as const
const BIO_MAX = 190

// 내 정보: (왼쪽) 보이는 모습 · 계정 | (오른쪽) 프로필 고치기 · 칭호 도감
// 프로필은 계정에 하나. 모든 아지트에서 똑같이 보인다.
export function MeView({ azitShort, profile: p, email, titles, dnd, dirty, nameError }: Props) {
  const unlocked = titles.filter((t) => t.unlocked).length
  return (
    <div className="main">
      <div className="scroll page me-page">
        <section className="me-left">
          <span className="mu b" style={{ fontSize: 12 }}>내 출입증 · 모든 아지트에서 이렇게 보여요</span>
          <ProfileCard p={p} />

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
          <h1 className="page-title" style={{ fontSize: 36 }}>프로필</h1>
          <form className="card col" style={{ gap: 18 }} noValidate>
            <div className="field">
              <span className="lbl">창문</span>
              <div className="banner-pick" style={{ background: p.color }}>
                {p.bannerSrc && <img src={p.bannerSrc} alt="" />}
                <button type="button" className="btn s sec">{p.bannerSrc ? '바꾸기' : '사진 올리기'}</button>
                {p.bannerSrc && <button type="button" className="btn s sec">지우기</button>}
              </div>
              <span className="hint">출입증 위 집 모양 창문이에요. 사진이 없으면 내 색으로 칠해져요</span>
            </div>

            <div className="field">
              <span className="lbl">프로필 사진</span>
              <div className="row" style={{ gap: 14 }}>
                <Avatar name={p.nickname} color={p.color} size={56} src={p.avatarSrc} />
                <button type="button" className="btn s sec">{p.avatarSrc ? '바꾸기' : '사진 올리기'}</button>
                {p.avatarSrc && <button type="button" className="btn s sec">지우기</button>}
              </div>
            </div>

            <div className="grid2">
              <div className="field">
                <label htmlFor="me-nick">닉네임</label>
                <input id="me-nick" className="inp" maxLength={16} defaultValue={p.nickname} />
              </div>
              <div className="field">
                <label htmlFor="me-name">이름</label>
                <div className={`inp name-inp${nameError ? ' err' : ''}`}>
                  <span className="mu">@</span>
                  <input id="me-name" maxLength={20} autoComplete="off" spellCheck={false} defaultValue={p.name} />
                </div>
                {nameError && <span role="alert" className="ferr">{nameError}</span>}
              </div>
            </div>

            <div className="field">
              <div className="row">
                <label htmlFor="me-bio">자기소개</label>
                <span className="count">{(p.bio ?? '').length}/{BIO_MAX}</span>
              </div>
              <textarea id="me-bio" className="inp" rows={3} maxLength={BIO_MAX} defaultValue={p.bio} placeholder="한 줄이면 충분해요" />
            </div>

            <div className="field">
              <span className="lbl">내 색</span>
              <Swatches selected={p.color} label="내 색" />
            </div>

            <div className="row end">
              {dirty && <button type="button" className="btn sec">되돌리기</button>}
              <button type="submit" className="btn p" disabled={!dirty}>저장</button>
            </div>
          </form>

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
        </section>
      </div>
    </div>
  )
}
