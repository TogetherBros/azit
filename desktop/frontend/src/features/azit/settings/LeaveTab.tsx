type Props = {
  myMessages: number
  myLiveCodes: number
  wipe?: boolean // 내 메시지 전부 지우기 선택
  presses?: 0 | 1 | 2 // 세 번 눌러서 나가기 진행도
}

const PRESS_LABEL = ['세 번 눌러서 나가기', '한 번 더...', '마지막으로 누르면 나가요']

// 아지트 나가기: 아지트와 방은 남고, 내 흔적만 어떻게 할지 정한다.
export function LeaveTab({ myMessages, myLiveCodes, wipe, presses = 0 }: Props) {
  const choices: [boolean, string, string][] = [
    [false, '남기기', '이름만 "나간 멤버"로 바뀌어요'],
    [true, '전부 지우기', '대화 흐름에 빈칸이 생겨요'],
  ]
  return (
    <>
      <h1 className="page-title sm">아지트 나가기</h1>
      <span className="hint">나가도 아지트와 방은 그대로 남아요. 내 흔적을 어떻게 할지만 정해요.</span>
      <div className="leave">
        <div className="card col">
          <span className="b">내 메시지 <span className="lbl-sub">{myMessages}개</span></span>
          <div className="grid2">
            {choices.map(([v, name, desc]) => (
              <button key={name} type="button" className={`choice${!!wipe === v ? ' on' : ''}`} aria-pressed={!!wipe === v}>
                <span className="choice-name">{name}</span>
                <span className="hint">{desc}</span>
              </button>
            ))}
          </div>
        </div>
        <div className="card hint">
          내가 만든 살아 있는 초대 코드 {myLiveCodes}개는 자동으로 폐기돼요. 칭호와 도감은 이 아지트 것이라 사라져요. 다시 들어오려면 새 초대 코드가 필요해요.
        </div>
        <button type="button" className="btn l dg hold-btn">
          <span className="hold-fill" style={{ width: `${presses * 50}%` }} />
          <span className="hold-label">{PRESS_LABEL[presses]}</span>
        </button>
      </div>
    </>
  )
}
