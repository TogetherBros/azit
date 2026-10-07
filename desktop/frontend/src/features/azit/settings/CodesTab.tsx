import type { CodeState, InviteCode } from '../../../api/types'
import { Seg } from '../../../components/Controls'

type Props = {
  codes: InviteCode[]
  hours?: 1 | 24 | 168
  maxUses?: 0 | 1 | 5
  made?: { code: string; left: string; uses: string } // 방금 만든 코드
}

const LABEL: Record<CodeState, string> = { live: '사용 가능', revoked: '폐기', expired: '만료', used: '다 씀' }

// 초대 코드: 새로 만들기 + 지금까지 만든 코드 목록
export function CodesTab({ codes, hours = 24, maxUses = 5, made }: Props) {
  return (
    <>
      <h1 className="page-title sm">초대 코드</h1>
      <div className="card col" style={{ gap: 14 }}>
        <span className="b">새 코드 만들기</span>
        <div className="row wrap" style={{ gap: 18 }}>
          <div className="field">
            <span className="lbl">유효 시간</span>
            <Seg label="유효 시간" value={hours} options={[[1, '1시간'], [24, '24시간'], [168, '7일']]} />
          </div>
          <div className="field">
            <span className="lbl">사용 횟수</span>
            <Seg label="사용 횟수" value={maxUses} options={[[1, '1번'], [5, '5번'], [0, '제한 없음']]} />
          </div>
          <button type="button" className="btn p" style={{ alignSelf: 'flex-end' }}>만들기</button>
        </div>
        {made && (
          <div className="row code-made">
            <span className="mono">{made.code}</span>
            <span className="grow hint">{made.left} · {made.uses}</span>
            <button type="button" className="btn s p">복사</button>
          </div>
        )}
      </div>

      <div className="col">
        {codes.map((c) => {
          const live = c.state === 'live'
          return (
            <div key={c.code} className={`lrow coderow${live ? '' : ' dead'}`}>
              <span className="coderow-code">{c.code}</span>
              <div className="grow col" style={{ gap: 2 }}>
                <span className="b" style={{ fontSize: 13 }}>{c.madeBy} 만듦 · {live ? c.left : LABEL[c.state]}</span>
                <span className="mu" style={{ fontSize: 12 }}>{c.joined.length ? `${c.joined.join(', ')} 들어옴` : '아직 아무도 안 씀'}</span>
              </div>
              <span className="coderow-uses">{c.uses}/{c.max || '∞'}</span>
              <span className={`tag${live ? ' live' : ''}`}>{LABEL[c.state]}</span>
              {live && (
                <>
                  <button type="button" className="btn s sec">복사</button>
                  <button type="button" className="btn s dg">폐기</button>
                </>
              )}
            </div>
          )
        })}
      </div>
      <span className="hint">코드가 단톡방 밖으로 퍼졌다면 폐기하세요. 폐기해도 이미 들어온 멤버는 그대로예요. 폐기는 되돌릴 수 없어요.</span>
    </>
  )
}
