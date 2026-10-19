// 선 아이콘 모음. 색은 currentColor 라서 부모 글자색을 따라간다.
export type IconName =
  | 'shake' | 'search' | 'plus' | 'gear' | 'bell' | 'send' | 'back'
  | 'x' | 'check' | 'chev' | 'wide' | 'narrow' | 'menu' | 'game' | 'star'
  | 'min' | 'max' | 'smile' | 'edit' | 'trash' | 'reply' | 'save'

const PATHS: Partial<Record<IconName, string>> = {
  shake: 'M2 12l3-6 4 12 4-12 4 12 3-6h2',
  plus: 'M12 5v14M5 12h14',
  gear: 'M4 6h9M17 6h3M4 12h3M11 12h9M4 18h11M19 18h1',
  bell: 'M18 8a6 6 0 0 0-12 0c0 7-3 9-3 9h18s-3-2-3-9M13.7 21a2 2 0 0 1-3.4 0',
  send: 'M22 2L11 13M22 2l-7 20-4-9-9-4z',
  back: 'M15 18l-6-6 6-6',
  x: 'M18 6L6 18M6 6l12 12',
  check: 'M20 6L9 17l-5-5',
  chev: 'M9 6l6 6-6 6',
  wide: 'M3 5h18v14H3zM9 5v14',
  narrow: 'M7 2h10v20H7z',
  menu: 'M4 7h16M4 12h16M4 17h16',
  min: 'M5 12h14',
  max: 'M5 5h14v14H5z',
  edit: 'M12 20h9M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4z',
  trash: 'M3 6h18M8 6V4h8v2M6 6l1 14h10l1-14M10 11v5M14 11v5',
  reply: 'M9 14L4 9l5-5M4 9h10a6 6 0 0 1 6 6v5',
  save: 'M12 3v12M7 10l5 5 5-5M5 21h14',
}

type Props = { name: IconName; size?: number; stroke?: number }

export function Icon({ name, size = 18, stroke = 2.2 }: Props) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={stroke}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      {name === 'search' ? (
        <>
          <circle cx="11" cy="11" r="7" />
          <path d="M21 21l-4.3-4.3" />
        </>
      ) : name === 'game' ? (
        <>
          <rect x="2" y="6" width="20" height="12" rx="5" />
          <path d="M7 10v4M5 12h4" />
        </>
      ) : name === 'smile' ? (
        <>
          <circle cx="12" cy="12" r="9" />
          <path d="M8.5 14.5a4.5 4.5 0 0 0 7 0M9 9.5h.01M15 9.5h.01" />
        </>
      ) : name === 'star' ? (
        <path d="M12 2l3 7 7 .6-5.3 4.6 1.6 7-6.3-3.8-6.3 3.8 1.6-7L2 9.6 9 9z" fill="currentColor" />
      ) : (
        <path d={PATHS[name]} />
      )}
    </svg>
  )
}
