import type { Member, Sticker } from '../../api/types'
import { Avatar } from '../../components/Avatar'
import { Icon, type IconName } from '../../components/Icon'
import { GameCard } from '../playground/GameCard'
import { GAMES } from '../playground/games'

export type ComposerPopup = 'plus' | 'mention' | 'sticker' | 'games' | null

type Props = {
  draft?: string
  popup?: ComposerPopup
  mentionCandidates?: Member[] // "@강" 까지 쳤을 때 후보
  mentionIndex?: number // 키보드로 고른 후보
  stickers?: Sticker[]
  liarLive?: string // 이 방에 열린 판이 있으면 "모이는 중 3/8 · 들어가기"
}

// + 메뉴 항목: [키, 이름, 설명, 아이콘 색, 아이콘]
const PLUS_ITEMS: [string, string, string, 'lime' | 'pink' | 'sky' | 'grape', IconName][] = [
  ['sticker', '짤', '아지트에 모아 둔 짤 보내기', 'grape', 'plus'],
  ['photo', '사진', '10MB까지 · 크면 줄여서 보내요', 'sky', 'plus'],
  ['shake', '흔들기', '메시지 없이 창 흔들기', 'pink', 'shake'],
  ['roulette', '돌림판', '누가 쏠지 정하기', 'lime', 'plus'],
  ['poll', '투표', '다같이 고르기', 'lime', 'plus'],
  ['games', '게임', '방에서 같이 놀기', 'pink', 'game'],
]

// 입력창은 [+] · 입력칸 · 보내기 세 개뿐. 나머지는 전부 + 안에 (디자인 원칙)
export function Composer({ draft = '', popup = null, mentionCandidates = [], mentionIndex = 0, stickers = [], liarLive }: Props) {
  return (
    <div className="composer">
      {popup === 'mention' && mentionCandidates.length > 0 && (
        <div className="popup mention" role="listbox" aria-label="흔들 사람">
          <span className="popup-title">누구를 흔들까요?</span>
          {mentionCandidates.map((m, i) => (
            <button
              key={m.id}
              type="button"
              role="option"
              aria-selected={i === mentionIndex}
              className={`pitem${i === mentionIndex ? ' on' : ''}`}
            >
              <Avatar name={m.nick} color={m.color} size={26} />
              <span className="pitem-name">{m.nick}</span>
              <span className="pitem-sub">{m.dnd ? '방해 금지라 알림만' : m.online ? '접속 중' : ''}</span>
            </button>
          ))}
        </div>
      )}

      {popup === 'plus' && (
        <div className="popup narrow" role="menu" aria-label="더하기">
          {PLUS_ITEMS.map(([key, name, desc, color, icon]) => (
            <button key={key} type="button" role="menuitem" className="pitem plus-item">
              <span className={`plus-ic ${color}`}><Icon name={icon} size={15} /></span>
              <span className="col" style={{ gap: 1 }}>
                <span className="plus-name">{name}</span>
                <span className="plus-desc">{desc}</span>
              </span>
            </button>
          ))}
        </div>
      )}

      {popup === 'sticker' && (
        <div className="popup wide" role="dialog" aria-label="짤 고르기">
          <div className="popup-bar">
            <button type="button" className="btn s icon sec" aria-label="더하기 메뉴로"><Icon name="back" size={16} /></button>
            <span className="grow b">짤 {stickers.length}</span>
            <button type="button" className="btn s sec">+ 올리기</button>
            <button type="button" className="btn s icon sec" aria-label="닫기"><Icon name="x" size={16} /></button>
          </div>
          {stickers.length > 0
            ? (
              <div className="stgrid scroll">
                {stickers.map((st) => (
                  <button key={st.id} type="button" className="stbtn" aria-label={`짤 ${st.name} 보내기`} title={st.name}>
                    <img src={st.src} alt="" />
                  </button>
                ))}
              </div>
            )
            : <p className="hint" style={{ margin: '6px 4px 10px' }}>아직 짤이 없어요. 올리거나, 채팅에 올라온 사진에서 "짤로 저장"을 눌러 보세요</p>}
        </div>
      )}

      {popup === 'games' && (
        <div className="popup wide" role="dialog" aria-label="게임 고르기">
          <div className="popup-bar">
            <button type="button" className="btn s icon sec" aria-label="더하기 메뉴로"><Icon name="back" size={16} /></button>
            <span className="grow b">게임</span>
            <span className="mu" style={{ fontSize: 11 }}>옆으로 넘겨 보세요</span>
            <button type="button" className="btn s icon sec" aria-label="닫기"><Icon name="x" size={16} /></button>
          </div>
          <div className="gstrip" role="list">
            {GAMES.map((g) => (
              <GameCard key={g.id} game={g} live={g.id === 'liar' && !!liarLive} tag={g.id === 'liar' ? liarLive : undefined} />
            ))}
          </div>
        </div>
      )}

      <div className="row">
        <button
          type="button"
          className={`btn xl-icon ${popup === 'plus' ? 'p' : 'sec'}`}
          aria-label="더하기"
          aria-expanded={popup === 'plus'}
        >
          <Icon name={popup === 'plus' ? 'x' : 'plus'} />
        </button>
        <label htmlFor="composer" className="sr">메시지</label>
        <textarea
          id="composer"
          rows={1}
          className="inp composer-input"
          placeholder="메시지 · @ 치면 흔들기"
          defaultValue={draft}
        />
        <button type="button" className="btn p xl-icon" aria-label="보내기" disabled={!draft.trim()}>
          <Icon name="send" />
        </button>
      </div>
    </div>
  )
}
