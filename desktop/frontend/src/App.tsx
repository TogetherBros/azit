import { useEffect, useRef, useState } from 'react'
import { isCancelled, isLoggedIn, loginErrorText, logout, startLogin, type LoginTask } from './api/auth'
import { UpdatePrompt } from './app/UpdatePrompt'
import { WindowFrame } from './app/WindowFrame'
import { MeScreen } from './features/me/MeScreen'
import { LoginScreen } from './features/auth/LoginScreen'

type Phase = 'checking' | 'out' | 'waiting' | 'in'

// 로그인 전이면 LoginScreen, 로그인하면 본 화면.
// 본 화면(MeScreen)은 프로필·홈·아지트 만들기·코드로 들어가기·채팅·아지트 설정(초대 코드·방·맞장구·짤)이 진짜 서버와 이어져 있다.
// 놀이터·알림·멤버 관리처럼 서버 API 가 아직 없는 화면은 디자인 미리보기(dev/Preview.tsx)에만 있다.
export default function App() {
  const [phase, setPhase] = useState<Phase>('checking')
  const [error, setError] = useState<string>()
  const task = useRef<LoginTask | null>(null)

  useEffect(() => {
    isLoggedIn()
      .then((ok) => setPhase(ok ? 'in' : 'out'))
      .catch(() => setPhase('out'))
  }, [])

  const login = () => {
    const t = startLogin()
    task.current = t
    setError(undefined)
    setPhase('waiting')
    t.done
      .then(() => {
        if (task.current === t) setPhase('in')
      })
      .catch((e) => {
        if (task.current !== t || isCancelled(e)) return
        setError(loginErrorText(e))
        setPhase('out')
      })
      .finally(() => {
        if (task.current === t) task.current = null
      })
  }

  const cancel = () => {
    task.current?.cancel()
    task.current = null
    setPhase('out')
  }

  // 서버에 못 알려도 앱의 토큰은 버려지니 결과와 상관없이 로그인 화면으로
  const leave = () => {
    setError(undefined)
    setPhase('out')
    void logout().catch(() => {})
  }

  return (
    <WindowFrame>
      <UpdatePrompt />
      {phase === 'in' ? (
        <MeScreen onLogout={leave} />
      ) : (
        <LoginScreen
          checking={phase === 'checking'}
          waiting={phase === 'waiting'}
          error={error}
          onLogin={login}
          onCancel={cancel}
        />
      )}
    </WindowFrame>
  )
}
