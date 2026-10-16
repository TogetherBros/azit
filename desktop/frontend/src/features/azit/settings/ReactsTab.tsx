import { useRef, useState } from 'react'
import type { ReactOption, Sticker } from '../../../api/types'
import { Icon } from '../../../components/Icon'

export const MAX_REACTS = 10
export const MAX_REACT_CHARS = 4
const SUGGEST = ['ㅋㅋ', 'ㄹㅇ', '헐', 'ㄱㄱ', 'ㅠㅠ', 'ㅇㅈ', '굿', '실화냐', '킹받네', '가보자고']

type Props = {
  reacts: ReactOption[] // 아지트 맞장구 목록 (순서대로, 최대 10개)
  draft: string // 만드는 중인 글자 (1~4자)
  error?: string // 넣기 실패 이유 (꽉 참, 이미 있음)
  readOnly?: boolean // 관리자가 아니면 보기만
  stickers?: Sticker[] // "짤에서 고르기" 펼쳤을 때
  busy?: boolean
  onAdd?: (text: string) => Promise<boolean> | void // 성공하면 true → 입력칸 비움
  onMove?: (index: number, dir: -1 | 1) => void
  onRemove?: (index: number) => void
  onToggleMemes?: () => void // "짤에서 고르기" 펼치기·접기
  onPickMeme?: (stickerId: string) => void
}

// 맞장구 공방: 1~4글자 맞장구를 만들어 보고, 아지트 목록(10칸)에 넣고 순서를 바꾼다.
export function ReactsTab({ reacts, draft: draft0, error: error0, readOnly, stickers, busy, onAdd, onMove, onRemove, onToggleMemes, onPickMeme }: Props) {
  const [draft, setDraft] = useState(draft0)
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
      <span className="hint">메시지에 마우스를 올리면 이 목록이 순서대로 나와요. 아지트 짤은 목록에 없어도 언제든 달 수 있어요.</span>

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
            <form className="row" style={{ gap: 10 }} noValidate onSubmit={(e) => { e.preventDefault(); void add() }}>
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
              <button type="submit" className="btn p" disabled={!canAdd}><Icon name="plus" size={14} /> 목록에 넣기</button>
            </form>
            {error ? <span className="rx-err">{error}</span> : <span className="hint">한글 · 영어 · 숫자 · 이모지 1~4자</span>}

            <div className="row wrap" style={{ gap: 6 }}>
              {SUGGEST.map((s) => {
                const used = reacts.some((r) => r.text === s)
                return (
                  <button key={s} type="button" className={`chip${s === text ? ' on' : ''}`} disabled={used} title={used ? '이미 목록에 있어요' : undefined} onClick={() => setDraft(s)}>{s}</button>
                )
              })}
            </div>

            <div className="rx-ctx card2">
              <span className="rx-ctx-cap">채팅에선 이렇게 보여요</span>
              <div className="row" style={{ gap: 8, alignItems: 'flex-start' }}>
                <span className="rx-ctx-av" style={{ background: 'var(--sky)' }}>류</span>
                <div className="col" style={{ gap: 4 }}>
                  <span className="bub">오늘 엔더 드래곤 잡으러 감?</span>
                  <div className="msg-reacts">
                    <span className="react">ㅋㅋ 3</span>
                    {chars.length > 0 && <span className="react on rx-new">{chars.join('')} 1</span>}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      )}

      <section className="card col" style={{ gap: 12 }}>
        <div className="row">
          <span className="grow b">아지트 맞장구 <span className={full ? 'rx-full' : 'mu'}>{reacts.length}/{MAX_REACTS}</span></span>
          {!readOnly && (
            <button type="button" className={`btn s ${stickers ? 'p' : 'sec'}`} disabled={full && !stickers} onClick={onToggleMemes}>{stickers ? '짤 접기' : '짤에서 고르기'}</button>
          )}
        </div>
        <ol className="rx-grid">
          {slots.map((r, i) => r
            ? (
              <li key={r.id} className={`rx-tile${r.img ? ' meme' : ''}${r.text === text && dup ? ' hit' : ''}`}>
                <span className="rx-pos mono">{i + 1}</span>
                {r.img ? <img src={r.img} alt={r.name} /> : <span className={`rx-tile-text n${Array.from(r.text ?? '').length}`}>{r.text}</span>}
                {r.img && <span className="rx-tile-name">{r.name}</span>}
                {!readOnly && (
                  <div className="rx-tools">
                    <button type="button" className="btn s icon sec" aria-label="앞으로" disabled={i === 0 || busy} onClick={onMove && (() => onMove(i, -1))}><Icon name="back" size={12} /></button>
                    <button type="button" className="btn s icon sec" aria-label="뒤로" disabled={i === reacts.length - 1 || busy} onClick={onMove && (() => onMove(i, 1))}><Icon name="chev" size={12} /></button>
                    <button type="button" className="btn s icon sec del" aria-label={`맞장구 빼기 ${r.text ?? r.name}`} disabled={busy} onClick={onRemove && (() => onRemove(i))}><Icon name="x" size={12} /></button>
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
        {stickers && (
          <div className="stgrid">
            {stickers.map((st) => (
              <button key={st.id} type="button" className={`stbtn${st.inReacts ? ' used' : ''}`} disabled={st.inReacts || full || busy} title={st.name} aria-label={st.name + (st.inReacts ? ' 이미 맞장구에 있음' : ' 맞장구에 넣기')} onClick={onPickMeme && (() => onPickMeme(st.id))}>
                <img src={st.src} alt="" />
              </button>
            ))}
          </div>
        )}
      </section>
    </>
  )
}
