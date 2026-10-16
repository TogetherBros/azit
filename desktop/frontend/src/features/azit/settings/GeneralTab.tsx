import { useRef } from 'react'
import { Mark } from '../../../components/Mark'
import { Swatches } from '../../../components/Swatches'

export type GeneralValues = {
  name: string
  color: string
  desc: string
  banner?: string // 배너 보기 주소 (새로 고른 사진이면 로컬 미리보기)
}

type Props = {
  value: GeneralValues
  memberCount: number
  readOnly?: boolean // 관리자가 아니면 보기만
  dirty?: boolean // 바꾼 게 있음
  busy?: boolean
  error?: string
  saved?: boolean // 방금 저장함
  onChange?: (v: GeneralValues) => void
  onBanner?: (file: File) => void // 배너 사진 고르기
  onSave?: () => void
  onReset?: () => void
}

const NAME_MAX = 20
const DESC_MAX = 40

// 기본 정보: 이름 · 색 · 한 줄 소개 · 배너 | 미리보기
export function GeneralTab({ value, memberCount, readOnly, dirty, busy, error, saved, onChange, onBanner, onSave, onReset }: Props) {
  const file = useRef<HTMLInputElement>(null)
  const set = (patch: Partial<GeneralValues>) => onChange?.({ ...value, ...patch })
  const short = Array.from(value.name.trim()).slice(0, 2).join('')

  return (
    <>
      <h1 className="page-title sm">기본 정보</h1>
      <div className="sg">
        <div className="sg-form">
          <div className="field">
            <label htmlFor="sg-name">아지트 이름 <span className="sg-count">{Array.from(value.name).length}/{NAME_MAX}</span></label>
            <input
              id="sg-name"
              className={`inp${error ? ' err' : ''}`}
              value={value.name}
              maxLength={NAME_MAX}
              readOnly={readOnly}
              onChange={(e) => set({ name: e.target.value })}
            />
          </div>
          <div className="field">
            <span className="lbl">아이콘 색</span>
            <Swatches selected={value.color} square label="아이콘 색" onPick={readOnly ? undefined : (color) => set({ color })} />
          </div>
          <div className="field">
            <label htmlFor="sg-desc">한 줄 소개 <span className="sg-count">{Array.from(value.desc).length}/{DESC_MAX}</span></label>
            <input
              id="sg-desc"
              className="inp"
              value={value.desc}
              maxLength={DESC_MAX}
              placeholder="어떤 아지트인지 한 줄로"
              readOnly={readOnly}
              onChange={(e) => set({ desc: e.target.value })}
            />
            <span className="hint">초대 코드 미리보기에 보여요</span>
          </div>
          <div className="field">
            <span className="lbl">배너</span>
            <div className="sg-banner" style={value.banner ? { backgroundImage: `url("${value.banner}")` } : { background: value.color }}>
              {!value.banner && <span className="sg-banner-empty">배너 없음 · 아이콘 색으로 보여요</span>}
            </div>
            {!readOnly && (
              <div className="row" style={{ gap: 6 }}>
                <button type="button" className="btn s sec" onClick={() => file.current?.click()}>사진 고르기</button>
                {value.banner && <button type="button" className="btn s sec" onClick={() => set({ banner: undefined })}>배너 빼기</button>}
                <input
                  ref={file}
                  type="file"
                  accept="image/png,image/jpeg,image/webp"
                  hidden
                  onChange={(e) => {
                    const f = e.target.files?.[0]
                    e.target.value = ''
                    if (f) onBanner?.(f)
                  }}
                />
              </div>
            )}
            <span className="hint">가운데를 12:5 로 잘라서 올려요</span>
          </div>
          {error && <span className="ferr" role="alert">{error}</span>}
          {readOnly
            ? <span className="hint">관리자만 바꿀 수 있어요</span>
            : (
              <div className="row">
                <button type="button" className="btn p" disabled={!dirty || busy} onClick={onSave}>{busy ? '저장 중…' : '저장'}</button>
                <button type="button" className="btn sec" disabled={!dirty || busy} onClick={onReset}>되돌리기</button>
                {saved && !dirty && <span className="hint">저장했어요</span>}
              </div>
            )}
        </div>

        <div className="sg-preview">
          <span className="mu b" style={{ fontSize: 12 }}>미리보기</span>
          <div className="row" style={{ gap: 12 }}>
            <span className="mark rail-size" style={{ background: value.color }}>{short || '?'}</span>
            <div className="card2 grow sg-name-chip" style={{ background: value.color }}>{value.name || '이름 없음'}</div>
          </div>
          <div className="card col" style={{ gap: 4 }}>
            <span className="mu" style={{ fontSize: 11 }}>초대 미리보기</span>
            <div className="row">
              <Mark label={short} color={value.color} size="sm" />
              <div className="col" style={{ gap: 2 }}>
                <span className="disp" style={{ fontSize: 15, lineHeight: 1.3 }}>{value.name}</span>
                <span className="mu" style={{ fontSize: 12 }}>{value.desc || '소개 없음'} · {memberCount}명</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}
