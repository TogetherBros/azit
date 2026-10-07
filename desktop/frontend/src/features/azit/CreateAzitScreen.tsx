import { AuthLayout } from '../../app/AuthLayout'
import { Swatches } from '../../components/Swatches'

const PRESETS = ['짤방', '게임방', '공지', '약속 잡기']

type Props = {
  name?: string
  color?: string
  presets?: string[] // 고른 기본 방
  error?: string
}

// 새 아지트 만들기. "다같이" 방은 항상 생기고, 나머지 방은 골라서 같이 만든다.
export function CreateAzitScreen({ name, color = '#FF5FA8', presets = ['짤방', '게임방'], error }: Props) {
  return (
    <AuthLayout blobs="none">
      <form className="card create-card" noValidate>
        <div className="row">
          <h1 className="create-title">새 아지트</h1>
          <button type="button" className="btn s sec">취소</button>
        </div>

        <div className="field">
          <label htmlFor="ca-name">아지트 이름</label>
          <input id="ca-name" className={`inp${error ? ' err' : ''}`} maxLength={30} placeholder="예: 동네 친구들" defaultValue={name} />
          {error && <span role="alert" className="ferr">{error}</span>}
        </div>

        <div className="field">
          <span className="lbl">아이콘 색</span>
          <Swatches selected={color} square label="아이콘 색" />
        </div>

        <div className="field">
          <span className="lbl">처음부터 있을 방 <span className="lbl-sub">· 다같이 방은 늘 생겨요</span></span>
          <div className="row wrap">
            {PRESETS.map((p) => {
              const on = presets.includes(p)
              return (
                <button key={p} type="button" className={`chip${on ? ' on' : ''}`} aria-pressed={on}>
                  {on ? p : `+ ${p}`}
                </button>
              )
            })}
          </div>
          <span className="hint">들어온 사람은 전부 주인이에요. 방과 아지트는 한번 만들면 아무도 지울 수 없어요.</span>
        </div>

        <button type="submit" className="btn l p">만들기</button>
      </form>
    </AuthLayout>
  )
}
