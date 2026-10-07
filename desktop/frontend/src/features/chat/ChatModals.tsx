// + 메뉴와 메시지 메뉴에서 뜨는 모달들: 돌림판, 투표, 사진 크게 보기, 메시지 지우기
import { Icon } from '../../components/Icon'
import { Modal } from '../../components/Modal'

type RouletteProps = {
  title?: string
  candidates: string[] // 멤버 닉네임 + 직접 넣은 후보
  selected: string[]
}

// 돌림판 만들기: 후보를 고르고 돌린다. 기본으로 접속 중인 사람이 골라져 있다.
export function RouletteModal({ title, candidates, selected }: RouletteProps) {
  return (
    <Modal title="돌림판">
      <div className="field">
        <label htmlFor="rl-title">뭘 정할까요?</label>
        <input id="rl-title" className="inp" placeholder="예: 점심 쏘기" defaultValue={title} />
      </div>
      <div className="field">
        <span className="lbl">후보 {selected.length}개</span>
        <div className="row wrap scroll" style={{ gap: 6, maxHeight: 160 }}>
          {candidates.map((c) => {
            const on = selected.includes(c)
            return <button key={c} type="button" className={`chip${on ? ' on' : ''}`} aria-pressed={on}>{c}</button>
          })}
        </div>
        <div className="row">
          <label htmlFor="rl-extra" className="sr">후보 추가</label>
          <input id="rl-extra" className="inp" placeholder="사람 말고 다른 후보 (예: 떡볶이)" />
          <button type="button" className="btn sec" disabled>추가</button>
        </div>
      </div>
      <button type="button" className="btn p" disabled={selected.length < 2}>
        {selected.length < 2 ? '후보를 2개 이상 골라요' : '돌리기'}
      </button>
    </Modal>
  )
}

type PollProps = { question?: string; options: string[] }

// 투표 올리기: 질문 + 보기 2~5개
export function PollModal({ question = '', options }: PollProps) {
  const ok = question.trim() !== '' && options.filter((o) => o.trim()).length >= 2
  return (
    <Modal title="투표">
      <div className="field">
        <label htmlFor="pl-q">질문</label>
        <input id="pl-q" className="inp" placeholder="예: 토요일 몇 시?" defaultValue={question} />
      </div>
      <div className="field">
        <span className="lbl">보기</span>
        {options.map((o, i) => (
          <div key={i} className="row">
            <label htmlFor={`pl-${i}`} className="sr">보기 {i + 1}</label>
            <input id={`pl-${i}`} className="inp" placeholder={`보기 ${i + 1}`} defaultValue={o} />
            {options.length > 2 && (
              <button type="button" className="btn icon sec" aria-label="보기 빼기"><Icon name="x" size={14} /></button>
            )}
          </div>
        ))}
        {options.length < 5 && <button type="button" className="btn s gh" style={{ alignSelf: 'flex-start' }}>+ 보기</button>}
      </div>
      <button type="button" className="btn p" disabled={!ok}>{ok ? '올리기' : '질문과 보기 2개가 필요해요'}</button>
    </Modal>
  )
}

export function ImageModal({ src }: { src: string }) {
  return (
    <Modal title="사진">
      <img src={src} alt="사진" style={{ borderRadius: 16 }} />
    </Modal>
  )
}

// 내 메시지 지우기 확인
export function DeleteMessageModal() {
  return (
    <Modal>
      <h2 className="disp modal-title">이 메시지를 지울까요?</h2>
      <p className="modal-text">모두의 화면에서 "지운 메시지"로 바뀌어요. 되돌릴 수 없어요.</p>
      <div className="modal-actions">
        <button type="button" className="btn sec">취소</button>
        <button type="button" className="btn dgs">지우기</button>
      </div>
    </Modal>
  )
}
