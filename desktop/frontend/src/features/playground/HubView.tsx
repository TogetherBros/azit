import type { GameRecord, Member, OpenGame, Stats } from '../../api/types'
import { Avatar } from '../../components/Avatar'
import { Stat } from '../../components/Controls'
import { GameCard } from './GameCard'
import { GAMES } from './games'
import { OpenGameForm } from './OpenGameForm'

type Props = {
  openGames: OpenGame[]
  records: GameRecord[]
  ranking: { member: Member; wins: number; me?: boolean }[]
  myStats: Stats
  rooms: { id: string; name: string; busy?: boolean }[]
  roomId: string
}

// 놀이터: 게임 목록 · 지금 열린 판 · 판 열기 · 최근 판 | 순위 · 내 전적
export function HubView({ openGames, records, ranking, myStats, rooms, roomId }: Props) {
  return (
    <div className="main">
      <div className="scroll page hub">
        <div className="hub-main">
          <div className="col" style={{ gap: 4 }}>
            <h1 className="page-title">놀이터</h1>
            <span className="mu">아지트에서 같이 하는 게임 · 판은 방마다 하나씩 열 수 있어요</span>
          </div>

          <section className="col">
            <h2 className="sec-title">게임</h2>
            <div className="gstrip">{GAMES.map((g) => <GameCard key={g.id} game={g} />)}</div>
          </section>

          <section className="col">
            <h2 className="sec-title">지금 열린 판</h2>
            {openGames.length
              ? (
                <div className="grid3">
                  {openGames.map((g) => (
                    <div key={g.id} className={`open-game${g.status === 'lobby' ? ' lobby' : ''}`}>
                      <span className="open-game-room">{g.room}</span>
                      <span className="open-game-info">
                        {g.status === 'lobby' ? `대기 ${g.players}/8` : `진행 중 · ${g.players}명`} · {g.topic}
                      </span>
                      <button type="button" className="btn s">
                        {g.joined ? '판으로 가기' : g.status === 'lobby' ? '참가' : '구경'}
                      </button>
                    </div>
                  ))}
                </div>
              )
              : <div className="card mu">열린 판이 없어요. 아래에서 하나 열어 보세요</div>}
          </section>

          <section className="card col" style={{ gap: 14 }}>
            <h2 className="sec-title">라이어게임 판 열기</h2>
            <OpenGameForm rooms={rooms} roomId={roomId} />
          </section>

          <section className="col">
            <h2 className="sec-title">최근 판</h2>
            {records.length
              ? records.map((r) => (
                <button key={r.id} type="button" className="lrow record">
                  <span className={`tag ${r.winner === 'liar' ? 'liar-win' : 'citizen-win'}`}>
                    {r.winner === 'liar' ? '라이어 승' : '시민 승'}
                  </span>
                  <span className="grow" style={{ fontSize: 13 }}>{r.text}</span>
                  <span className="mu" style={{ fontSize: 11 }}>{r.ago}</span>
                </button>
              ))
              : <span className="mu" style={{ fontSize: 13 }}>아직 끝난 판이 없어요</span>}
          </section>
        </div>

        <aside className="hub-side">
          <div className="card col" style={{ gap: 10 }}>
            <span className="disp" style={{ fontSize: 18 }}>라이어게임 순위</span>
            {ranking.map((r, i) => (
              <div key={r.member.id} className={`rank-row${r.me ? ' me' : ''}`}>
                <span className={`rank-no${i === 0 ? ' first' : ''}`}>{i + 1}</span>
                <Avatar name={r.member.nick} color={r.member.color} size={28} />
                <button type="button" className="plain rank-name">{r.member.nick}</button>
                <span className="rank-wins">{r.wins}승</span>
              </div>
            ))}
          </div>
          <div className="card col" style={{ gap: 10 }}>
            <span className="b">내 라이어게임 전적</span>
            <div className="grid2" style={{ gap: 8 }}>
              <Stat value={myStats.games} label="판" />
              <Stat value={myStats.catches} label="라이어 검거" />
              <Stat value={`${myStats.liarSurvive}/${myStats.liarGames}`} label="라이어로 생존" />
              <Stat value={myStats.snipes} label="제시어 저격" />
            </div>
          </div>
        </aside>
      </div>
    </div>
  )
}
