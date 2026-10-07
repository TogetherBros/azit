// 놀이터 게임 목록. 지금은 라이어게임만 되고 나머지는 "준비 중"으로만 보인다.
export type GameInfo = { id: string; name: string; desc: string; ready: boolean; mark: string; bg: string; fg: string }

export const GAMES: GameInfo[] = [
  { id: 'liar', name: '라이어게임', desc: '제시어 모르는 한 명 찾기 · 3~8명', ready: true, mark: '?', bg: 'var(--pinkbg)', fg: 'var(--pink)' },
  { id: 'chain', name: '끝말잇기', desc: '단어로 이어 가기 · 2~8명', ready: false, mark: '가', bg: 'var(--limebg)', fg: 'var(--lime)' },
  { id: 'chosung', name: '초성퀴즈', desc: 'ㄴㄹ 보고 맞히기 · 2~10명', ready: false, mark: 'ㅊ', bg: 'var(--skybg)', fg: 'var(--sky)' },
  { id: 'balance', name: '밸런스게임', desc: '둘 중 하나 고르기 · 인원 무관', ready: false, mark: 'A', bg: 'var(--orbg)', fg: 'var(--orange)' },
  { id: 'draw', name: '그림 맞히기', desc: '그리고 맞히기 · 3~8명', ready: false, mark: '~', bg: 'var(--grapebg)', fg: 'var(--grape)' },
]

export const TOPICS = ['음식', '장소', '직업', '동물']
