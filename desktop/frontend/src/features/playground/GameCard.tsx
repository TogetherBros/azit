import type { GameInfo } from './games'

type Props = { game: GameInfo; tag?: string; live?: boolean }

// 옆으로 넘기는 게임 카드 (+ 메뉴의 게임 고르기, 놀이터 맨 위)
export function GameCard({ game, tag, live }: Props) {
  const label = tag ?? (game.ready ? '판 열기' : '준비 중')
  return (
    <button type="button" className="gcard" disabled={!game.ready} aria-label={game.name + (game.ready ? '' : ' 준비 중')}>
      <span className="gart" style={{ background: game.bg, color: game.fg }}>{game.mark}</span>
      <span className="gcard-name">{game.name}</span>
      <span className="gcard-desc">{game.desc}</span>
      <span className={`tag${game.ready ? (live ? ' pink' : ' lime') : ''}`}>{label}</span>
    </button>
  )
}
