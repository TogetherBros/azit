import type { Azit } from '../../../api/types'
import { Toggle } from '../../../components/Controls'
import { Mark } from '../../../components/Mark'
import { Swatches } from '../../../components/Swatches'

type Props = {
  azit: Azit
  rooms: { id: string; name: string }[]
  landingId: string
  welcome: boolean
  lastChange: string // "강백호가 바꿈 · 3일 전"
  error?: string // "이름은 10분에 한 번 바꿀 수 있어요 · 7분 뒤에"
  dirty?: boolean
}

// 기본 정보: 이름 · 짧은 이름 · 색 · 소개 · 처음 방 · 입장 환영 | 미리보기
export function GeneralTab({ azit, rooms, landingId, welcome, lastChange, error, dirty }: Props) {
  return (
    <>
      <h1 className="page-title sm">기본 정보</h1>
      <div className="sg">
        <div className="sg-form">
          <div className="field">
            <label htmlFor="sg-name">아지트 이름 <span className="sg-count">{azit.name.length}/30</span></label>
            <input id="sg-name" className={`inp${error ? ' err' : ''}`} defaultValue={azit.name} />
            {error
              ? <span className="ferr" role="alert">{error}</span>
              : <span className="hint">마지막으로 {lastChange} · 10분에 한 번만 바꿀 수 있어요</span>}
          </div>
          <div className="grid2">
            <div className="field">
              <label htmlFor="sg-short">짧은 이름</label>
              <input id="sg-short" className="inp" maxLength={3} defaultValue={azit.short} />
              <span className="hint">왼쪽 네모에 들어가요 · 3자까지</span>
            </div>
            <div className="field">
              <span className="lbl">아이콘 색</span>
              <Swatches selected={azit.color} square label="아이콘 색" />
            </div>
          </div>
          <div className="field">
            <label htmlFor="sg-desc">한 줄 소개</label>
            <input id="sg-desc" className="inp" maxLength={40} defaultValue={azit.desc} />
            <span className="hint">초대 코드 미리보기에 보여요</span>
          </div>
          <div className="field">
            <label htmlFor="sg-landing">새 멤버가 처음 들어올 방</label>
            <select id="sg-landing" className="inp" defaultValue={landingId}>
              {rooms.map((r) => <option key={r.id} value={r.id}>{r.name}</option>)}
            </select>
          </div>
          <div className="lrow">
            <div className="grow col" style={{ gap: 2 }}>
              <span className="b">입장 환영</span>
              <span className="hint">누가 들어오면 알림 + 환영 흔들기 버튼</span>
            </div>
            <Toggle on={welcome} label="입장 환영" />
          </div>
          <div className="row">
            <button type="button" className="btn p" disabled={!dirty}>저장</button>
            <button type="button" className="btn sec" disabled={!dirty}>되돌리기</button>
          </div>
        </div>

        <div className="sg-preview">
          <span className="mu b" style={{ fontSize: 12 }}>미리보기</span>
          <div className="row" style={{ gap: 12 }}>
            <span className="mark rail-size" style={{ background: azit.color }}>{azit.short || '?'}</span>
            <div className="card2 grow sg-name-chip" style={{ background: azit.color }}>{azit.name || '이름 없음'}</div>
          </div>
          <div className="card col" style={{ gap: 4 }}>
            <span className="mu" style={{ fontSize: 11 }}>초대 미리보기</span>
            <div className="row">
              <Mark label={azit.short} color={azit.color} size="sm" />
              <div className="col" style={{ gap: 2 }}>
                <span className="disp" style={{ fontSize: 15, lineHeight: 1.3 }}>{azit.name}</span>
                <span className="mu" style={{ fontSize: 12 }}>{azit.desc || '소개 없음'} · {azit.memberCount}명</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}
