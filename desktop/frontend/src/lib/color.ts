// 사용자가 고를 수 있는 색 (아지트 아이콘, 내 색)
export const PICK_COLORS = ['#D4FF3A', '#FF5FA8', '#7BC4FF', '#FF8A3D', '#B79CFF', '#F4F4F6']

// 방 이름으로 색을 정한다. 같은 이름이면 누구 화면에서든 같은 색.
export function roomColor(name: string): string {
  let h = 0
  for (let i = 0; i < name.length; i++) h = (h * 31 + name.charCodeAt(i)) >>> 0
  return PICK_COLORS[h % 5]
}

// 아바타·방 네모에 들어갈 첫 글자
export function initial(name: string): string {
  return name.slice(0, 1) || '?'
}
