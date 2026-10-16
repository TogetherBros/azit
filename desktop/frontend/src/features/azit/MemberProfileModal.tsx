import type { MemberRes } from '../../api/server'
import { Modal } from '../../components/Modal'
import { ProfileCard } from '../../components/ProfileCard'

const ROLE_LABEL = { super_admin: '만든 사람', admin: '관리자', member: '멤버' } as const

type Props = {
  member: MemberRes
  self?: boolean
  onClose: () => void
}

// 채팅·멤버 탭에서 아바타나 이름을 누르면 뜨는 프로필. 출입증 카드에 이 아지트 닉네임·역할·들어온 날을 얹는다.
export function MemberProfileModal({ member: m, self, onClose }: Props) {
  return (
    <Modal onClose={onClose}>
      <div className="profile-modal">
        <ProfileCard
          p={{
            nickname: m.nickname,
            name: m.name,
            color: m.color,
            bio: m.bio || undefined,
            avatarSrc: m.avatar_url || undefined,
            bannerSrc: m.banner_url || undefined,
            since: dotDate(m.joined_at),
          }}
        />
        <div className="row wrap" style={{ gap: 6, justifyContent: 'center' }}>
          <span className={`tag${m.role === 'member' ? '' : ' lime'}`}>{ROLE_LABEL[m.role]}</span>
          {self && <span className="tag">나</span>}
          <span className="hint">이 아지트 닉네임으로 보여요 · {dotDate(m.joined_at)} 들어옴</span>
        </div>
        <button type="button" className="btn sec" onClick={onClose}>닫기</button>
      </div>
    </Modal>
  )
}

// "2026.10.14"
function dotDate(iso: string): string {
  const d = new Date(iso)
  return `${d.getFullYear()}.${String(d.getMonth() + 1).padStart(2, '0')}.${String(d.getDate()).padStart(2, '0')}`
}
