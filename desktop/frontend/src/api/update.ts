// 새 앱 버전 확인. 화면은 바인딩 대신 이 파일을 쓴다.
import { UpdateService } from '../../bindings/github.com/TogetherBros/azit/desktop/internal/bridge'

// Go 의 update.Info 와 같은 모양
export type UpdateInfo = {
  current: string
  latest: string
  available: boolean // 새 버전이 있고 이 OS 용 설치 파일도 있음
  url: string
  notes: string
}

export async function checkUpdate(): Promise<UpdateInfo> {
  return (await UpdateService.Check()) as unknown as UpdateInfo
}

// 설치 파일 주소를 브라우저로 연다
export async function openUpdate(url: string): Promise<void> {
  await UpdateService.Open(url)
}
