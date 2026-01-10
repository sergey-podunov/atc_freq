package wails

import (
	"embed"
	"io/fs"
)

//go:embed all:frontend/dist
var assets embed.FS

var Assets fs.FS

func init() {
	var err error
	// If frontend/dist exists in this directory, assets will contain:
	// frontend/dist/index.html etc.
	// fs.Sub(assets, "frontend/dist") will return an FS where index.html is at the root.
	Assets, err = fs.Sub(assets, "frontend/dist")
	if err != nil {
		panic(err)
	}
	if Assets == nil {
		panic("fs.Sub returned nil")
	}
}
