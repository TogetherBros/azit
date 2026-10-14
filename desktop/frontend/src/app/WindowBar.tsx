import { closeWindow, minimise, toggleMaximise } from '../api/window'
import { Icon } from '../components/Icon'

// 창 맨 위 줄. 잡고 끌면 창이 움직이고, 두 번 누르면 최대화.
// OS 기본 제목줄 대신 쓴다 (main.go 의 Frameless: true).
export function WindowBar() {
  return (
    <div className="winbar" onDoubleClick={toggleMaximise}>
      <div className="winbar-btns" onDoubleClick={(e) => e.stopPropagation()}>
        <button type="button" className="winbtn" aria-label="최소화" onClick={minimise}>
          <Icon name="min" size={14} stroke={2} />
        </button>
        <button type="button" className="winbtn" aria-label="최대화" onClick={toggleMaximise}>
          <Icon name="max" size={12} stroke={2} />
        </button>
        <button type="button" className="winbtn close" aria-label="닫기" onClick={closeWindow}>
          <Icon name="x" size={14} stroke={2} />
        </button>
      </div>
    </div>
  )
}
