package core

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"app/core/db/sqlc"
	"app/core/domain"
	"app/core/rlog"
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

	return func(w http.ResponseWriter, r *http.Request) {
		log := rlog.L(r.Context())

		log.InfoCtx(r.Context(), "fetching top skeleton playlists", "amount", playlistFetchLimit)

		topSkeletonPlaylists, err := s.Qry.GetTopPlaylists(r.Context(), playlistFetchLimit)
		if err != nil {
			log.ErrorCtx(r.Context(), "failed to query for top playlists", "amount", playlistFetchLimit, "err", err)
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

		playlists := fetchPlaylistsFromSkeletonPlaylists(r.Context(), s.Client, skeletonPlaylists, s.Spotify)

		w.Header().Set("Cache-Control", "public, max-age=5")
		s.Views.Render(w, "index.tmpl", map[string]any{
			"new_relic_head": template.HTML(newrelic.FromContext(r.Context()).BrowserTimingHeader().WithTags()), //nolint: gosec
			"playlists":      playlists,
		})
	}
}

func (s *Server) handlePlaylistsPaginationTop() http.HandlerFunc {
	const playlistFetchLimit = 12

	return func(w http.ResponseWriter, r *http.Request) {
		log := rlog.L(r.Context())

		if !views.TurboStreamRequest(r) {
			http.Redirect(w, r, RouteHome, http.StatusSeeOther)

			return
		}

		// query is in form of `top?after=playlist_id-upvotes`
		after := r.URL.Query().Get("after")
		if after == "" {
			log.InfoCtx(r.Context(), "missing after query param")
			s.Views.RenderStandardError(w)

			return
		}

		playlistIDAndUpvoteCount := strings.Split(after, "-")
		if len(playlistIDAndUpvoteCount) != 2 { //nolint: gomnd
			log.InfoCtx(r.Context(), "missing after query param", "after", after)
			s.Views.RenderStandardError(w)

			return
		}

		playlistID := playlistIDAndUpvoteCount[0]

		upvotes, err := strconv.ParseInt(playlistIDAndUpvoteCount[1], 10, 64)
		if err != nil {
			log.InfoCtx(r.Context(), "invalid after query param", "after", after, "playlist_id", playlistID, "err", err)
			s.Views.RenderStandardError(w)

			return
		}

		log.InfoCtx(r.Context(), "fetching top playlists after given playlist id", "playlist_id", playlistID, "upvotes", upvotes)

		nextTopSkeletonPlaylists, err := s.Qry.NextTopPlaylists(r.Context(), sqlc.NextTopPlaylistsParams{
			ID:      playlistID,
			Upvotes: upvotes,
			Limit:   playlistFetchLimit,
		})
		if err != nil {
			log.InfoCtx(r.Context(), "failed to query for next top playlists", "playlist_id", playlistID, "upvotes", upvotes, "err", err)
			s.Views.RenderStandardError(w)

			return
		}

		if len(nextTopSkeletonPlaylists) == 0 {
			log.InfoCtx(r.Context(), "no more playlists to fetch", "playlist_id", playlistID, "upvotes", upvotes)
			w.WriteHeader(http.StatusNoContent)

			return
		}

		log.InfoCtx(r.Context(), "next top playlists returned", "amount", len(nextTopSkeletonPlaylists))

		var skeletonPlaylists []SkeletonPlaylist
		for _, playlist := range nextTopSkeletonPlaylists {
			skeletonPlaylists = append(skeletonPlaylists, SkeletonPlaylist{
				ID:      playlist.ID,
				Upvotes: playlist.Upvotes,
			})
		}

		playlists := fetchPlaylistsFromSkeletonPlaylists(r.Context(), s.Client, skeletonPlaylists, s.Spotify)

		w.Header().Set("Cache-Control", "public, max-age=5")

		s.Views.Render(w, "playlist/_top.stream.tmpl", map[string]any{
			"playlists": playlists,
		})
	}
}

func fetchPlaylistsFromSkeletonPlaylists(ctx context.Context, client *http.Client, skeletonPlaylists []SkeletonPlaylist, spotifyClient *spotify.Client) []*domain.Playlist {
	log := rlog.L(ctx)

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
		log.ErrorCtx(ctx, "failed to fetch playlist", "err", err)
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
