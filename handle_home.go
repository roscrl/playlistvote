package main

import (
	"app/db/sqlc"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"

	"app/services/spotify"

	"github.com/newrelic/go-agent/v3/newrelic"
)

func (s *Server) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Println("getting top 500 skeleton playlists all time")
		skeletonPlaylists, err := s.qry.GetTop500PlaylistsByUpvotesAllTime(req.Context())
		if err != nil {
			log.Printf("failed to query for top 500 playlists: %v", err)
			s.views.Render(w, "error.tmpl", map[string]any{})
			return
		}

		playlistIdsToFetch := len(skeletonPlaylists)
		log.Println("skeleton playlist ids returned", playlistIdsToFetch)

		var playlists []spotify.Playlist
		var mtx sync.Mutex
		var wg sync.WaitGroup

		wg.Add(playlistIdsToFetch)
		errorChannel := make(chan error, playlistIdsToFetch)

		for _, skeletonPlaylist := range skeletonPlaylists {
			go func(skeletonPlaylist sqlc.Playlist) {
				defer wg.Done()

				playlist, err := s.spotify.PlaylistMetadata(req.Context(), skeletonPlaylist.ID)
				if err != nil {
					err := fmt.Errorf("fetching playlist %s from spotify: %w", skeletonPlaylist.ID, err)
					errorChannel <- err
					return
				}
				playlist.Upvotes = skeletonPlaylist.Upvotes

				if playlist.ColorsCommonFour == nil {
					playlist.ColorsCommonFour, err = playlist.ProminentFourCoverColors()
					if err != nil {
						err := fmt.Errorf("fetching playlist %s prominent colors: %w", playlist.ID, err)
						errorChannel <- err
						return
					}
				}

				if playlist.ArtistsCommonFour == nil {
					playlist.ArtistsCommonFour = playlist.MostCommonFourArtists()
				}

				mtx.Lock()
				defer mtx.Unlock()
				playlists = append(playlists, *playlist)
			}(skeletonPlaylist)
		}

		wg.Wait()
		close(errorChannel)

		for err := range errorChannel {
			log.Printf("error fetching playlist for home page: %v", err)
			noticeError(req, err)
		}

		w.Header().Set("Cache-Control", "public, max-age=300")
		s.views.Render(w, "index.tmpl", map[string]any{
			"new_relic_head": template.HTML(newrelic.FromContext(req.Context()).BrowserTimingHeader().WithTags()),
			"playlists":      playlists,
		})
	}
}
