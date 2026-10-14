import { useEffect, useRef, useState } from 'react'
import { isCancelled, isLoggedIn, loginErrorText, startLogin, type LoginTask } from './api/auth'
import { WindowFrame } from './app/WindowFrame'
import { Preview } from './dev/Preview'
import { LoginScreen } from './features/auth/LoginScreen'

type Phase = 'checking' | 'out' | 'waiting' | 'in'

// 로그인 전이면 LoginScreen, 로그인하면 본 화면.
// 본 화면은 아직 진짜가 없어서 미리보기(Preview)를 띄운다.
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

  return (
    <WindowFrame>
      {phase === 'in' ? (
        <Preview />
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
