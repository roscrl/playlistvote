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

const (
	assetsPath = "views/assets/dist"
)

func (s *Server) handleAssets() http.HandlerFunc {
	if s.cfg.Env == config.PROD {
		subFS, err := fs.Sub(assetsFS, assetsPath)
		if err != nil {
			log.Fatal(err)
		}

		return http.FileServer(http.FS(subFS)).ServeHTTP
	}

	return http.FileServer(http.Dir("./" + assetsPath + "/")).ServeHTTP
}
