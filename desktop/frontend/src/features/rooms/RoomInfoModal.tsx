import type { Member, Room } from '../../api/types'
import { Avatar } from '../../components/Avatar'
import { Mark } from '../../components/Mark'
import { CloseButton, Modal } from '../../components/Modal'
import { roomColor } from '../../lib/color'

type Props = {
  room: Room
  madeBy: string // "강백호가 만듦"
  messageCount: number
  members: Member[] // 접속 중인 사람 먼저
  meId: string
}

// 채팅 머리의 "N명" 을 누르면 뜨는 방 정보
export function RoomInfoModal({ room, madeBy, messageCount, members, meId }: Props) {
  return (
    <Modal>
      <div className="modal-head">
        <Mark label={room.name.slice(0, 1)} color={roomColor(room.name)} size="md" />
        <div className="grow col" style={{ gap: 2 }}>
          <span className="disp" style={{ fontSize: 22 }}>{room.name}</span>
          <span className="hint">{madeBy} · 메시지 {messageCount}개{room.landing ? ' · 처음 방' : ''}</span>
        </div>
        <CloseButton />
      </div>

      <span className="lbl">
        멤버 {members.length}명 <span className="lbl-sub">· 아지트 멤버는 모든 방에 있어요</span>
      </span>
      <div className="col scroll" style={{ gap: 2, maxHeight: 340 }}>
        {members.map((m) => (
          <button key={m.id} type="button" className="pitem">
            <Avatar name={m.nick} color={m.color} size={30} />
            <span className="pitem-name">{m.nick}{m.id === meId ? ' (나)' : ''}</span>
            <span className="pitem-sub">{m.dnd ? '방해 금지' : m.online ? '접속 중' : ''}</span>
          </button>
        ))}
      </div>
    </Modal>
  )
}
