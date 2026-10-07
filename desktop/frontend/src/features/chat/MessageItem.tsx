import type { ChatMessage, LiarCard, MessagePart, PollCard, ReactOption, RouletteCard, UserMessage } from '../../api/types'
import { Avatar } from '../../components/Avatar'
import { Icon } from '../../components/Icon'

type Props = {
  msg: ChatMessage
  cont?: boolean // 같은 사람이 이어 보낸 메시지 → 아바타·이름 생략
  showTime?: boolean // 묶음의 마지막 → 시간 표시
  reactOptions: ReactOption[] // 이 아지트의 맞장구 (최대 8개)
}

// 메시지 한 줄. 시스템 메시지(입장, 흔들기 알림)는 가운데 알약 모양.
export function MessageItem({ msg, cont, showTime, reactOptions }: Props) {
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

  return (
    <div className={`msg${mine ? ' me' : ''}`}>
      {!mine && (cont
        ? <span className="msg-av-gap" />
        : (
          <button type="button" className="plain" style={{ flexShrink: 0 }} aria-label={`${author.nick} 프로필`}>
            <Avatar name={author.nick} color={author.color} size={34} />
          </button>
        ))}

      <div className="msg-col">
        {!mine && !cont && (
          <span className="msg-name">
            <button type="button" className="plain">{author.nick}</button>
            {author.title && <span className="msg-title">{author.title}</span>}
          </span>
        )}

        <div className="msg-line">
          <Body msg={msg} />
          {showTime && <span className="msg-time">{msg.time}</span>}
        </div>

        {msg.shookMe && <span className="msg-shook">나를 흔들었어요</span>}

        {msg.reactions && msg.reactions.length > 0 && (
          <div className="msg-reacts">
            {msg.reactions.map((r) => (
              <button key={r.label} type="button" className={`react${r.on ? ' on' : ''}`} aria-pressed={!!r.on} aria-label={`${r.label} ${r.count}`}>
                {r.img ? <img className="rimg" src={r.img} alt="" /> : r.label} {r.count}
              </button>
            ))}
          </div>
        )}
      </div>

      {!msg.deleted && (
        <div className="acts" role="group" aria-label="맞장구">
          {reactOptions.map((r) => (
            <button key={r.id} type="button" aria-label={r.text ?? r.name}>
              {r.img ? <img className="rimg" src={r.img} alt="" /> : r.text}
            </button>
          ))}
          {(msg.kind === 'image' || msg.kind === 'sticker') && <button type="button" className="lime">짤로 저장</button>}
          {!mine && <button type="button">흔들며 답하기</button>}
          {mine && (msg.kind === 'text' || msg.kind === 'image' || msg.kind === 'sticker') && (
            <button type="button" className="red">지우기</button>
          )}
        </div>
      )}
    </div>
  )
}

// 종류별 본문
function Body({ msg }: { msg: UserMessage }) {
  if (msg.deleted) return <div className="bub deleted">지운 메시지</div>
  switch (msg.kind) {
    case 'text':
      return <div className="bub"><Parts parts={msg.parts} /></div>
    case 'image':
      return (
        <button type="button" className="plain" aria-label="사진 크게 보기">
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
