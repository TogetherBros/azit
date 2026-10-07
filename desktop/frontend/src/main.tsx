import React from 'react'
import ReactDOM from 'react-dom/client'
// 폰트는 앱 안에 같이 묶는다 (인터넷 없이도 보이게)
import 'pretendard/dist/web/variable/pretendardvariable-dynamic-subset.css'
import '@fontsource/bagel-fat-one'
import '@fontsource/space-mono/400.css'
import '@fontsource/space-mono/700.css'
import './styles/index.css'
import App from './App'

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
