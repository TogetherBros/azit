// 창 조작 (테두리 없는 창이라 - □ x 를 직접 그린다)
import { Window } from '@wailsio/runtime'

// 브라우저로 화면만 볼 때(vite)는 Wails 가 없어서 실패한다. 무시한다.
const quiet = (p: Promise<void>) => void p.catch(() => {})

export const minimise = () => quiet(Window.Minimise())
export const toggleMaximise = () => quiet(Window.ToggleMaximise())
export const closeWindow = () => quiet(Window.Close())
