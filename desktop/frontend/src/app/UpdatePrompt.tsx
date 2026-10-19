import { useEffect, useState } from 'react'
import { checkUpdate, openUpdate, type UpdateInfo } from '../api/update'
import { Modal } from '../components/Modal'

const RECHECK_MS = 6 * 60 * 60 * 1000 // 켜 둔 채로 오래 써도 6시간마다 다시 확인

// 새 버전이 나오면 "업데이트하세요" 창을 띄운다.
// 앱을 켤 때 한 번, 그 뒤로 6시간마다 확인한다. "나중에"를 누른 버전은 이번 실행 동안 다시 묻지 않는다.
// 개발 빌드(wails3 dev)에서는 Go 쪽이 확인을 안 해서 절대 안 뜬다.
export function UpdatePrompt() {
  const [info, setInfo] = useState<UpdateInfo | null>(null)
  const [skipped, setSkipped] = useState<string>()

  useEffect(() => {
    const check = () => {
      checkUpdate()
        .then((u) => setInfo(u.available ? u : null))
        .catch(() => {}) // 서버가 잠깐 안 돼도 앱 쓰는 데는 지장 없게 조용히 넘긴다
    }
    check()
    const t = window.setInterval(check, RECHECK_MS)
    return () => window.clearInterval(t)
  }, [])

  if (!info || info.latest === skipped) return null

  const later = () => setSkipped(info.latest)
  const get = () => {
    void openUpdate(info.url).catch(() => {})
    later()
  }

  return (
    <Modal title="새 버전이 나왔어요" onClose={later}>
      <p className="modal-text">
        지금 {info.current} → 새 버전 {info.latest}
        <br />
        받아서 설치하면 그대로 이어서 쓸 수 있어요.
      </p>
      {info.notes && <p className="modal-text">{info.notes}</p>}
      <div className="modal-actions">
        <button type="button" className="btn sec" onClick={later}>나중에</button>
        <button type="button" className="btn p" onClick={get}>받기</button>
      </div>
    </Modal>
  )
}
