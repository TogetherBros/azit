import { useRef, useState } from 'react'
import type { ReactOption } from '../../../api/types'
import { Icon } from '../../../components/Icon'
import { Modal } from '../../../components/Modal'

export const MAX_REACTS = 10
export const MAX_REACT_CHARS = 4

type Props = {
  reacts: ReactOption[] // 아지트 맞장구 목록 (순서대로, 최대 10개)
  draft: string // 만드는 중인 글자 (1~4자)
  error?: string // 넣기 실패 이유 (꽉 참, 이미 있음)
  readOnly?: boolean // 관리자가 아니면 보기만
  busy?: boolean
  onAdd?: (text: string) => Promise<boolean> | void // 성공하면 true → 입력칸 비움
  onMove?: (index: number, dir: -1 | 1) => void
  onRemove?: (index: number) => void
}

// 맞장구 공방: 1~4글자 맞장구를 만들어 보고, 아지트 목록(10칸)에 넣고 순서를 바꾼다.
export function ReactsTab({ reacts, draft: draft0, error: error0, readOnly, busy, onAdd, onMove, onRemove }: Props) {
  const [draft, setDraft] = useState(draft0)
  const [removing, setRemoving] = useState<number>() // 지울까요? 물어보는 중인 칸 번호
  const inputRef = useRef<HTMLInputElement>(null)
  const chars = Array.from(draft.trim()).slice(0, MAX_REACT_CHARS)
  const text = chars.join('')
  const full = reacts.length >= MAX_REACTS
  const dup = text !== '' && reacts.some((r) => r.text === text)
  const canAdd = !readOnly && chars.length > 0 && !full && !dup && !busy
  const error = error0 ?? (dup ? `${text}는 이미 목록에 있어요` : full && chars.length ? '10개 꽉 찼어요. 하나 빼고 넣어 주세요' : undefined)

  const add = async () => {
    if (!canAdd || !onAdd) return
    const ok = await onAdd(text)
    if (ok) setDraft('')
  }
  const slots = Array.from({ length: MAX_REACTS }, (_, i) => reacts[i])

  return (
    <>
      <h1 className="page-title sm">맞장구</h1>

      {readOnly && <div className="card2 rx-lock">관리자만 바꿀 수 있어요. 보기만 할 수 있어요.</div>}

      {!readOnly && (
        <section className="card rx-lab">
          <div className="rx-stage">
            <div className={`rx-stamp n${chars.length}${chars.length ? '' : ' empty'}`} aria-hidden>
              {chars.length ? chars.join('') : '?'}
            </div>
            <span className="rx-stage-cap">미리보기</span>
          </div>

          <div className="rx-form col">
            <span className="b">글자 맞장구 만들기</span>
            <form className="row wrap rx-row" noValidate onSubmit={(e) => { e.preventDefault(); void add() }}>
              <label htmlFor="rx-draft" className="sr">맞장구 글자 (1~4자)</label>
              <div className="rx-cells">
                {Array.from({ length: MAX_REACT_CHARS }, (_, i) => i).map((i) => (
                  <span key={i} aria-hidden className={`rx-cell${chars[i] ? ' on' : ''}${i === chars.length ? ' cur' : ''}`}>{chars[i] ?? ''}</span>
                ))}
                {/* 칸 위에 투명하게 덮은 진짜 입력칸 (한글 조합이 칸마다 끊기지 않게 한 칸으로 받는다) */}
                <input
                  id="rx-draft"
                  ref={inputRef}
                  className="rx-input"
                  value={draft}
                  autoComplete="off"
                  spellCheck={false}
                  onChange={(e) => {
                    const v = e.target.value
                    setDraft(Array.from(v).length > MAX_REACT_CHARS ? Array.from(v).slice(0, MAX_REACT_CHARS).join('') : v)
                  }}
                />
              </div>
              <span className="rx-count mono">{chars.length}/{MAX_REACT_CHARS}</span>
              <button type="submit" className="btn p rx-add" disabled={!canAdd}><Icon name="plus" size={14} /> 목록에 넣기</button>
            </form>
            {error ? <span className="rx-err">{error}</span> : <span className="hint">한글 · 영어 · 숫자 · 이모지 1~4자</span>}

          </div>
        </section>
      )}

      <section className="card col" style={{ gap: 12 }}>
        <span className="b">아지트 맞장구 <span className={full ? 'rx-full' : 'mu'}>{reacts.length}/{MAX_REACTS}</span></span>
        <ol className="rx-grid">
          {slots.map((r, i) => r
            ? (
              <li key={r.id} className={`rx-tile${r.img ? ' meme' : ''}${r.text === text && dup ? ' hit' : ''}`}>
                <span className="rx-pos mono">{i + 1}</span>
                {r.img ? <img src={r.img} alt={r.name} /> : <span className={`rx-tile-text n${Array.from(r.text ?? '').length}`}>{r.text}</span>}
                {r.img && <span className="rx-tile-name">{r.name}</span>}
                {!readOnly && (
                  <button type="button" className="rx-x" aria-label={`맞장구 빼기 ${r.text ?? r.name}`} disabled={busy} onClick={() => setRemoving(i)}><Icon name="x" size={11} /></button>
                )}
                {!readOnly && (
                  <div className="rx-tools">
                    <button type="button" className="btn s icon sec" aria-label="앞으로" disabled={i === 0 || busy} onClick={onMove && (() => onMove(i, -1))}><Icon name="back" size={12} /></button>
                    <button type="button" className="btn s icon sec" aria-label="뒤로" disabled={i === reacts.length - 1 || busy} onClick={onMove && (() => onMove(i, 1))}><Icon name="chev" size={12} /></button>
                  </div>
                )}
              </li>
            )
            : (
              <li key={`empty-${i}`} className="rx-tile empty">
                <span className="rx-pos mono">{i + 1}</span>
                {i === reacts.length && !readOnly && chars.length > 0 && !dup
                  ? <span className={`rx-tile-text ghost n${chars.length}`}>{chars.join('')}</span>
                  : <span className="rx-empty-dot" />}
              </li>
            ))}
        </ol>
      </section>

      {removing !== undefined && reacts[removing] && (
        <Modal onClose={() => setRemoving(undefined)}>
          <h2 className="disp modal-title">"{reacts[removing].text ?? reacts[removing].name}" 맞장구를 지울까요?</h2>
          <p className="modal-text">목록에서 빠져요. 이미 메시지에 달린 맞장구는 그대로 남아요.</p>
          <div className="modal-actions">
            <button type="button" className="btn sec" onClick={() => setRemoving(undefined)}>취소</button>
            <button type="button" className="btn dgs" disabled={busy} onClick={() => { onRemove?.(removing); setRemoving(undefined) }}>지우기</button>
          </div>
        </Modal>
      )}
    </>
  )
}
