package main

import (
	"app/services/spotify"
	"html/template"
	"log"
	"net/http"

	"github.com/newrelic/go-agent/v3/newrelic"
)

func (s *Server) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		var playlists []spotify.Playlist
		if s.mocking {
			playlists = spotify.MockPlaylists()
		} else {
			log.Println("getting top 500 playlists all time")
			playlistIds, err := s.qry.GetTop500PlaylistsByUpvotesAllTime(req.Context())
			if err != nil {
				s.views.Render(w, "error.tmpl", map[string]any{})
				return
			}

			log.Println("playlist ids returned", len(playlistIds))

			for _, playlistId := range playlistIds {
				playlist, err := s.spotify.PlaylistMetadata(req.Context(), playlistId.ID)
				if err != nil {
					log.Printf("failed to get playlist %s: %v", playlistId, err)
					s.views.Render(w, "error.tmpl", map[string]any{})
					return
				}
				playlists = append(playlists, *playlist)
			}
		}

		w.Header().Set("Cache-Control", "public, max-age=300")
		s.views.Render(w, "index.tmpl", map[string]any{
			"new_relic_head": template.HTML(newrelic.FromContext(req.Context()).BrowserTimingHeader().WithTags()),
			"playlists":      playlists,
		})
	}
}
