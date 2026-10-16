import { useState } from 'react'
import type { Member } from '../../../api/types'
import { Avatar } from '../../../components/Avatar'
import { Modal } from '../../../components/Modal'

export type MemberRole = 'super_admin' | 'admin' | 'member'

const RANK: Record<MemberRole, number> = { member: 1, admin: 2, super_admin: 3 }
const ROLE_TAG: Record<MemberRole, string> = { super_admin: '만든 사람', admin: '관리자', member: '' }

type Props = {
  members: Member[] // role 이 채워진 멤버. 역할 높은 순
  meId: string
  myRole: MemberRole
  error?: string
  busy?: boolean
  onNick?: (nick: string) => Promise<boolean> | void // 성공하면 true → 고치기 칸 닫음
  onRole?: (id: string, role: 'admin' | 'member') => void
  onKick?: (id: string) => void
  onProfile?: (id: string) => void // 아바타 누르기 → 프로필
}

// 멤버: 찾기 · 목록. 내 줄에서 아지트 닉네임 바꾸기, 만든 사람은 관리자 올리기·내리기, 관리자 이상은 아래 역할 내보내기.
export function MembersTab({ members, meId, myRole, error, busy, onNick, onRole, onKick, onProfile }: Props) {
  const [query, setQuery] = useState('')
  const [editing, setEditing] = useState(false)
  const [nick, setNick] = useState('')
  const [kicking, setKicking] = useState<Member>() // 내보낼까요? 물어보는 중인 사람

  const q = query.trim().toLowerCase()
  const shown = q ? members.filter((m) => m.nick.toLowerCase().includes(q) || (m.name ?? '').toLowerCase().includes(q)) : members

  const saveNick = async () => {
    if (!nick.trim() || busy || !onNick) return
    if ((await onNick(nick.trim())) !== false) setEditing(false)
  }

  return (
    <>
      <div className="row">
        <h1 className="page-title sm grow">멤버 {members.length}명</h1>
        <label htmlFor="sm-q" className="sr">멤버 찾기</label>
        <input id="sm-q" className="inp member-search" placeholder="닉네임이나 이름" value={query} onChange={(e) => setQuery(e.target.value)} />
      </div>
      {error && <span role="alert" className="ferr">{error}</span>}

      <div className="col">
        {shown.map((m) => {
          const self = m.id === meId
          const role = (m.role ?? 'member') as MemberRole
          const canKick = !self && RANK[myRole] >= RANK.admin && RANK[myRole] > RANK[role]
          const canRole = !self && myRole === 'super_admin' && role !== 'super_admin'
          return (
            <div key={m.id} className="lrow">
              <button type="button" className="plain" style={{ flexShrink: 0 }} aria-label={`${m.nick} 프로필`} onClick={onProfile && (() => onProfile(m.id))}>
                <Avatar name={m.nick} color={m.color} size={36} src={m.src} />
              </button>
              <div className="grow col" style={{ gap: 2, minWidth: 0 }}>
                {self && editing
                  ? (
                    <form className="row" style={{ gap: 6 }} onSubmit={(e) => { e.preventDefault(); void saveNick() }}>
                      <label htmlFor="sm-nick" className="sr">이 아지트 닉네임</label>
                      <input id="sm-nick" className="inp member-nick" maxLength={16} value={nick} autoFocus onChange={(e) => setNick(e.target.value)} />
                      <button type="submit" className="btn s p" disabled={!nick.trim() || busy}>저장</button>
                      <button type="button" className="btn s sec" onClick={() => setEditing(false)}>취소</button>
                    </form>
                  )
                  : (
                    <span className="b ell">
                      {m.nick}
                      {self && <> <span className="tag lime">나</span></>}
                      {ROLE_TAG[role] && <> <span className="tag">{ROLE_TAG[role]}</span></>}
                    </span>
                  )}
                <span className="member-sub">{m.name && <span className="mono">@{m.name}</span>}{m.joinedAgo && <> · {m.joinedAgo} 들어옴</>}</span>
              </div>
              {self && !editing && <button type="button" className="btn s sec" onClick={() => { setNick(m.nick); setEditing(true) }}>닉네임 바꾸기</button>}
              {canRole && (
                <button type="button" className="btn s sec" disabled={busy} onClick={onRole && (() => onRole(m.id, role === 'admin' ? 'member' : 'admin'))}>
                  {role === 'admin' ? '멤버로 내리기' : '관리자로 올리기'}
                </button>
              )}
              {canKick && <button type="button" className="btn s dg" disabled={busy} onClick={() => setKicking(m)}>내보내기</button>}
            </div>
          )
        })}
        {shown.length === 0 && <span className="hint">찾는 멤버가 없어요</span>}
      </div>
      <span className="hint">이 아지트 닉네임은 여기서만 보여요. 내보낸 사람도 초대 코드가 있으면 다시 들어올 수 있어요.</span>

      {kicking && (
        <Modal onClose={() => setKicking(undefined)}>
          <h2 className="disp modal-title">{kicking.nick}님을 내보낼까요?</h2>
          <p className="modal-text">보낸 메시지는 남아요. 새 초대 코드가 있으면 다시 들어올 수 있어요.</p>
          <div className="modal-actions">
            <button type="button" className="btn sec" onClick={() => setKicking(undefined)}>취소</button>
            <button type="button" className="btn dgs" disabled={busy} onClick={() => { onKick?.(kicking.id); setKicking(undefined) }}>내보내기</button>
          </div>
        </Modal>
      )}
    </>
  )
}
