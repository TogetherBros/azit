import { PICK_COLORS } from '../lib/color'

type Props = { selected: string; square?: boolean; label: string }

// 색 고르기. square = 아지트 아이콘 색(네모), 아니면 내 색(동그라미)
export function Swatches({ selected, square, label }: Props) {
  return (
    <div className="swatches" role="group" aria-label={label}>
      {PICK_COLORS.map((c) => (
        <button
          key={c}
          type="button"
          className={`swatch${square ? ' sq' : ''}${c === selected ? ' on' : ''}`}
          style={{ background: c }}
          aria-label={`색 ${c}`}
          aria-pressed={c === selected}
        />
      ))}
    </div>
  )
}
