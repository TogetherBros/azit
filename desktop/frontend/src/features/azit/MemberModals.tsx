import type { Member, Stats } from '../../api/types'
import { Avatar } from '../../components/Avatar'
import { Stat } from '../../components/Controls'
import { Icon } from '../../components/Icon'
import { CloseButton, Modal } from '../../components/Modal'

type MemberProps = {
  member: Member
  stats: Stats
  titleCount: number
  self?: boolean
  roomName?: string // 지금 보고 있는 방 ("다같이에서 흔들기")
}

// 아바타·이름을 누르면 뜨는 멤버 프로필
export function MemberModal({ member, stats, titleCount, self, roomName }: MemberProps) {
  const status = member.dnd ? '방해 금지 중 · ' : member.online ? '접속 중 · ' : ''
  return (
    <Modal>
      <div className="row" style={{ alignItems: 'flex-start' }}>
        <div className="row grow" style={{ gap: 14 }}>
          <Avatar name={member.nick} color={member.color} size={64} />
          <div className="col" style={{ gap: 3 }}>
            <span className="disp" style={{ fontSize: 24 }}>{member.nick}</span>
            <span className="profile-id">{member.id}</span>
            {member.title && <span className="tag title" style={{ background: member.color }}>{member.title}</span>}
          </div>
        </div>
        <CloseButton />
      </div>
      <span className="hint">{status}{member.via} · {member.joinedAgo}</span>
      <div className="grid3" style={{ gap: 8 }}>
        <Stat value={stats.games} label="라이어 판" />
        <Stat value={stats.catches} label="라이어 검거" />
        <Stat value={stats.liarSurvive} label="라이어로 생존" />
        <Stat value={stats.citizenWins} label="시민 승리" />
        <Stat value={stats.shakes} label="흔들기" />
        <Stat value={titleCount} label="칭호" />
      </div>
      {self
        ? <button type="button" className="btn p">내 정보 고치기</button>
        : (
          <div className="row wrap">
            <button type="button" className="btn g"><Icon name="shake" size={16} />{roomName ? `${roomName}에서 흔들기` : '흔들기'}</button>
            <button type="button" className="btn sec">@ 불러서 말하기</button>
            <button type="button" className="btn dg">내보내기</button>
          </div>
        )}
    </Modal>
  )
}

type KickProps = { nick: string; ban?: boolean }

// 내보내기 확인. 모두 주인이라 누구나 할 수 있고, 누구나 되돌릴 수 있다.
export function KickModal({ nick, ban }: KickProps) {
  return (
    <Modal>
      <h2 className="disp modal-title">{nick} 님을 내보낼까요?</h2>
      <p className="modal-text">
        변경 기록에 내 이름으로 남고, 누구든 되돌릴 수 있어요.{' '}
        {ban ? '다시 못 들어오게 막으면 새 코드로도 못 들어와요.' : '새 코드를 받으면 다시 들어올 수 있어요.'}
      </p>
      <label className="check">
        <input type="checkbox" defaultChecked={ban} />
        다시 못 들어오게 막기
      </label>
      <div className="modal-actions">
        <button type="button" className="btn sec">취소</button>
        <button type="button" className="btn dgs">내보내기</button>
      </div>
    </Modal>
  )
}
