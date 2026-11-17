package main

import (
	"embed"
	"fmt"
	"net"
	"net/http"
	"path/filepath"

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

	previewURL, err := startPreviewServer(`C:\Users\david\SkinMods\preview`)
	if err != nil {
		panic(err)
	}

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

	app.RegisterService(application.NewService(NewModService(app, db, previewURL)))

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

func startPreviewServer(previewDir string) (string, error) {
	absDir, err := filepath.Abs(previewDir)
	if err != nil {
		return "", err
	}

	listener, err := net.Listen("tcp", "127.0.0.1:1302")
	if err != nil {
		return "", err
	}

	mux := http.NewServeMux()
	mux.Handle("/previews/", http.StripPrefix(
		"/previews/",
		http.FileServer(http.Dir(absDir)),
	))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://wails.localhost:9245")
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		mux.ServeHTTP(w, r)
	})

	server := &http.Server{
		Handler: handler,
	}

	go func() {
		_ = server.Serve(listener)
	}()

	return fmt.Sprintf("http://%s/previews/", listener.Addr().String()), nil
}
