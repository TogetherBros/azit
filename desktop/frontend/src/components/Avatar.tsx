import { initial } from '../lib/color'

type Props = { name: string; color?: string; size?: number }

// 사람 동그라미. 크기와 색은 데이터마다 달라서 style 로 넘긴다.
export function Avatar({ name, color, size = 32 }: Props) {
  return (
    <span
      className="av"
      style={{ width: size, height: size, background: color, fontSize: Math.round(size * 0.4) }}
      aria-hidden="true"
    >
      {initial(name)}
    </span>
  )
}
