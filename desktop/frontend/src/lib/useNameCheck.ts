import { useEffect, useState } from 'react'
import { checkName, NAME_RULE } from '../api/profile'

export type NameState = 'idle' | 'checking' | 'ok' | 'taken' | 'invalid'

// 이름을 치는 동안 쓸 수 있는지 확인한다. 규칙은 바로 보고, 겹치는지는 멈추고 0.35초 뒤에 서버에 묻는다.
// mine = 지금 내 이름. 그대로면 묻지 않는다 (idle).
// 서버에 못 물으면 idle 로 둔다. 저장할 때 서버가 한 번 더 막는다.
export function useNameCheck(name: string, mine?: string): NameState {
  const [state, setState] = useState<NameState>('idle')

  useEffect(() => {
    if (!name || name === mine) return setState('idle')
    if (!NAME_RULE.test(name)) return setState('invalid')

    setState('checking')
    let live = true
    const t = setTimeout(() => {
      checkName(name)
        .then((r) => live && setState(r.available ? 'ok' : 'taken'))
        .catch(() => live && setState('idle'))
    }, 350)
    return () => {
      live = false
      clearTimeout(t)
    }
  }, [name, mine])

  return state
}

export const NAME_MSG: Record<NameState, string> = {
  idle: '영문 소문자, 숫자, 밑줄(_) · 5~20자',
  checking: '확인하는 중...',
  ok: '쓸 수 있는 이름이에요',
  taken: '이미 누가 쓰고 있어요',
  invalid: '영문 소문자, 숫자, 밑줄(_)만, 5~20자로 써 주세요',
}
