// 작은 입력 부품: 토글, 세그먼트(하나 고르기), 숫자 칸

type ToggleProps = { on: boolean; label: string }

export function Toggle({ on, label }: ToggleProps) {
  return (
    <button type="button" className={`tgl${on ? ' on' : ''}`} aria-pressed={on} aria-label={label}>
      <span />
    </button>
  )
}

type SegProps<T extends string | number> = { label: string; value: T; options: [T, string][] }

export function Seg<T extends string | number>({ label, value, options }: SegProps<T>) {
  return (
    <div className="seg" role="group" aria-label={label}>
      {options.map(([v, text]) => (
        <button key={String(v)} type="button" className={v === value ? 'on' : ''} aria-pressed={v === value}>
          {text}
        </button>
      ))}
    </div>
  )
}

type StatProps = { value: number | string; label: string }

// 카드 안의 숫자 칸 (판 수, 라이어 검거 ...)
export function Stat({ value, label }: StatProps) {
  return (
    <div className="card2 stat">
      <div className="stat-num">{value}</div>
      <div className="stat-label">{label}</div>
    </div>
  )
}
