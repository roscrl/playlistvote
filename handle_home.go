package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"sync"

	"app/db/sqlc"
	"app/domain/spotify"
	"app/views"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func (s *Server) handleHome() http.HandlerFunc {
	const topPlaylistsToEagerLoad = 500

	return func(w http.ResponseWriter, req *http.Request) {
		log.Println("getting top 500 skeleton playlists all time")

		skeletonPlaylists, err := s.qry.GetTopPlaylists(req.Context(), topPlaylistsToEagerLoad)
		if err != nil {
			log.Printf("failed to query for top 500 playlists: %v", err)
			s.views.RenderError(w, "")

			return
		}

		playlistIdsToFetch := len(skeletonPlaylists)
		log.Println("skeleton playlist ids returned", playlistIdsToFetch)

		var (
			playlists []spotify.Playlist
			mtx       sync.Mutex
			wg        sync.WaitGroup
		)

		wg.Add(playlistIdsToFetch)
		errors := make(chan error, playlistIdsToFetch)

		for _, skeletonPlaylist := range skeletonPlaylists {
			go func(skeletonPlaylist sqlc.GetTopPlaylistsRow) {
				defer wg.Done()

				playlist, err := s.spotify.PlaylistMetadata(req.Context(), skeletonPlaylist.ID)
				if err != nil {
					err := fmt.Errorf("fetching playlist %s from spotify: %w", skeletonPlaylist.ID, err)
					errors <- err

					return
				}

				playlist.Upvotes = skeletonPlaylist.Upvotes

				playlist.ColorsCommonFour, err = playlist.MostProminentFourCoverColors(req.Context(), s.client)
				if err != nil {
					err := fmt.Errorf("fetching playlist %s prominent colors: %w", playlist.ID, err)
					errors <- err

					return
				}

				playlist.ArtistsCommonFour = playlist.MostCommonFourArtists()

				playlist.EagerLoadImage = true

				mtx.Lock()
				playlists = append(playlists, *playlist)
				mtx.Unlock()
			}(skeletonPlaylist)
		}

		wg.Wait()
		close(errors)

		for err := range errors {
			log.Printf("error fetching playlist for home page: %v", err)
			noticeError(req, err)
		}

		sort.Slice(playlists, func(i, j int) bool {
			if playlists[i].Upvotes == playlists[j].Upvotes {
				return playlists[i].Name < playlists[j].Name
			}

			return playlists[i].Upvotes > playlists[j].Upvotes
		})

		w.Header().Set("Cache-Control", "public, max-age=5")
		s.views.Render(w, "index.tmpl", map[string]any{
			"new_relic_head": template.HTML(newrelic.FromContext(req.Context()).BrowserTimingHeader().WithTags()), //nolint: gosec
			"playlists":      playlists,
		})
	}
}

func (s *Server) handleTopPlaylistsAfterCursor() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if !views.TurboStreamRequest(req) {
			http.Redirect(w, req, HomeRoute, http.StatusSeeOther)

			return
		}

		// query is in form of `top?after=playlist_id`
		afterPlaylistID := req.URL.Query().Get("after")
		if afterPlaylistID == "" {
			log.Printf("invalid after query param: %s", afterPlaylistID)
			s.views.RenderError(w, "")

			return
		}

		const TopPlaylistsFetchLimit = 30

		nextPlaylists, err := s.qry.NextTopPlaylists(req.Context(), sqlc.NextTopPlaylistsParams{
			ID:    afterPlaylistID,
			Limit: TopPlaylistsFetchLimit,
		})
		if err != nil {
			log.Printf("failed to query for next top playlists: %v", err)
			s.views.RenderError(w, "")

			return
		}

		log.Printf("next playlists returned: %d", len(nextPlaylists))
	}
}
