// 아지트·방을 나타내는 둥근 네모. sm 40 · md 42 · lg 72 · xl 80
type Props = { label: string; color: string; size: 'sm' | 'md' | 'lg' | 'xl' }

export function Mark({ label, color, size }: Props) {
  return (
    <span className={`mark ${size}`} style={{ background: color }} aria-hidden="true">
      {label}
    </span>
  )
}
