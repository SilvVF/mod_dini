package main

import (
	"embed"

	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
	bolt "go.etcd.io/bbolt"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/build
var assets embed.FS

func init() {
	// Register a custom event whose associated data type is string.
	// This is not required, but the binding generator will pick up registered events
	// and provide a strongly typed JS/TS API for them.
	application.RegisterEvent[ModSaved]("mod-update")
	application.RegisterEvent[string]("mod-removed")
	application.RegisterEvent[map[int]ModSaved]("mods-loaded")
	application.RegisterEvent[DataProgress]("mod-download-progress")
	application.RegisterEvent[Setting]("setting-updated")
}

func main() {

	// todo(open in Local app data dir)
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := application.New(application.Options{
		Name:        "mod_dini",
		Description: "hey mr dini",
		Services:    []application.Service{},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.RegisterService(application.NewService(NewModService(app, db)))

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "mod_dini",
		// Window sized to the golden ratio (1000 / 618 ≈ 1.618).
		Width:            1000,
		Height:           618,
		BackgroundColour: application.NewRGB(6, 7, 15),
		URL:              "/",
	})

	if err = app.Run(); err != nil {
		log.Fatal(err)
	}
}
