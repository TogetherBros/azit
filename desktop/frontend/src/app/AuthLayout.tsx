import type { ReactNode } from 'react'

type Props = {
  blobs?: 'both' | 'lime' | 'none'
  children: ReactNode
}

// 들어오기 화면들의 바탕: 가운데 정렬 + 흐릿한 원
export function AuthLayout({ blobs = 'both', children }: Props) {
  return (
    <div className="authwrap">
      {blobs !== 'none' && <span className="blob lime" />}
      {blobs === 'both' && <span className="blob pink" />}
      {children}
    </div>
  )
}
