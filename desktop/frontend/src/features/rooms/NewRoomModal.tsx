import { useState } from 'react'
import { Modal } from '../../components/Modal'

type Props = {
  memberCount: number
  name?: string
  error?: string
  busy?: boolean
  onSubmit?: (name: string) => void
  onClose?: () => void
}

// 방 만들기. 아지트 멤버는 모든 방에 자동으로 들어간다.
export function NewRoomModal({ memberCount, name: name0 = '', error, busy, onSubmit, onClose }: Props) {
  const [name, setName] = useState(name0)
  return (
    <Modal title="새 방" onClose={onClose}>
      <form
        className="col"
        style={{ gap: 16 }}
        noValidate
        onSubmit={(e) => {
          e.preventDefault()
          if (name.trim() && !busy) onSubmit?.(name.trim())
        }}
      >
        <div className="field">
          <label htmlFor="nr-name">방 이름</label>
          <input
            id="nr-name"
            className={`inp${error ? ' err' : ''}`}
            maxLength={20}
            placeholder="예: 시험 끝나고"
            value={name}
            onChange={(e) => setName(e.target.value)}
            autoFocus
          />
          {error
            ? <span role="alert" className="ferr">{error}</span>
            : <span className="hint">아지트 멤버 {memberCount}명이 모두 들어가요. 지울 수는 없어요</span>}
        </div>
        <button type="submit" className="btn p" disabled={!name.trim() || busy}>{busy ? '만드는 중…' : '만들기'}</button>
      </form>
    </Modal>
  )
}
