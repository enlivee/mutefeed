package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/*
var staticFS embed.FS

// Handler returns the http.Handler serving embedded static files
func Handler() http.Handler {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(sub))
}

// GetIndexHTML returns index.html raw bytes
func GetIndexHTML() ([]byte, error) {
	return staticFS.ReadFile("static/index.html")
}
