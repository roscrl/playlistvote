package main

import (
	"html/template"
	"log"
	"net/http"
	"sync"

	"app/services/spotify"

	"github.com/newrelic/go-agent/v3/newrelic"
)

func (s *Server) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var playlists []spotify.Playlist
		var mtx sync.Mutex
		var wg sync.WaitGroup
		// REDO

		if s.cfg.Mocking {
			playlists = spotify.MockPlaylists("services/spotify/mock_playlists.json")
			s.views.Render(w, "index.tmpl", map[string]any{
				"new_relic_head": template.HTML(newrelic.FromContext(req.Context()).BrowserTimingHeader().WithTags()),
				"playlists":      playlists,
			})
			return
		}

		log.Println("getting top 500 playlists all time")
		playlistInDb, err := s.qry.GetTop500PlaylistsByUpvotesAllTime(req.Context())
		if err != nil {
			log.Printf("failed to query for top 500 playlists: %v", err)
			s.views.Render(w, "error.tmpl", map[string]any{})
			return
		}

		log.Println("playlist ids returned", len(playlistInDb))
		wg.Add(len(playlistInDb))

		for _, playlist := range playlistInDb {
			go func(playlistID string) {
				defer wg.Done()
				playlist, err := s.spotify.PlaylistMetadata(req.Context(), playlistID)
				if err != nil {
					log.Printf("failed to get playlist %s: %v", playlistID, err)
					s.views.Render(w, "error.tmpl", map[string]any{})
					return
				}
				mtx.Lock()
				defer mtx.Unlock()
				playlists = append(playlists, *playlist)
			}(playlist.ID)
		}

		wg.Wait()

		// TODO this is dumb fix n+1 query, maybe dto time
		for i := range playlists {
			upvotes, err := s.qry.GetPlaylistUpvotes(req.Context(), playlists[i].ID)
			if err != nil {
				log.Printf("failed to query upvotes for playlist %s: %v", playlists[i].ID, err)
				s.views.Render(w, "error.tmpl", map[string]any{
					"error": "failed to query upvotes for playlist " + playlists[i].ID,
				})
				return
			}
			playlists[i].Upvotes = upvotes
		}

		w.Header().Set("Cache-Control", "public, max-age=300")
		s.views.Render(w, "index.tmpl", map[string]any{
			"new_relic_head": template.HTML(newrelic.FromContext(req.Context()).BrowserTimingHeader().WithTags()),
			"playlists":      playlists,
		})
	}
}
