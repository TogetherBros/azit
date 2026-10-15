// 로그인 관련 Go 호출. 화면은 바인딩 대신 이 파일을 쓴다.
import { CancelError } from '@wailsio/runtime'
import { AuthService } from '../../bindings/github.com/TogetherBros/azit/desktop/internal/bridge'

export type LoginTask = {
  done: Promise<void>
  cancel: () => void
}

// 기본 브라우저로 구글 로그인을 시작한다. cancel 하면 Go 쪽 기다림도 멈춘다 (ctx 취소).
export function startLogin(): LoginTask {
  const call = AuthService.Login()
  return { done: call, cancel: () => void call.cancel() }
}

export function isLoggedIn(): Promise<boolean> {
  return AuthService.IsLoggedIn()
}

export function isCancelled(e: unknown): boolean {
  return e instanceof CancelError
}

// Go 에서 온 오류를 화면에 띄울 문장으로 바꾼다.
export function loginErrorText(e: unknown): string {
  const msg = e instanceof Error ? e.message : String(e)
  if (msg.startsWith('구글 로그인 취소')) return '구글 로그인을 취소했어요. 다시 들어가려면 버튼을 눌러 주세요'
  if (/refused|dial tcp|connect/i.test(msg)) return '서버에 연결하지 못했어요. 잠시 뒤에 다시 해 주세요'
  if (/[가-힣]/.test(msg) && !msg.includes('state')) return msg
  return '로그인하지 못했어요. 다시 해 주세요'
}

// 로그아웃. 서버가 꺼져 있어도 앱의 토큰은 버려진다.
export async function logout(): Promise<void> {
  await AuthService.Logout()
}
