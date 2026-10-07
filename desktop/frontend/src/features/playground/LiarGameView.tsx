import type { LiarPhase, Seat } from '../../api/types'
import { Avatar } from '../../components/Avatar'
import { Icon } from '../../components/Icon'

type Props = {
  room: string
  phase: LiarPhase
  seats: Seat[]
  topic: string
  mode: string // "일반 라이어"
  player: boolean // 내가 참가자인지 (아니면 구경)
  host?: boolean
  // 내 제시어 카드
  word?: string
  amLiar?: boolean
  revealed?: boolean
  // 단계별
  round?: string // "1/2바퀴"
  turn?: string // 지금 말하는 사람
  myTurn?: boolean
  votedCount?: number
  myVote?: string
  accused?: string
  accusedMe?: boolean
  result?: { winner: 'liar' | 'citizen'; liars: string; word: string; note: string }
  chat: { who?: string; color?: string; text: string }[]
}

const PHASES: [LiarPhase, string][] = [
  ['lobby', '대기'],
  ['describe', '설명'],
  ['vote', '투표'],
  ['guess', '역전 기회'],
  ['done', '결과'],
]

// 사람 수에 맞춰 탁자 둘레에 자리를 놓는다 (12시 방향부터 시계 방향)
function seatPos(i: number, n: number) {
  const ang = (i / n) * Math.PI * 2 - Math.PI / 2
  return { left: `${50 + Math.cos(ang) * 38}%`, top: `${50 + Math.sin(ang) * 36}%` }
}

// 라이어게임 판. 왼쪽은 탁자, 오른쪽은 그 방 채팅.
export function LiarGameView(p: Props) {
  const pi = PHASES.findIndex(([ph]) => ph === p.phase)
  const n = p.seats.length

  return (
    <div className="main game">
      <section className="game-stage">
        <header className="head">
          <button type="button" className="btn s icon sec" aria-label="방으로 돌아가기"><Icon name="back" /></button>
          <h1 className="game-title">라이어게임</h1>
          <span className="mu ell grow" style={{ fontSize: 12 }}>{p.room}</span>
          <ol className="phases" aria-label="진행 단계">
            {PHASES.map(([ph, label], i) => (
              <li key={ph} className={`phase${i === pi ? ' now' : i < pi ? ' past' : ''}`} aria-current={i === pi ? 'step' : undefined}>
                {label}
              </li>
            ))}
          </ol>
        </header>

        <div className="table-area">
          <div className="table" />
          <div className="table-center"><Center {...p} n={n} /></div>
          {p.seats.map((s, i) => {
            const votable = p.phase === 'vote' && p.player && !s.me
            const showVotes = (p.phase === 'guess' || p.phase === 'done') && !!s.votes
            return (
              <button
                key={s.member.id}
                type="button"
                className={`seat${s.talking ? ' talking' : ''}${s.myVote ? ' my-vote' : ''}`}
                style={{ ...seatPos(i, n), cursor: votable ? 'pointer' : 'default' }}
                disabled={!votable}
                aria-label={s.member.nick + (votable ? ' 투표하기' : '')}
              >
                <span className="seat-av">
                  <Avatar name={s.member.nick} color={s.member.color} size={50} />
                  {p.phase === 'done' && s.liar && <span className="tag orange seat-liar">라이어</span>}
                  {showVotes && <span className="tag white seat-votes">{s.votes}표</span>}
                </span>
                <span className="seat-name">{s.member.nick}{s.host ? ' · 방장' : ''}{s.me ? ' · 나' : ''}</span>
                {s.desc
                  ? <span className="seat-say">{s.desc}</span>
                  : s.talking && <span className="seat-say now">말하는 중</span>}
                {s.voted && <span className="seat-voted">투표함</span>}
              </button>
            )
          })}
        </div>

        <div className="game-action"><Action {...p} n={n} /></div>
      </section>

      <aside className="game-chat">
        <div className="head"><span className="grow ell">{p.room} 채팅</span></div>
        <div className="scroll mini-chat">
          {p.chat.map((c, i) =>
            c.who
              ? <div key={i}><b style={{ color: c.color }}>{c.who}</b> {c.text}</div>
              : <div key={i} className="sys">{c.text}</div>,
          )}
        </div>
        <form className="row mini-chat-form">
          <label htmlFor="mini-chat" className="sr">채팅</label>
          <input id="mini-chat" className="inp" placeholder="채팅" />
          <button type="submit" className="btn s p" disabled>보내기</button>
        </form>
      </aside>
    </div>
  )
}

// 탁자 가운데: 모인 인원 / 내 제시어 카드 / 구경 중 / 결과
function Center(p: Props & { n: number }) {
  if (p.phase === 'lobby') {
    return (
      <div className="table-msg">
        <span className="table-msg-big">{p.n}/8명 모임</span>
        <span className="mu" style={{ fontSize: 13 }}>3명부터 시작 · {p.topic} · {p.mode}</span>
      </div>
    )
  }
  if (p.phase === 'done' && p.result) {
    const liarWin = p.result.winner === 'liar'
    return (
      <div className="table-msg">
        <span className={`table-msg-big ${liarWin ? 'liar' : 'citizen'}`}>{liarWin ? '라이어 승리' : '시민 승리'}</span>
        <span style={{ fontSize: 14 }}>라이어 <b>{p.result.liars}</b> · 제시어 <b>{p.result.word}</b></span>
        <span className="mu" style={{ fontSize: 12 }}>{p.result.note}</span>
      </div>
    )
  }
  if (!p.player) {
    return (
      <div className="table-msg">
        <span className="table-msg-big" style={{ fontSize: 22 }}>구경 중</span>
        <span className="mu" style={{ fontSize: 12 }}>제시어는 끝나면 공개돼요</span>
      </div>
    )
  }
  const kind = !p.revealed ? '' : p.amLiar ? ' liar' : ' citizen'
  return (
    <button type="button" className={`word-card${kind}`} aria-pressed={!!p.revealed}>
      <span className="word-card-topic">주제 · {p.topic}</span>
      <span className="word-card-word">{p.revealed ? (p.amLiar ? '라이어' : p.word) : '••••'}</span>
      <span className="word-card-hint">
        {!p.revealed ? '눌러서 보기' : p.amLiar ? '남의 설명으로 제시어를 맞혀요' : '눌러서 가리기'}
      </span>
    </button>
  )
}

// 아래 행동 칸: 단계마다 다르다
function Action(p: Props & { n: number }) {
  switch (p.phase) {
    case 'lobby':
      return (
        <div className="row wrap">
          {p.player ? <button type="button" className="btn sec">나가기</button> : <button type="button" className="btn g">참가</button>}
          {p.host
            ? (
              <>
                <button type="button" className="btn sec" disabled={p.n >= 4}>연습봇으로 4명 채우기</button>
                <button type="button" className="btn g" disabled={p.n < 3}>{p.n < 3 ? `${3 - p.n}명 더 필요` : '시작'}</button>
              </>
            )
            : <span className="mu" style={{ fontSize: 13 }}>방장이 시작하면 바로 제시어가 나와요</span>}
        </div>
      )
    case 'describe':
      return p.myTurn
        ? (
          <form className="row">
            <span className="tag pink">내 차례 · {p.round}</span>
            <label htmlFor="liar-say" className="sr">설명</label>
            <input id="liar-say" className="inp grow" placeholder="한 줄 설명 · 너무 티 나면 라이어가 맞혀요" />
            <button type="submit" className="btn g" disabled>말하기</button>
          </form>
        )
        : (
          <div className="row">
            <span className="game-action-text mu">{p.turn} 차례 · {p.round} · 기다리는 중</span>
            {p.player && <button type="button" className="btn s sec">이 차례 넘기기</button>}
          </div>
        )
    case 'vote':
      return (
        <div className="row">
          <span className="game-action-text">
            {!p.player ? '투표 중이에요' : p.myVote ? `${p.myVote}에게 투표했어요 · 바꿀 수 있어요` : '라이어 같은 사람 자리를 눌러 투표해요'}
            {' '}· {p.votedCount}/{p.n}명
          </span>
          {p.host && <button type="button" className="btn s sec">투표 마감</button>}
        </div>
      )
    case 'guess':
      return p.accusedMe
        ? (
          <form className="row">
            <span className="tag orange">잡혔어요 · 역전 기회</span>
            <label htmlFor="liar-guess" className="sr">제시어 추측</label>
            <input id="liar-guess" className="inp grow" placeholder="제시어가 뭐였을까요?" />
            <button type="submit" className="btn g">맞히기</button>
          </form>
        )
        : (
          <div className="row">
            <span className="game-action-text">{p.accused} 님이 잡혔어요 · 제시어를 맞히면 역전</span>
            {p.host && <button type="button" className="btn s sec">역전 기회 넘기기</button>}
          </div>
        )
    case 'done':
      return (
        <div className="row wrap">
          <button type="button" className="btn sec">{p.room}로</button>
          <button type="button" className="btn g">같은 설정으로 한 판 더</button>
          <button type="button" className="btn sec">놀이터</button>
        </div>
      )
  }
}
