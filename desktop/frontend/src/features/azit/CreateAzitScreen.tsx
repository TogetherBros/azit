import { useState } from 'react'
import { AuthLayout } from '../../app/AuthLayout'
import { Swatches } from '../../components/Swatches'

const PRESETS = ['짤방', '게임방', '공지', '약속 잡기']

export type NewAzit = { name: string; color: string; presets: string[] }

type Props = {
  name?: string
  color?: string
  presets?: string[] // 고른 기본 방
  error?: string
  busy?: boolean
  onSubmit?: (a: NewAzit) => void
  onCancel?: () => void
}

// 새 아지트 만들기. "아지트" 방은 항상 생기고, 나머지 방은 골라서 같이 만든다.
export function CreateAzitScreen({ name: name0 = '', color: color0 = '#FF5FA8', presets: presets0 = ['짤방', '게임방'], error, busy, onSubmit, onCancel }: Props) {
  const [name, setName] = useState(name0)
  const [color, setColor] = useState(color0)
  const [presets, setPresets] = useState(presets0)

  const toggle = (p: string) => setPresets((ps) => (ps.includes(p) ? ps.filter((x) => x !== p) : [...ps, p]))

  return (
    <AuthLayout blobs="none">
      <form
        className="card create-card"
        noValidate
        onSubmit={(e) => {
          e.preventDefault()
          if (name.trim() && !busy) onSubmit?.({ name: name.trim(), color, presets })
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

        <div className="field">
          <span className="lbl">처음부터 있을 방 <span className="lbl-sub">· 아지트 방은 늘 생겨요</span></span>
          <div className="row wrap">
            {PRESETS.map((p) => {
              const on = presets.includes(p)
              return (
                <button key={p} type="button" className={`chip${on ? ' on' : ''}`} aria-pressed={on} onClick={() => toggle(p)}>
                  {on ? p : `+ ${p}`}
                </button>
              )
            })}
          </div>
          <span className="hint">방은 한번 만들면 지울 수 없어요. 초대 코드는 만든 뒤 아지트 설정에서 만들 수 있어요.</span>
        </div>

        <button type="submit" className="btn l p" disabled={!name.trim() || busy}>{busy ? '만드는 중…' : '만들기'}</button>
      </form>
    </AuthLayout>
  )
}
