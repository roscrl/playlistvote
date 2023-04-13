package main

import (
	"html/template"
	"net/http"

	"app/services/spotify"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func (s *Server) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		playlists := spotify.MockPlaylists() // TODO hit db

		s.views.Render(w, "index.tmpl", map[string]any{
			"new_relic_head": template.HTML(newrelic.FromContext(req.Context()).BrowserTimingHeader().WithTags()),
			"playlists":      playlists,
		})
	}
}
