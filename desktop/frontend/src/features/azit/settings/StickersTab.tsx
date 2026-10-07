import type { ReactOption, Sticker } from '../../../api/types'
import { Icon } from '../../../components/Icon'

type Props = {
  reacts: ReactOption[] // 순서대로 (최대 8개)
  stickers: Sticker[] // 최대 60개
  pickingSticker?: boolean // "짤에서 고르기" 펼침
  renamingId?: string // 이름 바꾸는 중인 짤
}

// 짤 · 맞장구: 맞장구 순서/추가/빼기 + 짤 모음 관리
export function StickersTab({ reacts, stickers, pickingSticker, renamingId }: Props) {
  const full = reacts.length >= 8
  return (
    <>
      <h1 className="page-title sm">짤 · 맞장구</h1>
      <span className="hint">아지트 멤버 모두가 같이 쓰고 같이 고쳐요. 바꾼 건 변경 기록에 남아요.</span>

      <section className="card col" style={{ gap: 12 }}>
        <div className="row">
          <span className="grow b">맞장구 {reacts.length}/8</span>
          <span className="hint">메시지에 마우스를 올리면 이 순서대로 나와요</span>
        </div>
        <div className="row wrap" style={{ gap: 8 }}>
          {reacts.map((r, i) => (
            <div key={r.id} className="row react-edit">
              {r.img ? <img src={r.img} alt={r.name} /> : <span className="react-edit-label">{r.text}</span>}
              <button type="button" className="btn s icon sec" aria-label="앞으로" disabled={i === 0}><Icon name="back" size={13} /></button>
              <button type="button" className="btn s icon sec" aria-label="뒤로" disabled={i === reacts.length - 1}><Icon name="chev" size={13} /></button>
              <button type="button" className="btn s icon sec del" aria-label={`맞장구 빼기 ${r.text ?? r.name}`}><Icon name="x" size={13} /></button>
            </div>
          ))}
        </div>
        <form className="row wrap">
          <label htmlFor="rx-text" className="sr">글자 맞장구</label>
          <input id="rx-text" className="inp" style={{ width: 160 }} maxLength={4} placeholder="글자로 (4자까지)" />
          <button type="submit" className="btn sec" disabled>글자 넣기</button>
          <button type="button" className={`btn ${pickingSticker ? 'p' : 'sec'}`} disabled={full || !stickers.length}>짤에서 고르기</button>
        </form>
        {pickingSticker && (
          <div className="stgrid">
            {stickers.map((st) => (
              <button key={st.id} type="button" className={`stbtn${st.inReacts ? ' used' : ''}`} disabled={st.inReacts} aria-label={st.name + (st.inReacts ? ' 이미 맞장구에 있음' : ' 맞장구에 넣기')}>
                <img src={st.src} alt="" />
              </button>
            ))}
          </div>
        )}
      </section>

      <section className="card col" style={{ gap: 12 }}>
        <div className="row">
          <span className="grow b">짤 {stickers.length}/60</span>
          <button type="button" className="btn p">+ 짤 올리기</button>
        </div>
        <span className="hint">채팅에 올라온 사진이나 짤에 마우스를 올려 "짤로 저장"을 눌러도 여기 들어와요. 5MB까지, 작게 줄여서 저장해요.</span>
        {stickers.length
          ? (
            <div className="st-cards">
              {stickers.map((st) => (
                <div key={st.id} className="card2 st-card">
                  <img src={st.src} alt={st.name} />
                  {renamingId === st.id
                    ? (
                      <form className="row" style={{ gap: 4 }}>
                        <label htmlFor={`sn-${st.id}`} className="sr">짤 이름</label>
                        <input id={`sn-${st.id}`} className="inp" style={{ height: 34, padding: '0 8px' }} maxLength={10} defaultValue={st.name} />
                        <button type="submit" className="btn s p">저장</button>
                      </form>
                    )
                    : <button type="button" className="plain st-card-name">{st.name}<span>이름 바꾸기</span></button>}
                  <span className="mu" style={{ fontSize: 11 }}>{st.by} · {st.ago}</span>
                  <div className="row" style={{ gap: 4 }}>
                    <button type="button" className="btn s sec grow" disabled={st.inReacts || full}>{st.inReacts ? '맞장구에 있음' : '맞장구로'}</button>
                    <button type="button" className="btn s icon dg" aria-label={`짤 빼기 ${st.name}`}><Icon name="x" size={14} /></button>
                  </div>
                </div>
              ))}
            </div>
          )
          : <div className="card2 hint" style={{ textAlign: 'center', padding: 24 }}>아직 짤이 없어요</div>}
      </section>
    </>
  )
}
