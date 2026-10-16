import { useState } from 'react'
import { AuthLayout } from '../../app/AuthLayout'
import { Swatches } from '../../components/Swatches'

export type NewAzit = { name: string; color: string }

type Props = {
  name?: string
  color?: string
  error?: string
  busy?: boolean
  onSubmit?: (a: NewAzit) => void
  onCancel?: () => void
}

// 새 아지트 만들기. 이름과 색만 정하고, 방은 기본 방("아지트") 하나만 생긴다.
export function CreateAzitScreen({ name: name0 = '', color: color0 = '#FF5FA8', error, busy, onSubmit, onCancel }: Props) {
  const [name, setName] = useState(name0)
  const [color, setColor] = useState(color0)

  return (
    <AuthLayout blobs="none">
      <form
        className="card create-card"
        noValidate
        onSubmit={(e) => {
          e.preventDefault()
          if (name.trim() && !busy) onSubmit?.({ name: name.trim(), color })
        }}
      >
        <div className="row">
          <h1 className="create-title">새 아지트</h1>
          <button type="button" className="btn s sec" onClick={onCancel}>취소</button>
        </div>

        <div className="field">
          <label htmlFor="ca-name">아지트 이름</label>
          <input
            id="ca-name"
            className={`inp${error ? ' err' : ''}`}
            maxLength={30}
            placeholder="예: 동네 친구들"
            value={name}
            onChange={(e) => setName(e.target.value)}
            autoFocus
          />
          {error && <span role="alert" className="ferr">{error}</span>}
        </div>

        <div className="field">
          <span className="lbl">아이콘 색</span>
          <Swatches selected={color} square label="아이콘 색" onPick={setColor} />
        </div>

        <button type="submit" className="btn l p" disabled={!name.trim() || busy}>{busy ? '만드는 중…' : '만들기'}</button>
      </form>
    </AuthLayout>
  )
}
