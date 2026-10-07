type Props = { text: string; action?: string }

// 화면 아래 잠깐 뜨는 알림. action 이 있으면 "되돌리기" 같은 버튼이 붙는다.
export function Toast({ text, action }: Props) {
  return (
    <div className="toast" role="status">
      <span>{text}</span>
      {action && <button type="button">{action}</button>}
    </div>
  )
}
