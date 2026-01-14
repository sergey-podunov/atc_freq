package main

import (
	"atc_freq/internal/app"
	"atc_freq/internal/sim"
	wailsAssets "atc_freq/wails"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

func main() {
	connection, err := sim.NewConnection()
	if err != nil {
		fmt.Printf("Can't create application: %v\n", err)
		os.Exit(1)
	}

	a := app.NewApp(connection)
	_ = wails.Run(&options.App{
		Title:  "ATC Frequency Finder",
		Width:  600,
		Height: 500,
		AssetServer: &assetserver.Options{
			Assets: wailsAssets.Assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        a.AddContext,
		Bind: []interface{}{
			a,
		},
	})
}
