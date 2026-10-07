import type { Member } from '../../api/types'
import { Avatar } from '../../components/Avatar'
import { Modal } from '../../components/Modal'

type Props = { members: Member[] } // 나를 뺀 멤버, 접속 중인 사람 먼저

// + 메뉴 → 흔들기. 메시지 없이 상대 창을 흔든다.
export function ShakeModal({ members }: Props) {
  return (
    <Modal title="누구를 흔들까요?">
      <span className="hint">
        메시지 없이 흔들면 그 사람 창이 흔들리고 알림이 가요. 메시지에 @로 부르는 것도 흔들기예요. 1분에 3번까지.
      </span>
      <div className="col scroll" style={{ gap: 2, maxHeight: 360 }}>
        {members.map((m) => (
          <button key={m.id} type="button" className="pitem">
            <Avatar name={m.nick} color={m.color} size={30} />
            <span className="pitem-name">{m.nick}</span>
            <span className="pitem-sub">{m.dnd ? '방해 금지라 알림만' : m.online ? '접속 중' : ''}</span>
          </button>
        ))}
      </div>
    </Modal>
  )
}
