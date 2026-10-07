import { AuthLayout } from '../../app/AuthLayout'
import logoBig from '../../assets/logo-big.webp'

type Props = {
  waiting?: boolean // 기본 브라우저에서 구글 로그인 중
  error?: string // "멈춘 계정이에요 ..." 등
}

// 첫 화면. 구글 로그인은 앱 안이 아니라 기본 브라우저에서 한다 (desktop/internal/login).
export function LoginScreen({ waiting, error }: Props) {
  return (
    <AuthLayout>
      <div className="auth-body">
        <img src={logoBig} alt="아지트" className="auth-logo" />
        <p className="auth-lead">
          친구들끼리만 쓰는 아지트예요.
          <br />
          구글 계정으로 들어와요.
        </p>

        {waiting ? (
          <div className="card auth-wait" role="status">
            <span className="auth-wait-dot" />
            <span className="b">브라우저에서 로그인을 마쳐 주세요</span>
            <span className="hint">끝나면 여기로 알아서 돌아와요</span>
            <button type="button" className="btn s sec">취소</button>
          </div>
        ) : (
          <button type="button" className="btn l p">구글로 들어가기</button>
        )}

        {error && <div role="alert" className="ferr auth-center">{error}</div>}

        <p className="hint auth-foot">처음 들어오면 친구에게 받은 초대 코드가 필요해요.</p>
      </div>
    </AuthLayout>
  )
}
