import { Modal } from '../../components/Modal'

type Props = { memberCount: number; name?: string; error?: string }

// 방 만들기. 아지트 멤버는 모든 방에 자동으로 들어간다.
export function NewRoomModal({ memberCount, name, error }: Props) {
  return (
    <Modal title="새 방">
      <form className="col" style={{ gap: 16 }} noValidate>
        <div className="field">
          <label htmlFor="nr-name">방 이름</label>
          <input id="nr-name" className={`inp${error ? ' err' : ''}`} maxLength={20} placeholder="예: 시험 끝나고" defaultValue={name} />
          {error
            ? <span role="alert" className="ferr">{error}</span>
            : <span className="hint">아지트 멤버 {memberCount}명이 모두 들어가요. 지울 수는 없고 보관만 돼요</span>}
        </div>
        <button type="submit" className="btn p">만들기</button>
      </form>
    </Modal>
  )
}
