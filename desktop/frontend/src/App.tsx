import { Preview } from './dev/Preview'

// 지금은 화면 미리보기만 띄운다. JS 단계에서 로그인 여부에 따라
// LoginScreen / JoinScreen / Shell 중 하나를 고르는 진짜 App 으로 바꾼다.
export default function App() {
  return <Preview />
}
