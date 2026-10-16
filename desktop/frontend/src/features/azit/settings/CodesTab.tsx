import type { CodeState, InviteCode } from '../../../api/types'
import { Seg } from '../../../components/Controls'

export type CodeExpiry = '30m' | '1h' | '1d' | '7d' | 'never'

type Props = {
  codes: InviteCode[]
  expiresIn?: CodeExpiry
  maxUses?: 0 | 1 | 5 | 10
  made?: { code: string; left: string; uses: string } // 방금 만든 코드
  readOnly?: boolean // 관리자가 아니면 만들 수 없다
  busy?: boolean
  error?: string
  onExpires?: (v: CodeExpiry) => void
  onMaxUses?: (v: 0 | 1 | 5 | 10) => void
  onCreate?: () => void
  onCopy?: (code: string) => void
  onRevoke?: (code: string) => void
}

const LABEL: Record<CodeState, string> = { live: '사용 가능', revoked: '폐기', expired: '만료', used: '다 씀' }

// 초대 코드: 새로 만들기 + 지금 쓸 수 있는 코드 목록
export function CodesTab({ codes, expiresIn = '7d', maxUses = 0, made, readOnly, busy, error, onExpires, onMaxUses, onCreate, onCopy, onRevoke }: Props) {
  return (
    <>
      <h1 className="page-title sm">초대 코드</h1>
      {readOnly
        ? <div className="card2 rx-lock">초대 코드는 관리자만 만들고 볼 수 있어요.</div>
        : (
          <div className="card col" style={{ gap: 14 }}>
            <span className="b">새 코드 만들기</span>
            <div className="row wrap" style={{ gap: 18 }}>
              <div className="field">
                <span className="lbl">유효 시간</span>
                <Seg label="유효 시간" value={expiresIn} onChange={onExpires} options={[['30m', '30분'], ['1h', '1시간'], ['1d', '1일'], ['7d', '7일'], ['never', '계속']]} />
              </div>
              <div className="field">
                <span className="lbl">사용 횟수</span>
                <Seg label="사용 횟수" value={maxUses} onChange={onMaxUses} options={[[1, '1번'], [5, '5번'], [10, '10번'], [0, '제한 없음']]} />
              </div>
              <button type="button" className="btn p" style={{ alignSelf: 'flex-end' }} disabled={busy} onClick={onCreate}>
                {busy ? '만드는 중…' : '만들기'}
              </button>
            </div>
            {error && <span role="alert" className="ferr">{error}</span>}
            {made && (
              <div className="row code-made">
                <span className="mono">{made.code}</span>
                <span className="grow hint">{made.left} · {made.uses}</span>
                <button type="button" className="btn s p" onClick={onCopy && (() => onCopy(made.code))}>복사</button>
              </div>
            )}
          </div>
        )}

      {!readOnly && (
        <div className="col">
          {codes.length === 0 && <div className="card2 hint" style={{ textAlign: 'center', padding: 20 }}>쓸 수 있는 코드가 없어요</div>}
          {codes.map((c) => {
            const live = c.state === 'live'
            return (
              <div key={c.code} className={`lrow coderow${live ? '' : ' dead'}`}>
                <span className="coderow-code">{c.code}</span>
                <div className="grow col" style={{ gap: 2 }}>
                  <span className="b" style={{ fontSize: 13 }}>{c.madeBy} 만듦 · {live ? c.left : LABEL[c.state]}</span>
                  {c.joined.length > 0 && <span className="mu" style={{ fontSize: 12 }}>{c.joined.join(', ')} 들어옴</span>}
                </div>
                <span className="coderow-uses">{c.uses}/{c.max || '∞'}</span>
                <span className={`tag${live ? ' live' : ''}`}>{LABEL[c.state]}</span>
                {live && (
                  <>
                    <button type="button" className="btn s sec" onClick={onCopy && (() => onCopy(c.code))}>복사</button>
                    <button type="button" className="btn s dg" onClick={onRevoke && (() => onRevoke(c.code))}>폐기</button>
                  </>
                )}
              </div>
            )
          })}
        </div>
      )}
      <span className="hint">코드가 단톡방 밖으로 퍼졌다면 폐기하세요. 폐기해도 이미 들어온 멤버는 그대로예요. 폐기는 되돌릴 수 없어요.</span>
    </>
  )
}
