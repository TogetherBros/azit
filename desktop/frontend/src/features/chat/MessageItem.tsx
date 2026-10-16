import { useCallback, useState, type MouseEvent } from 'react'
import type { ChatMessage, LiarCard, MessagePart, PollCard, Reaction, ReactOption, RouletteCard, UserMessage } from '../../api/types'
import { Avatar } from '../../components/Avatar'
import { Icon } from '../../components/Icon'
import { ReactPicker } from './ReactPicker'

type Props = {
  msg: ChatMessage
  cont?: boolean // 같은 사람이 이어 보낸 메시지 → 아바타·이름 생략
  showTime?: boolean // 묶음의 마지막 → 시간 표시
  reactOptions: ReactOption[] // 이 아지트의 맞장구 (최대 10개)
  memeOptions?: ReactOption[] // 아지트 짤 전부 (맞장구 고르기 창 아래 칸)
  onReact?: (msg: UserMessage, r: ReactOption) => void // 메뉴에서 맞장구 달기
  onToggle?: (msg: UserMessage, r: Reaction) => void // 달린 맞장구 누르기 (내 거면 떼기)
  onEdit?: (msg: UserMessage) => void
  onDelete?: (msg: UserMessage) => void
  onRetry?: (msg: UserMessage) => void
  onImage?: (src: string) => void // 사진 크게 보기
  onProfile?: (accountId: string) => void // 아바타·이름 누르기 → 프로필
  canDelete?: boolean // 남의 메시지도 지울 수 있는지 (관리자)
}

// 메시지 한 줄. 시스템 메시지(입장, 흔들기 알림)는 가운데 알약 모양.
export function MessageItem({ msg, cont, showTime, reactOptions, memeOptions = [], onReact, onToggle, onEdit, onDelete, onRetry, onImage, onProfile, canDelete }: Props) {
  const [picker, setPicker] = useState<HTMLElement | null>(null) // 열린 맞장구 고르기 창의 기준 버튼
  const closePicker = useCallback(() => setPicker(null), [])

  if (msg.kind === 'sys') {
    return (
      <div className="sysm">
        {msg.shake && <span className="sysm-icon"><Icon name="shake" size={14} /></span>}
        <span>{msg.text}</span>
        {msg.action && <button type="button" className="plain">{msg.action}</button>}
      </div>
    )
  }

  const { author, mine } = msg
  const canReact = !!onReact && !msg.deleted && !msg.pending && !msg.failed
  const isOn = (o: ReactOption) => !!msg.reactions?.some((r) => r.on && (o.memeId ? r.memeId === o.memeId : r.text === o.text))
  const togglePicker = (e: MouseEvent<HTMLButtonElement>) => {
    const el = e.currentTarget
    setPicker((cur) => (cur === el ? null : el))
  }

  return (
    <div className={`msg${mine ? ' me' : ''}${msg.pending ? ' pending' : ''}`}>
      {!mine && (cont
        ? <span className="msg-av-gap" />
        : (
          <button type="button" className="plain" style={{ flexShrink: 0 }} aria-label={`${author.nick} 프로필`} onClick={onProfile && (() => onProfile(author.id))}>
            <Avatar name={author.nick} color={author.color} size={34} src={author.src} />
          </button>
        ))}

      <div className="msg-col">
        {!mine && !cont && (
          <span className="msg-name">
            <button type="button" className="plain" onClick={onProfile && (() => onProfile(author.id))}>{author.nick}</button>
            {author.title && <span className="msg-title">{author.title}</span>}
          </span>
        )}

        <div className="msg-line">
          <Body msg={msg} onImage={onImage} />
          {msg.failed
            ? <button type="button" className="plain msg-failed" onClick={onRetry && (() => onRetry(msg))}>못 보냈어요 · 다시 보내기</button>
            : (showTime || msg.edited) && (
              <span className="msg-time">{msg.edited && '고침 '}{showTime && msg.time}</span>
            )}
        </div>

        {msg.shookMe && <span className="msg-shook">나를 흔들었어요</span>}

        {msg.reactions && msg.reactions.length > 0 && (
          <div className="msg-reacts">
            {msg.reactions.map((r) => (
              <button key={r.label} type="button" className={`react${r.on ? ' on' : ''}`} aria-pressed={!!r.on} aria-label={`${r.label} ${r.count}`} title={r.img ? r.label : undefined} onClick={onToggle && (() => onToggle(msg, r))}>
                {r.img ? <img className="rimg" src={r.img} alt="" /> : <span className="react-label">{r.label}</span>}
                <span className="react-n">{r.count}</span>
              </button>
            ))}
            {canReact && (
              <button type="button" className="react add" aria-label="맞장구 달기" title="맞장구 달기" onClick={togglePicker}>
                <Icon name="smile" size={15} />
              </button>
            )}
          </div>
        )}
      </div>

      {!msg.deleted && !msg.pending && !msg.failed && (
        <div className={`acts${picker ? ' open' : ''}`} role="group" aria-label="메시지 메뉴">
          {/* 미리보기 화면(onReact 없음)은 예전처럼 맞장구 목록을 그대로 보여 준다 */}
          {!onReact && reactOptions.map((r) => (
            <button key={r.id} type="button" aria-label={r.text ?? r.name} title={r.name}>
              {r.img ? <img className="rimg" src={r.img} alt="" /> : r.text}
            </button>
          ))}
          {!onReact && (msg.kind === 'image' || msg.kind === 'sticker') && <button type="button" className="lime">짤로 저장</button>}
          {!onReact && !mine && <button type="button">흔들며 답하기</button>}
          {canReact && (
            <button type="button" className="ic" aria-label="맞장구" title="맞장구" aria-expanded={!!picker} onClick={togglePicker}>
              <Icon name="smile" size={17} />
            </button>
          )}
          {mine && onEdit && msg.kind === 'text' && (
            <button type="button" className="ic" aria-label="고치기" title="고치기" onClick={() => onEdit(msg)}>
              <Icon name="edit" size={16} />
            </button>
          )}
          {(mine || canDelete) && (msg.kind === 'text' || msg.kind === 'image' || msg.kind === 'sticker') && (
            <button type="button" className="ic red" aria-label="지우기" title="지우기" onClick={onDelete && (() => onDelete(msg))}>
              <Icon name="trash" size={16} />
            </button>
          )}
        </div>
      )}

      {picker && onReact && (
        <ReactPicker
          anchor={picker}
          reacts={reactOptions}
          memes={memeOptions}
          isOn={isOn}
          onPick={(o) => onReact(msg, o)}
          onClose={closePicker}
        />
      )}
    </div>
  )
}

// 종류별 본문
function Body({ msg, onImage }: { msg: UserMessage; onImage?: (src: string) => void }) {
  if (msg.deleted) return <div className="bub deleted">지운 메시지</div>
  switch (msg.kind) {
    case 'text':
      return <div className="bub"><Parts parts={msg.parts} /></div>
    case 'image':
      return (
        <button type="button" className="plain" aria-label="사진 크게 보기" onClick={onImage && (() => onImage(msg.src))}>
          <img className="msg-img" src={msg.src} alt="사진" />
        </button>
      )
    case 'sticker':
      return <img className="msg-sticker" src={msg.src} alt={`짤 ${msg.name}`} />
    case 'game':
      return <LiarCardView game={msg.game} />
    case 'roulette':
      return <RouletteView r={msg.roulette} />
    case 'poll':
      return <PollView poll={msg.poll} />
  }
}

// 본문: 글자 + @이름(흔들기) 조각
function Parts({ parts }: { parts: MessagePart[] }) {
  return (
    <>
      {parts.map((p, i) =>
        'text' in p
          ? <span key={i}>{p.text}</span>
          : (
            <span key={i} className={`mention${p.mine ? ' mine' : ''}`}>
              {p.shook ? <Icon name="shake" size={12} stroke={2.8} /> : '@'}
              {p.mention}
            </span>
          ),
      )}
    </>
  )
}

// 채팅에 올라온 라이어 판
function LiarCardView({ game }: { game: LiarCard }) {
  const ended = game.status === 'done' || game.status === 'cancel'
  const status = game.status === 'lobby' ? `대기 ${game.players}/8` : game.status === 'playing' ? '진행 중' : game.status === 'done' ? '끝' : '취소됨'
  return (
    <div className={`liar-card${ended ? ' done' : ''}`}>
      <div className="row">
        <span className="liar-card-title">라이어게임 판</span>
        <span className="tag">{status}</span>
      </div>
      <span className="liar-card-info">{game.info}</span>
      {game.result && <span className="liar-card-result">{game.result}</span>}
      <div className="row">
        {game.status === 'lobby' && !game.joined && <button type="button" className="btn s on-pink">참가</button>}
        <button type="button" className="btn s ghost-pink">
          {ended ? '결과 보기' : game.joined ? '판으로 가기' : '구경'}
        </button>
      </div>
    </div>
  )
}

function RouletteView({ r }: { r: RouletteCard }) {
  return (
    <div className="card2 wheel-card">
      <div className="row">
        <span className="tag lime">돌림판</span>
        <span className="mu grow" style={{ fontSize: 13 }}>{r.title || '제목 없음'}</span>
      </div>
      <div className="wheel-opts">
        {r.options.map((o) => (
          <span key={o} className={`wheel-opt${o === r.result ? ' win' : ''}`}>{o}{o === r.result ? ' 당첨' : ''}</span>
        ))}
      </div>
      <button type="button" className="btn s sec" style={{ alignSelf: 'flex-start' }} disabled={r.spinsLeft <= 0}>
        {r.spinsLeft <= 0 ? '다시 돌리기 끝' : `다시 돌리기 · ${r.spinsLeft}번 남음`}
      </button>
    </div>
  )
}

function PollView({ poll }: { poll: PollCard }) {
  return (
    <div className="card2 poll-card">
      <div className="row">
        <span className="tag sky">투표</span>
        <span className="poll-q">{poll.question}</span>
        {poll.closed && <span className="tag">마감</span>}
      </div>
      {poll.options.map((o) => {
        const pct = poll.total ? Math.round((o.count / poll.total) * 100) : 0
        return (
          <button key={o.label} type="button" className={`poll-opt${o.mine ? ' mine' : ''}`} disabled={poll.closed} aria-pressed={!!o.mine}>
            <span className="poll-bar" style={{ width: `${pct}%` }} />
            <span className="grow">{o.label}</span>
            <span className="poll-count">{o.count}</span>
          </button>
        )
      })}
      {poll.canClose && !poll.closed && <button type="button" className="btn s sec" style={{ alignSelf: 'flex-start' }}>마감하기</button>}
      <span className="hint">{poll.total}명 참여</span>
    </div>
  )
}
