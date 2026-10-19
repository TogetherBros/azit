// 앱 설정: 이 컴퓨터에만 저장한다 (서버에는 안 간다). 화면에 바로 적용되는 건 여기서 적용한다.

export type FontSize = 's' | 'm' | 'l'

export type AppPrefs = {
  shakeWindow: boolean // 흔들기 받으면 창이 흔들림
  fontSize: FontSize
  reduceMotion: boolean
  clock: '12' | '24'
  autostart: boolean // (아직 준비 중)
  closeToTray: boolean // (아직 준비 중)
}

export const DEFAULT_PREFS: AppPrefs = {
  shakeWindow: true,
  fontSize: 'm',
  reduceMotion: false,
  clock: '12',
  autostart: false,
  closeToTray: false,
}

const KEY = 'azit.prefs'

let current: AppPrefs = load()
apply(current)

function load(): AppPrefs {
  try {
    const raw = localStorage.getItem(KEY)
    return raw ? { ...DEFAULT_PREFS, ...(JSON.parse(raw) as Partial<AppPrefs>) } : DEFAULT_PREFS
  } catch {
    return DEFAULT_PREFS
  }
}

export function getPrefs(): AppPrefs {
  return current
}

export function setPrefs(next: AppPrefs) {
  current = next
  try {
    localStorage.setItem(KEY, JSON.stringify(next))
  } catch {
    // 저장이 막혀 있어도 이번에 켜 둔 동안은 적용된다
  }
  apply(next)
}

// 글자 크기·움직임 줄이기는 <html> 에 클래스를 달아 CSS 가 처리한다 (base.css)
function apply(p: AppPrefs) {
  if (typeof document === 'undefined') return
  const root = document.documentElement
  root.classList.remove('fs-s', 'fs-m', 'fs-l')
  root.classList.add(`fs-${p.fontSize}`)
  root.classList.toggle('reduce-motion', p.reduceMotion)
}
