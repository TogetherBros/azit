# 아지트 (azit)

초대받은 친구들만 쓰는 데스크톱 메신저. 아지트(모임) 안에 방을 만들어 떠들고, 짤·맞장구·흔들기로 논다.

## 받기

| 운영체제 | 설치 파일 |
| --- | --- |
| Windows 10 / 11 | [azit-windows-setup.exe](https://api.togetherbros.uk/download/azit-windows-setup.exe) |
| macOS (Intel · Apple Silicon) | [azit-mac.zip](https://api.togetherbros.uk/download/azit-mac.zip) |
| Linux | [azit-linux.AppImage](https://api.togetherbros.uk/download/azit-linux.AppImage) |

위 링크가 안 열리면 [GitHub 최신 릴리스](https://github.com/TogetherBros/azit/releases/latest)에서 같은 파일을 받아요.

한 번 설치하면 새 버전이 나올 때 앱이 알려 줘요. 다시 찾아올 필요 없어요.

### 설치

**Windows**
1. `azit-windows-setup.exe` 를 실행해요. 관리자 권한 없이 내 계정에만 설치돼요.
2. "Windows의 PC 보호" 파란 창이 뜨면 **추가 정보 → 실행** 을 눌러요. 코드 서명을 아직 안 해서 뜨는 경고예요.

**macOS**
1. `azit-mac.zip` 을 풀고 `azit.app` 을 **응용 프로그램** 폴더로 옮겨요.
2. 처음 한 번은 앱을 **우클릭 → 열기 → 열기** 로 켜요. 그냥 더블클릭하면 "확인되지 않은 개발자" 경고로 막혀요.
   - 그래도 안 열리면 터미널에서 `xattr -dr com.apple.quarantine /Applications/azit.app`

**Linux**
```bash
chmod +x azit-linux.AppImage
./azit-linux.AppImage
```
켜지지 않으면 `libfuse2` (Ubuntu 24.04 는 `libfuse2t64`) 를 깔아 주세요.

### 처음 쓰기

1. **구글로 들어가기** 로 로그인해요.
2. 친구한테 받은 초대 코드(`XXXX-XXXX`)를 넣으면 그 아지트에 들어가요. 직접 아지트를 만들 수도 있어요.
3. 앱 설정 → 앱 에서 **컴퓨터 켜면 같이 켜기**, **창 닫으면 숨기기**(트레이)를 켤 수 있어요. 자동 실행은 지금 Windows 만 돼요.

---

## 개발

서버와 앱 모두 Go. 서버는 Fiber + SQLite(GORM), 앱은 Wails v3 + React(TypeScript).

```
server/    API 서버 (REST + 웹소켓 /v1/ws)
desktop/   데스크톱 앱 (Go: desktop/, 화면: desktop/frontend/)
proto/     서버와 앱이 같이 쓰는 요청·응답 모양
deploy/    운영 docker-compose, Caddy
docs/      api-spec.html (API 명세, 브라우저로 열기)
```

### 서버 켜기

```bash
cd server
cp .env.example .env   # JWT_SECRET, 구글 클라이언트 값 채우기. 로컬에선 DEV=1
go run .
```
기본 주소는 `http://127.0.0.1:8081`. 테이블은 켤 때 자동으로 만들어져요.

### 앱 켜기

```bash
cd desktop
wails3 dev
```
화면만 볼 때는 `cd desktop/frontend && npm run dev` (서버 연결 없이 그리기만).

### 새 버전 내기

`v` 로 시작하는 태그를 올리면 GitHub Actions 가 Windows·macOS·Linux 로 빌드해서 위 다운로드 주소에 올려요 (`.github/workflows/release-desktop.yml`).

```bash
git tag -a v0.2.0 -m "고친 것: ..."   # -m 내용이 앱의 업데이트 알림에 그대로 보여요
git push origin v0.2.0
```

서버는 `server/`, `proto/`, `deploy/` 가 바뀐 채로 `main` 에 push 하면 자동 배포돼요 (`.github/workflows/deploy-server.yml`).
