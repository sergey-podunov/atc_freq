package main

import (
	"atc_freq/internal/app"
	wailsAssets "atc_freq/wails"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

func main() {
	a := app.NewApp()
	_ = wails.Run(&options.App{
		Title:  "ATC Frequency Finder",
		Width:  600,
		Height: 500,
		AssetServer: &assetserver.Options{
			Assets: wailsAssets.Assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        a.Startup,
		Bind: []interface{}{
			a,
		},
	})
}
