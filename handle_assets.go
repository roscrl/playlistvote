package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"app/config"
)

//go:embed views/assets/dist
var assetsFS embed.FS

func (s *Server) handleAssets() http.HandlerFunc {
	if s.env == config.PROD {
		subFS, err := fs.Sub(assetsFS, "views/assets/dist")
		if err != nil {
			log.Fatal(err)
		}

		return http.FileServer(http.FS(subFS)).ServeHTTP
	} else {
		return http.FileServer(http.Dir("./views/assets/dist/")).ServeHTTP
	}
}
