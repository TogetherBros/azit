type Props = {
  owner?: boolean // 만든 사람은 나갈 수 없다
  presses?: 0 | 1 | 2 // 세 번 눌러서 나가기 진행도
  busy?: boolean
  error?: string
  onPress?: () => void
}

const PRESS_LABEL = ['세 번 눌러서 나가기', '한 번 더...', '마지막으로 누르면 나가요']

// 아지트 나가기: 아지트와 방, 내 메시지는 그대로 남는다. 실수로 나가지 않게 세 번 눌러야 나가진다.
export function LeaveTab({ owner, presses = 0, busy, error, onPress }: Props) {
  return (
    <>
      <h1 className="page-title sm">아지트 나가기</h1>
      <span className="hint">나가도 아지트와 방, 내가 보낸 메시지는 그대로 남아요.</span>
      <div className="leave">
        <div className="card hint">
          {owner
            ? '만든 사람은 아지트를 나갈 수 없어요.'
            : '다시 들어오려면 새 초대 코드가 필요해요.'}
        </div>
        {error && <span role="alert" className="ferr">{error}</span>}
        {!owner && (
          <button type="button" className="btn l dg hold-btn" disabled={busy} onClick={onPress}>
            <span className="hold-fill" style={{ width: `${presses * 50}%` }} />
            <span className="hold-label">{busy ? '나가는 중…' : PRESS_LABEL[presses]}</span>
          </button>
        )}
      </div>
    </>
  )
}
