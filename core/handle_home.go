package core

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"app/core/db/sqlc"
	"app/core/domain"
	"app/core/spotify"
	"app/core/views"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type SkeletonPlaylist struct {
	ID      string
	Upvotes int64
}

func (s *Server) handleHome() http.HandlerFunc {
	const playlistFetchLimit = 30

	return func(w http.ResponseWriter, req *http.Request) {
		log.Printf("getting top %v skeleton playlists", playlistFetchLimit)

		topSkeletonPlaylists, err := s.Qry.GetTopPlaylists(req.Context(), playlistFetchLimit)
		if err != nil {
			log.Printf("failed to query for top %v playlists: %v", playlistFetchLimit, err)
			s.Views.RenderStandardError(w)

			return
		}

		var skeletonPlaylists []SkeletonPlaylist
		for _, playlist := range topSkeletonPlaylists {
			skeletonPlaylists = append(skeletonPlaylists, SkeletonPlaylist{
				ID:      playlist.ID,
				Upvotes: playlist.Upvotes,
			})
		}

		playlists := fetchPlaylistsFromSkeletonPlaylists(req.Context(), s.Client, skeletonPlaylists, s.Spotify)

		w.Header().Set("Cache-Control", "public, max-age=5")
		s.Views.Render(w, "index.tmpl", map[string]any{
			"new_relic_head": template.HTML(newrelic.FromContext(req.Context()).BrowserTimingHeader().WithTags()), //nolint: gosec
			"playlists":      playlists,
		})
	}
}

func (s *Server) handlePlaylistsPaginationTop() http.HandlerFunc {
	const fetchLimit = 12

	return func(w http.ResponseWriter, req *http.Request) {
		if !views.TurboStreamRequest(req) {
			http.Redirect(w, req, RouteHome, http.StatusSeeOther)

			return
		}

		// query is in form of `top?after=playlist_id-upvotes`
		after := req.URL.Query().Get("after")
		if after == "" {
			log.Printf("invalid after query param: %s", after)
			s.Views.RenderStandardError(w)

			return
		}

		playlistIDAndUpvoteCount := strings.Split(after, "-")
		if len(playlistIDAndUpvoteCount) != 2 { //nolint: gomnd
			log.Printf("invalid after query param: %s", after)
			s.Views.RenderStandardError(w)

			return
		}

		playlistID := playlistIDAndUpvoteCount[0]

		upvotes, err := strconv.ParseInt(playlistIDAndUpvoteCount[1], 10, 64)
		if err != nil {
			log.Printf("invalid after query param: %s", after)
			s.Views.RenderStandardError(w)

			return
		}

		log.Printf("fetching next top playlists after given playlist id: %s, upvotes: %d", playlistID, upvotes)

		nextTopSkeletonPlaylists, err := s.Qry.NextTopPlaylists(req.Context(), sqlc.NextTopPlaylistsParams{
			ID:      playlistID,
			Upvotes: upvotes,
			Limit:   fetchLimit,
		})
		if err != nil {
			log.Printf("failed to query for next top playlists: %v", err)
			s.Views.RenderStandardError(w)

			return
		}

		if len(nextTopSkeletonPlaylists) == 0 {
			log.Printf("no more playlists to fetch")
			w.WriteHeader(http.StatusNoContent)

			return
		}

		log.Printf("next top playlists returned: %d", len(nextTopSkeletonPlaylists))

		var skeletonPlaylists []SkeletonPlaylist
		for _, playlist := range nextTopSkeletonPlaylists {
			skeletonPlaylists = append(skeletonPlaylists, SkeletonPlaylist{
				ID:      playlist.ID,
				Upvotes: playlist.Upvotes,
			})
		}

		playlists := fetchPlaylistsFromSkeletonPlaylists(req.Context(), s.Client, skeletonPlaylists, s.Spotify)

		w.Header().Set("Cache-Control", "public, max-age=5")

		s.Views.Render(w, "playlist/_top.stream.tmpl", map[string]any{
			"playlists": playlists,
		})
	}
}

func fetchPlaylistsFromSkeletonPlaylists(ctx context.Context, client *http.Client, skeletonPlaylists []SkeletonPlaylist, spotifyClient *spotify.Client) []*domain.Playlist {
	var (
		playlists []*domain.Playlist
		mtx       sync.Mutex
		wg        sync.WaitGroup
	)

	countPlaylistIdsToFetch := len(skeletonPlaylists)

	wg.Add(countPlaylistIdsToFetch)
	errors := make(chan error, countPlaylistIdsToFetch)

	for _, skeletonPlaylist := range skeletonPlaylists {
		go func(skeletonPlaylist SkeletonPlaylist) {
			defer wg.Done()

			playlistAPIResponse, err := spotifyClient.Playlist(ctx, skeletonPlaylist.ID)
			if err != nil {
				err := fmt.Errorf("fetching playlist %s from spotify: %w", skeletonPlaylist.ID, err)
				errors <- err

				return
			}

			playlist, err := playlistAPIResponse.ToPlaylist()
			if err != nil {
				err := fmt.Errorf("transforming playlist %s to playlist: %w", skeletonPlaylist.ID, err)
				errors <- err

				return
			}

			err = playlist.AttachMetadata(ctx, client, skeletonPlaylist.Upvotes)
			if err != nil {
				err := fmt.Errorf("attaching metadata to playlist %s: %w", skeletonPlaylist.ID, err)
				errors <- err

				return
			}

			mtx.Lock()
			playlists = append(playlists, playlist)
			mtx.Unlock()
		}(skeletonPlaylist)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		// TODO handle fetch failures due to deleted playlists
		log.Printf("error fetching playlist: %v", err)
		noticeError(ctx, err)
	}

	sort.Slice(playlists, func(i, j int) bool {
		if playlists[i].Upvotes == playlists[j].Upvotes {
			return playlists[i].ID > playlists[j].ID
		}

		return playlists[i].Upvotes > playlists[j].Upvotes
	})

	return playlists
}
