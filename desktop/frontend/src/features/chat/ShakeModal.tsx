import type { Member } from '../../api/types'
import { Avatar } from '../../components/Avatar'
import { Modal } from '../../components/Modal'

type Props = {
  members: Member[] // 나를 뺀 멤버
  onPick?: (id: string) => void // 고르면 그 사람을 흔든다
  onClose?: () => void
}

// + 메뉴 → 흔들기. 글을 쓰지 않고 바로 한 사람을 흔든다 ("@닉네임" 만 있는 메시지로 보낸다).
export function ShakeModal({ members, onPick, onClose }: Props) {
  return (
    <Modal title="누구를 흔들까요?" onClose={onClose}>
      <span className="hint">
        고르면 그 사람 창이 흔들리고 알림이 가요. 메시지에 @로 부르는 것도 흔들기예요.
      </span>
      <div className="col scroll" style={{ gap: 2, maxHeight: 360 }}>
        {members.length === 0 && <span className="hint">아직 흔들 사람이 없어요</span>}
        {members.map((m) => (
          <button key={m.id} type="button" className="pitem" onClick={onPick && (() => onPick(m.id))}>
            <Avatar name={m.nick} color={m.color} size={30} src={m.src} />
            <span className="pitem-name">{m.nick}</span>
            <span className="pitem-sub">{m.via ?? (m.online ? '접속 중' : '')}</span>
          </button>
        ))}
      </div>
    </Modal>
  )
}
