package main

import (
	"embed"

	"log"
	"time"

	"github.com/TogetherBros/azit/desktop/internal/bridge"
	"github.com/TogetherBros/azit/desktop/internal/client"
	"github.com/wailsapp/wails/v3/pkg/application"
)

var assets embed.FS

const (
	serverURL      = "http://127.0.0.1:8081"
	googleClientID = "1089773756666-p61lrvmkpkkrfb5ofpu3pglv7fb8607n.apps.googleusercontent.com"
)

func init() {
	application.RegisterEvent[string]("time")
}

func main() {
	api := client.New(serverURL)
	auth := bridge.NewAuthService(api, googleClientID)

	app := application.New(application.Options{
		Name:        "azit",
		Description: "A demo of using raw HTML & CSS",
		Services: []application.Service{
			application.NewService(&GreetService{}),
			application.NewService(auth),
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

	err := app.Run()

	if err != nil {
		log.Fatal(err)
	}
}
