import { initial } from '../lib/color'

type Props = { name: string; color?: string; size?: number; src?: string }

// 사람 동그라미. 사진이 있으면 사진, 없으면 첫 글자 + 색.
// 크기와 색은 데이터마다 달라서 style 로 넘긴다.
export function Avatar({ name, color, size = 32, src }: Props) {
  return (
    <span
      className="av"
      style={{ width: size, height: size, background: color, fontSize: Math.round(size * 0.4) }}
      aria-hidden="true"
    >
      {src ? <img src={src} alt="" className="av-img" /> : initial(name)}
    </span>
  )
}
