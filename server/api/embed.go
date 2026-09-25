package api

import (
	"embed"
	"net/http"
)

//go:embed web
var webFS embed.FS

func ServeWeb(w http.ResponseWriter, r *http.Request) {
	http.FileServer(http.FS(webFS)).ServeHTTP(w, r)
}
