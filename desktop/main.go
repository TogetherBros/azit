package main

import (
	"context"
	"embed"
	"os"
	"slices"

	"log"
	"time"

	"github.com/TogetherBros/azit/desktop/internal/bridge"
	"github.com/TogetherBros/azit/desktop/internal/client"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// 화면 파일(React 빌드 결과)을 exe 안에 넣는다. 배포용 exe 는 이걸로 화면을 띄운다
// (wails3 dev 때는 vite 개발 서버에서 받아서 상관없다). frontend/dist/.gitkeep 이 있어서
// 화면을 아직 안 빌드했어도 go build 는 된다 (그땐 빈 창).
//
//go:embed all:frontend/dist
var assets embed.FS

// 트레이 아이콘 (작업 표시줄 오른쪽 아래)
//
//go:embed build/windows/icon.ico
var trayIcon []byte

const (
	googleClientID = "1089773756666-p61lrvmkpkkrfb5ofpu3pglv7fb8607n.apps.googleusercontent.com"
)

func init() {
	application.RegisterEvent[string]("time")
	application.RegisterEvent[string](bridge.EventRealtime) // 서버 이벤트 JSON 그대로
	application.RegisterEvent[string](bridge.EventRTOpen)   // 실시간 연결됨 (data 없음)
}

func main() {
	api := client.New(serverURL)
	auth := bridge.NewAuthService(api, googleClientID)
	profile := bridge.NewProfileService(api, auth)
	apiService := bridge.NewAPIService(api, auth)
	updates := bridge.NewUpdateService(appVersion, updateURL) // 새 버전 알림 (serverURL, updateURL 은 server_dev.go / server_prod.go)
	realtime := bridge.NewRealtime(api, auth)                 // 서버 실시간 연결 (React 는 rt 이벤트만 듣는다)
	appService := bridge.NewAppService()                      // 자동 실행, 창 닫으면 트레이로

	app := application.New(application.Options{
		Name:        "azit",
		Description: "A demo of using raw HTML & CSS",
		Services: []application.Service{
			application.NewService(&GreetService{}),
			application.NewService(auth),
			application.NewService(profile),
			application.NewService(apiService),
			application.NewService(updates),
			application.NewService(appService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// 컴퓨터 켜며 자동 실행된 거면 창 없이 트레이에만 둔다
	startHidden := slices.Contains(os.Args[1:], bridge.AutostartFlag)

	win := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "아지트",
		Width:            1000,
		Height:           618,
		MinWidth:         400,
		MinHeight:        500,
		Frameless:        true,
		BackgroundColour: application.NewRGB(14, 14, 18),
		URL:              "/",
		Hidden:           startHidden,
	})

	// 창 닫기(X): "창 닫으면 트레이로" 가 켜져 있으면 끄지 않고 숨긴다
	win.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		if bridge.HideOnClose(appService) {
			e.Cancel()
			win.Hide()
		}
	})

	// 트레이: 누르면 창 보이기, 메뉴로 열기·끝내기
	show := func() { win.Show(); win.Focus() }
	trayMenu := application.NewMenu()
	trayMenu.Add("아지트 열기").OnClick(func(*application.Context) { show() })
	trayMenu.AddSeparator()
	trayMenu.Add("끝내기").OnClick(func(*application.Context) {
		bridge.MarkQuitting(appService)
		app.Quit()
	})
	tray := app.SystemTray.New()
	tray.SetIcon(trayIcon)
	tray.SetTooltip("아지트")
	tray.SetMenu(trayMenu)
	tray.OnClick(show)

	go func() {
		for {
			now := time.Now().Format(time.RFC1123)
			app.Event.Emit("time", now)
			time.Sleep(time.Second)
		}
	}()

	// 앱이 켜져 있는 내내 실시간 연결을 유지한다 (로그인 전이면 기다린다). 앱이 끝나면 같이 끝난다.
	rtCtx, stopRT := context.WithCancel(context.Background())
	defer stopRT()
	go realtime.Run(rtCtx)

	err := app.Run()

	if err != nil {
		log.Fatal(err)
	}
}
