// 앱 설정 중 운영체제에 닿는 것 (Go AppService): 컴퓨터 켜면 자동 실행, 창 닫으면 트레이로.
// 브라우저로 화면만 볼 때(vite)는 Wails 가 없어서 실패한다. 그땐 "지원 안 함"으로 본다.
import { AppService } from '../../bindings/github.com/TogetherBros/azit/desktop/internal/bridge'

// 자동 실행: 진짜 값은 레지스트리에 있어서 읽어 온다. 못 켜는 환경이면 supported=false
export async function autostartState(): Promise<{ supported: boolean; on: boolean }> {
  try {
    const supported = await AppService.AutostartSupported()
    return { supported, on: supported ? await AppService.Autostart() : false }
  } catch {
    return { supported: false, on: false }
  }
}

export const setAutostart = (on: boolean) => AppService.SetAutostart(on)

export const setCloseToTray = (on: boolean) => void AppService.SetCloseToTray(on).catch(() => {})
