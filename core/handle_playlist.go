package core

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"app/core/db/sqlc"
	"app/core/rlog"
	"app/core/spotify"
	"app/core/views"
)

func (s *Server) handlePlaylistView() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := rlog.L(r.Context())

		playlistURI := getField(r, 0)
		playlistID := strings.TrimPrefix(playlistURI, spotify.URIPlaylistPrefix)

		{
			playlistExists, err := s.Qry.PlaylistExists(r.Context(), playlistID)
			if err != nil {
				log.InfoCtx(r.Context(), "failed to check if playlist exists", "playlist_id", playlistID, "err", err)
				s.Views.RenderStandardError(w)

				return
			}

			if playlistExists == 0 {
				log.InfoCtx(r.Context(), "playlist does not exist", "playlist_id", playlistID)
				s.Views.RenderError(w, "playlist does not exist on our side. add it on the home page", http.StatusNotFound)

				return
			}
		}

		playlistAPIResponse, err := s.Spotify.Playlist(r.Context(), playlistID)
		if err != nil {
			log.InfoCtx(r.Context(), "failed to get playlist", "playlist_id", playlistID, "err", err)
			s.Views.RenderStandardError(w)

			return
		}

		upvotes, err := s.Qry.GetPlaylistUpvotes(r.Context(), playlistID)
		if err != nil {
			log.InfoCtx(r.Context(), "failed to query for playlist upvotes", "playlist_id", playlistID, "err", err)
			s.Views.RenderStandardError(w)

			return
		}

		playlist, err := playlistAPIResponse.ToPlaylist()
		if err != nil {
			log.InfoCtx(r.Context(), "failed to transform playlist api response to playlist", "playlist_id", playlistID, "err", err)
			s.Views.RenderStandardError(w)

			return
		}

		err = playlist.AttachMetadata(r.Context(), s.Client, upvotes, time.Time{})
		if err != nil {
			log.InfoCtx(r.Context(), "failed to attach metadata to playlist", "playlist_id", playlistID, "err", err)
			s.Views.RenderStandardError(w)

			return
		}

		w.Header().Set("Cache-Control", "public, max-age=3600")
		s.Views.Render(w, "playlist/view.tmpl", map[string]any{
			"playlist": playlist,
		})
	}
}

func (s *Server) handlePlaylistCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := rlog.L(r.Context())

		if !views.TurboStreamRequest(r) {
			http.Redirect(w, r, RouteHome, http.StatusSeeOther)

			return
		}

		if err := r.ParseForm(); err != nil {
			log.InfoCtx(r.Context(), "failed to parse form", "err", err)
			s.Views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
				"error": "Failed to parse form",
			})

			return
		}

		playlistLinkOrID := r.Form.Get("playlist_link_or_id")

		log.InfoCtx(r.Context(), "given playlist link or id", "playlist_link_or_id", playlistLinkOrID)

		if strings.HasPrefix(playlistLinkOrID, spotify.AlbumCopiedPrefix) {
			s.Views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
				"playlist_input": playlistLinkOrID,
				"error":          "Looks like you copied a album link, try again with a playlist link",
			})

			return
		}

		playlistID := strings.TrimPrefix(playlistLinkOrID, spotify.UserCopiedPlaylistPrefix)
		playlistID = strings.Split(playlistID, "?")[0]
		log.InfoCtx(r.Context(), "parsed playlist id", "playlist_id", playlistID)

		{
			playlistExists, err := s.Qry.PlaylistExists(r.Context(), playlistID)
			if err != nil {
				log.InfoCtx(r.Context(), "failed to check if playlist exists", "playlist_id", playlistID, "err", err)
				s.Views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
					"playlist_id":    playlistID,
					"playlist_input": playlistLinkOrID,
					"error":          "Oops, something went wrong on checking if playlist already exists",
				})

				return
			}

			if playlistExists == 1 {
				log.InfoCtx(r.Context(), "playlist already exists", "playlist_id", playlistID)
				http.Redirect(w, r, RoutePlaylistBase+"/"+playlistID, http.StatusSeeOther)

				return
			}
		}

		log.InfoCtx(r.Context(), "playlist does not exist, fetching from spotify", "playlist_id", playlistID)

		seg := startSegment(r, "SpotifyPlaylistGet")

		playlistAPIResponse, err := s.Spotify.Playlist(r.Context(), playlistID)
		if err != nil {
			if errors.Is(err, spotify.ErrPlaylistNotFound) {
				log.InfoCtx(r.Context(), "playlist not found", "playlist_id", playlistID)
				s.Views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
					"playlist_id":    playlistID,
					"playlist_input": playlistLinkOrID,
					"error":          playlistID + " not found in Spotify, double check the link and try again",
				})
			} else if errors.Is(err, spotify.ErrTooManyRequests) {
				log.InfoCtx(r.Context(), "too many requests for spotify", "playlist_id", playlistID)
				s.Views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
					"playlist_id":    playlistID,
					"playlist_input": playlistLinkOrID,
					"error":          "Too many requests for the Spotify API, ping @spotify on Twitter (kindly!) so they increase the rate limit for PlaylistAPIResponse Vote!",
				})
			} else {
				log.InfoCtx(r.Context(), "failed to fetch playlist from spotify", "playlist_id", playlistID, "err", err)
				s.Views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
					"playlist_id":    playlistID,
					"playlist_input": playlistLinkOrID,
					"error":          "Oops, something went wrong on fetching the playlist from Spotify",
				})
			}

			return
		}

		seg.End()

		playlist, err := playlistAPIResponse.ToPlaylist()
		if err != nil {
			if errors.Is(err, spotify.ErrPlaylistEmpty) {
				log.InfoCtx(r.Context(), "playlist is empty", "playlist_id", playlistID)
				s.Views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
					"playlist_id":    playlistID,
					"playlist_input": playlistLinkOrID,
					"error":          playlistID + " is an empty playlist! Add some songs and try again",
				})
			} else {
				log.InfoCtx(r.Context(), "failed to convert playlist", "playlist_id", playlistID, "err", err)
				s.Views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
					"playlist_id":    playlistID,
					"playlist_input": playlistLinkOrID,
					"error":          "Oops, something went wrong handling your playlist, try again later! Make sure you have at least 4 tracks with 4 different artists in your playlist!",
				})
			}

			return
		}

		seg = startSegment(r, "DBPlaylistAdd")
		_, err = s.Qry.AddPlaylist(r.Context(), sqlc.AddPlaylistParams{
			ID:      playlistID,
			Upvotes: 1,
		})

		seg.End()

		if err != nil {
			log.InfoCtx(r.Context(), "failed to add playlist", "playlist_id", playlistID, "err", err)
			s.Views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
				"playlist_id":    playlistID,
				"playlist_input": playlistLinkOrID,
				"error":          "Oops, something went wrong inserting your playlist to the database, try again later!",
			})

			return
		}

		err = playlist.AttachMetadata(r.Context(), s.Client, 1, time.Time{})
		if err != nil {
			log.InfoCtx(r.Context(), "failed to attach metadata to playlist", "playlist_id", playlistID, "err", err)
			s.Views.RenderStandardError(w)

			return
		}

		log.InfoCtx(r.Context(), "added new playlist to db", "playlist_id", playlistID)
		s.Views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
			"playlist": playlist,
		})
	}
}
