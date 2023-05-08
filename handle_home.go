package main

import (
	"app/db/sqlc"
	"app/services/spotify"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"sync"

	"github.com/newrelic/go-agent/v3/newrelic"
)

func (s *Server) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Println("getting top 500 skeleton playlists all time")
		skeletonPlaylists, err := s.qry.GetTop500PlaylistsByUpvotesAllTime(req.Context())
		if err != nil {
			log.Printf("failed to query for top 500 playlists: %v", err)
			s.views.RenderError(w, "")
			return
		}

		playlistIdsToFetch := len(skeletonPlaylists)
		log.Println("skeleton playlist ids returned", playlistIdsToFetch)

		var playlists []spotify.Playlist
		var mtx sync.Mutex
		var wg sync.WaitGroup

		wg.Add(playlistIdsToFetch)
		errors := make(chan error, playlistIdsToFetch)

		for _, skeletonPlaylist := range skeletonPlaylists {
			go func(skeletonPlaylist sqlc.Playlist) {
				defer wg.Done()

				playlist, err := s.spotify.PlaylistMetadata(req.Context(), skeletonPlaylist.ID)
				if err != nil {
					err := fmt.Errorf("fetching playlist %s from spotify: %w", skeletonPlaylist.ID, err)
					errors <- err
					return
				}
				playlist.Upvotes = skeletonPlaylist.Upvotes

				playlist.ColorsCommonFour, err = playlist.ProminentFourCoverColors()
				if err != nil {
					err := fmt.Errorf("fetching playlist %s prominent colors: %w", playlist.ID, err)
					errors <- err
					return
				}

				playlist.ArtistsCommonFour = playlist.MostCommonFourArtists()

				mtx.Lock()
				defer mtx.Unlock()
				playlists = append(playlists, *playlist)
			}(skeletonPlaylist)
		}

		wg.Wait()
		close(errors)

		for err := range errors {
			log.Printf("error fetching playlist for home page: %v", err)
			noticeError(req, err)
		}

		sort.Slice(playlists, func(i, j int) bool {
			return playlists[i].Upvotes > playlists[j].Upvotes
		})

		if len(playlists) > 26 {
			for i := 0; i < 5; i++ {
				playlists[i].EagerLoadImage = true
			}
		} else if len(playlists) > 5 {
			for i := 0; i < 5; i++ {
				playlists[i].EagerLoadImage = true
			}
		}

		w.Header().Set("Cache-Control", "public, max-age=5")
		s.views.Render(w, "index.tmpl", map[string]any{
			"new_relic_head": template.HTML(newrelic.FromContext(req.Context()).BrowserTimingHeader().WithTags()),
			"playlists":      playlists,
		})
	}
}
