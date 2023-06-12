package core

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
	PathAssets         = "core/views/assets/dist"
	PathEmbeddedAssets = "views/assets/dist"
)

func (s *Server) handleAssets() http.HandlerFunc {
	if s.Cfg.Env == config.PROD {
		subFS, err := fs.Sub(assetsFS, PathEmbeddedAssets)
		if err != nil {
			log.Fatal(err)
		}

		return http.FileServer(http.FS(subFS)).ServeHTTP
	}

	return http.FileServer(http.Dir("./" + PathAssets + "/")).ServeHTTP
}
