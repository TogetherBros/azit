import { useEffect, useRef, useState } from 'react'
import { isCancelled, isLoggedIn, loginErrorText, logout, startLogin, type LoginTask } from './api/auth'
import { WindowFrame } from './app/WindowFrame'
import { MeScreen } from './features/me/MeScreen'
import { LoginScreen } from './features/auth/LoginScreen'

type Phase = 'checking' | 'out' | 'waiting' | 'in'

// 로그인 전이면 LoginScreen, 로그인하면 본 화면.
// 본 화면은 아직 프로필 만들기·내 정보(MeScreen)만 진짜다. 디자인 미리보기는 dev/Preview.tsx.
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
