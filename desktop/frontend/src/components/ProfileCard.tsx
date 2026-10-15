import { Avatar } from './Avatar'

export type Profile = {
  nickname: string // 닉네임 (겹쳐도 됨, 크게 보임)
  name: string // 이름 (중복 불가, @ 와 함께 보임)
  color: string // 내 색: 배너와 사진 바탕, 이름, 칭호 도장
  bio?: string
  title?: string // 장착한 칭호
  avatarSrc?: string
  bannerSrc?: string
  since?: string // 처음 들어온 날 "2026.10.14"
}

// 아지트 출입증. 다른 사람에게 보이는 내 프로필이고, 모든 아지트에서 똑같이 보인다.
// 배너 · 비스듬히 붙인 사진 스티커 · 칭호 도장 · 말풍선 소개 · 뜯는 선 아래 가입일
export function ProfileCard({ p }: { p: Profile }) {
  return (
    <article className="pass" style={{ ['--me' as string]: p.color }} aria-label={`${p.nickname || '닉네임'} 출입증`}>
      <div className="pass-top">
        <span>AZIT PASS</span>
        <span className="pass-hole" aria-hidden="true" />
        <span>{p.since ?? 'NEW'}</span>
      </div>

      <div className="pass-window">
        {p.bannerSrc && <img src={p.bannerSrc} alt="" onError={(e) => { e.currentTarget.style.display = 'none' }} />}
      </div>
      <span className="pass-av">
        <Avatar name={p.nickname || '?'} color={p.color} size={76} src={p.avatarSrc} />
      </span>

      <div className="pass-body">
        <span className="pass-nick">{p.nickname || '닉네임'}</span>
        {p.name
          ? <span className="pass-name">@{p.name}</span>
          : <span className="hint">이름을 정해 주세요</span>}
        {p.title && <span className="pass-stamp">{p.title}</span>}
        {p.bio && <p className="pass-bio">{p.bio}</p>}
      </div>

      <div className="pass-foot">
        <span>{p.since ? `${p.since} 부터 아지트` : '오늘 처음 와요'}</span>
        <span className="pass-mark" aria-hidden="true">⌂</span>
      </div>
    </article>
  )
}
