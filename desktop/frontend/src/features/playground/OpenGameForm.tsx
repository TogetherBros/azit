import { Seg } from '../../components/Controls'
import { Modal } from '../../components/Modal'
import { TOPICS } from './games'

type Props = {
  rooms: { id: string; name: string; busy?: boolean }[]
  roomId: string
  topic?: string
  mode?: 'normal' | 'fool'
  liars?: 1 | 2
  rounds?: 1 | 2 | 3
}

// 라이어 판 열기 설정. 놀이터 안에도, 채팅 + 메뉴의 모달 안에도 들어간다.
export function OpenGameForm({ rooms, roomId, topic = '음식', mode = 'normal', liars = 1, rounds = 2 }: Props) {
  const room = rooms.find((r) => r.id === roomId)
  const busy = !!room?.busy
  return (
    <div className="col" style={{ gap: 14 }}>
      <div className="field">
        <label htmlFor="og-room">어느 방에 열까요?</label>
        <select id="og-room" className="inp" defaultValue={roomId}>
          {rooms.map((r) => <option key={r.id} value={r.id}>{r.name}{r.busy ? ' (판 있음)' : ''}</option>)}
        </select>
      </div>
      <div className="field">
        <span className="lbl">주제</span>
        <div className="row wrap" style={{ gap: 6 }}>
          {TOPICS.map((t) => (
            <button key={t} type="button" className={`chip${t === topic ? ' on' : ''}`} aria-pressed={t === topic}>{t}</button>
          ))}
        </div>
      </div>
      <div className="grid3">
        <div className="field">
          <span className="lbl">방식</span>
          <Seg label="방식" value={mode} options={[['normal', '일반'], ['fool', '바보']]} />
          <span className="hint">{mode === 'fool' ? '라이어도 비슷한 다른 제시어를 받아요' : '라이어는 주제만 알아요'}</span>
        </div>
        <div className="field">
          <span className="lbl">라이어 수</span>
          <Seg label="라이어 수" value={liars} options={[[1, '1명'], [2, '2명']]} />
          <span className="hint">2명은 7명 이상일 때만</span>
        </div>
        <div className="field">
          <span className="lbl">설명 바퀴</span>
          <Seg label="설명 바퀴" value={rounds} options={[[1, '1'], [2, '2'], [3, '3']]} />
        </div>
      </div>
      {busy && <span className="ferr">이 방엔 이미 판이 열려 있어요</span>}
      <button type="button" className="btn g" disabled={busy}>{room?.name}에 판 열기</button>
    </div>
  )
}

export function OpenGameModal(props: Props) {
  return (
    <Modal title="라이어 판 열기">
      <OpenGameForm {...props} />
    </Modal>
  )
}
