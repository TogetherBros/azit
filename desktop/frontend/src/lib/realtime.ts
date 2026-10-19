// 서버 실시간 이벤트. desktop Go(bridge/realtime.go)가 웹소켓으로 받은 걸 Wails 이벤트로 넘겨 준다.
//   rt      : 서버 이벤트 JSON 그대로 ({"type": "...", "data": {...}})
//   rt:open : 연결됨 (처음이든 다시든). 끊겨 있던 사이 놓친 걸 HTTP 로 다시 받으라는 신호
// 화면은 onRealtime / onReconnect 로 구독하고, 돌려받은 함수로 구독을 끊는다 (useEffect 정리).
import { Events } from '@wailsio/runtime'
import type { MessageRes, NotificationRes } from '../api/server'

export type MessageRef = { azit_id: number; room_id: number; message_id: number }

// 아지트에서 무엇이 바뀌었나 → 그 목록을 다시 받는다
export type AzitChange = { azit_id: number; what: 'rooms' | 'info' | 'members' | 'memes' | 'reacts' }

export type RealtimeEvent =
  | { type: 'message.created'; data: MessageRes }
  | { type: 'message.updated'; data: MessageRef } // 고침·지움·맞장구 → 그 방을 다시 받는다
  | { type: 'notification.created'; data: NotificationRes }
  | { type: 'azit.changed'; data: AzitChange }
  | { type: 'presence'; data: { account_id: number; online: boolean } } // 같은 아지트 사람이 앱을 켰다/껐다

type Listener = (e: RealtimeEvent) => void

const listeners = new Set<Listener>()
const openListeners = new Set<() => void>()
let started = false

// Wails 이벤트는 앱 전체에 한 번만 듣고, 여기서 화면들에게 나눠 준다.
// 브라우저로 화면만 볼 때(vite)는 Wails 가 없어서 실패한다. 그땐 아무 일도 안 일어난다 (화면은 느린 폴링으로 돈다).
function start() {
  if (started) return
  started = true
  try {
    Events.On('rt', (ev) => {
      let e: RealtimeEvent
      try {
        e = JSON.parse(String(ev.data)) as RealtimeEvent
      } catch {
        return
      }
      listeners.forEach((f) => f(e))
    })
    Events.On('rt:open', () => openListeners.forEach((f) => f()))
  } catch {
    // Wails 밖
  }
}

export function onRealtime(f: Listener): () => void {
  start()
  listeners.add(f)
  return () => { listeners.delete(f) }
}

export function onReconnect(f: () => void): () => void {
  start()
  openListeners.add(f)
  return () => { openListeners.delete(f) }
}
