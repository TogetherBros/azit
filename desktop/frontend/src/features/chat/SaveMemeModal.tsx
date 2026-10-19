import { useState } from 'react'
import { Modal } from '../../components/Modal'

type Props = {
  busy?: boolean
  error?: string
  onCancel?: () => void
  onSave?: (name: string) => void
}

// 채팅 사진 → 아지트 짤. 이름만 물어본다 (그림은 서버가 원래 사진을 복사한다).
export function SaveMemeModal({ busy, error, onCancel, onSave }: Props) {
  const [name, setName] = useState('')
  const ok = name.trim().length > 0 && !busy

  return (
    <Modal onClose={onCancel}>
      <h2 className="disp modal-title">짤로 저장</h2>
      <p className="modal-text">이 사진이 아지트 짤에 들어가요. 누구나 보내고 맞장구로 쓸 수 있어요.</p>
      <form className="col" style={{ gap: 6 }} onSubmit={(e) => { e.preventDefault(); if (ok) onSave?.(name.trim()) }}>
        <label htmlFor="save-meme-name" className="lbl">짤 이름</label>
        <input id="save-meme-name" className="inp" maxLength={20} value={name} autoFocus onChange={(e) => setName(e.target.value)} />
        {error && <span role="alert" className="ferr">{error}</span>}
        <div className="modal-actions">
          <button type="button" className="btn sec" onClick={onCancel}>취소</button>
          <button type="submit" className="btn p" disabled={!ok}>{busy ? '저장하는 중…' : '저장'}</button>
        </div>
      </form>
    </Modal>
  )
}
