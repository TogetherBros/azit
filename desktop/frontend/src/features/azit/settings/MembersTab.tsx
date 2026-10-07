import type { Member } from '../../../api/types'
import { Avatar } from '../../../components/Avatar'

type Props = {
  members: Member[]
  meId: string
  bans: { id: string; name: string }[] // 다시 못 들어오게 막은 사람
}

// 멤버: 찾기 · 목록(내보내기) · 막은 사람 풀기
export function MembersTab({ members, meId, bans }: Props) {
  return (
    <>
      <div className="row">
        <h1 className="page-title sm grow">멤버 {members.length}/50</h1>
        <label htmlFor="sm-q" className="sr">멤버 찾기</label>
        <input id="sm-q" className="inp member-search" placeholder="이름이나 아이디" />
      </div>
      <div className="col">
        {members.map((m) => {
          const self = m.id === meId
          return (
            <div key={m.id} className="lrow">
              <button type="button" className="plain" aria-label={`${m.nick} 프로필`}>
                <Avatar name={m.nick} color={m.color} size={36} />
              </button>
              <div className="grow col" style={{ gap: 2 }}>
                <span className="b">
                  {m.nick}
                  {self && <> <span className="tag lime">나</span></>}
                  {m.title && <span className="msg-title">{m.title}</span>}
                </span>
                <span className="member-sub"><span className="mono">{m.id}</span> · {m.via} · {m.joinedAgo}</span>
              </div>
              <span className="mu hide-md" style={{ fontSize: 11 }}>{m.online ? '접속 중' : ''}</span>
              {!self && <button type="button" className="btn s sec">내보내기</button>}
            </div>
          )
        })}
      </div>
      {bans.length > 0 && (
        <div className="col">
          <span className="b">다시 못 들어오는 사람</span>
          {bans.map((b) => (
            <div key={b.id} className="lrow">
              <span className="grow">{b.name} <span className="mono mu" style={{ fontSize: 11 }}>{b.id}</span></span>
              <button type="button" className="btn s sec">풀기</button>
            </div>
          ))}
        </div>
      )}
    </>
  )
}
