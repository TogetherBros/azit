package main

import (
	"context"
	"embed"

	"log"
	"time"

	"github.com/TogetherBros/azit/desktop/internal/bridge"
	"github.com/TogetherBros/azit/desktop/internal/client"
	"github.com/wailsapp/wails/v3/pkg/application"
)

var assets embed.FS

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
	realtime := bridge.NewRealtime(api, auth) // 서버 실시간 연결 (React 는 rt 이벤트만 듣는다)

	app := application.New(application.Options{
		Name:        "azit",
		Description: "A demo of using raw HTML & CSS",
		Services: []application.Service{
			application.NewService(&GreetService{}),
			application.NewService(auth),
			application.NewService(profile),
			application.NewService(apiService),
			application.NewService(updates),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "아지트",
		Width:            1000,
		Height:           618,
		MinWidth:         400,
		MinHeight:        500,
		Frameless:        true,
		BackgroundColour: application.NewRGB(14, 14, 18),
		URL:              "/",
	})

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
