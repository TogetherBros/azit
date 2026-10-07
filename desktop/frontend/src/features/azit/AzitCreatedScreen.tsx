import { AuthLayout } from '../../app/AuthLayout'
import { Mark } from '../../components/Mark'

type Props = { name: string; short: string; color: string; code: string }

// 아지트를 만든 직후: 첫 초대 코드를 크게 보여 준다.
export function AzitCreatedScreen({ name, short, color, code }: Props) {
  return (
    <AuthLayout blobs="none">
      <div className="card created-card">
        <Mark label={short} color={color} size="xl" />
        <h1 className="created-title">{name} 완성</h1>
        <p className="mu">아직 나 혼자예요. 이 코드를 친구들한테 보내요.</p>
        <div className="code-box">
          <span className="code-big">{code}</span>
          <span className="hint">7일 동안 · 횟수 제한 없음 · 설정에서 바꿀 수 있어요</span>
          <button type="button" className="btn p">코드 복사</button>
        </div>
        <button type="button" className="btn l sec" style={{ width: '100%' }}>아지트로 가기</button>
      </div>
    </AuthLayout>
  )
}
