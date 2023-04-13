package main

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"app/db/sqlc"
	"app/services/spotify"
	"app/views"
)

func (s *Server) handlePostPlaylist() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if !views.TurboStreamRequest(req) {
			http.Redirect(w, req, HomeRoute, http.StatusSeeOther)
			return
		}

		if err := req.ParseForm(); err != nil {
			log.Printf("failed to parse form: %v", err)
			s.views.RenderStream(w, "playlist/_new.stream.tmpl", map[string]any{
				"error": "Failed to parse form",
			})
			return
		}

		log.Printf("given playlist link or id: %s", req.Form.Get("playlist_link_or_id"))
		playlistLinkOrID := req.Form.Get("playlist_link_or_id")
		playlistID := strings.TrimPrefix(playlistLinkOrID, spotify.UserCopiedPlaylistPrefix)
		playlistID = strings.Split(playlistID, "?")[0]
		log.Printf("parsed playlist id: %s", playlistID)

		res, err := s.qry.PlaylistExists(req.Context(), playlistID)
		if err != nil {
			log.Printf("failed to check if playlist %s exists: %v", playlistID, err)
			s.views.RenderStream(w, "playlist/_new.stream.tmpl", map[string]any{
				"playlist_id":    playlistID,
				"playlist_input": playlistLinkOrID,
				"error":          "Oops, something went wrong on checking if playlist already exists",
			})
			return
		}

		if exists, ok := res.(int64); ok {
			if exists == 1 {
				log.Printf("playlist %s already exists", playlistID)
				// TODO redirect to playlist page
				s.views.RenderStream(w, "playlist/_new.stream.tmpl", map[string]any{
					"playlist_id":    playlistID,
					"playlist_input": playlistLinkOrID,
					"error":          "Playlist " + playlistID + " already exists",
				})
				return
			}
		} else {
			log.Printf("something went wrong running query to check if playlist %s exists: %v", playlistID, err)
			s.views.RenderStream(w, "playlist/_new.stream.tmpl", map[string]any{
				"playlist_id":    playlistID,
				"playlist_input": playlistLinkOrID,
				"error":          "Oops, something went wrong on checking if playlist already exists",
			})
			return
		}

		log.Printf("fetching new playlist %s", playlistID)
		seg := startSegment(req, "SpotifyGetPlaylist")
		playlist, err := s.spotify.Playlist(req.Context(), playlistID)
		seg.End()
		if err != nil {
			if errors.Is(err, spotify.PlaylistNotFound) {
				log.Printf("playlist %s is empty", playlistID)
				s.views.RenderStream(w, "playlist/_new.stream.tmpl", map[string]any{
					"playlist_id":    playlistID,
					"playlist_input": playlistLinkOrID,
					"error":          playlistID + " not found in Spotify, double check the link and try again",
				})
				return
			} else if errors.Is(err, spotify.PlaylistEmptyErr) {
				log.Printf("playlist %s is empty", playlistID)
				s.views.RenderStream(w, "playlist/_new.stream.tmpl", map[string]any{
					"playlist_id":    playlistID,
					"playlist_input": playlistLinkOrID,
					"error":          playlistID + " is an empty playlist! Add some songs and try again",
				})
				return
			} else if errors.Is(err, spotify.TooManyRequestsErr) {
				log.Printf("too many requests for playlist %s", playlistID)
				s.views.RenderStream(w, "playlist/_new.stream.tmpl", map[string]any{
					"playlist_id":    playlistID,
					"playlist_input": playlistLinkOrID,
					"error":          "Too many requests for the Spotify API, ping @spotify on Twitter (kindly!) so they increase the rate limit for Playlist Vote!",
				})
				return
			} else {
				log.Printf("failed to fetch playlist %s playlist: %v", playlistID, err)
				s.views.RenderStream(w, "playlist/_new.stream.tmpl", map[string]any{
					"playlist_id":    playlistID,
					"playlist_input": playlistLinkOrID,
					"error":          "Oops, something went wrong handling your playlist, try again later!",
				})
				return
			}
		}

		seg = startSegment(req, "DBAddPlaylist")
		_, err = s.qry.AddPlaylist(req.Context(), sqlc.AddPlaylistParams{
			ID:      playlistID,
			Upvotes: 1,
			AddedAt: time.Now().Unix(),
		})
		seg.End()
		if err != nil {
			log.Printf("failed to add playlist %s playlist: %v", playlistID, err)
			s.views.RenderStream(w, "playlist/_new.stream.tmpl", map[string]any{
				"playlist_id":    playlistID,
				"playlist_input": playlistLinkOrID,
				"error":          "Oops, something went wrong inserting your playlist to our database, try again later!",
			})
			return
		}

		log.Printf("added new playlist %s to db", playlistID)
		s.views.RenderStream(w, "playlist/_new.stream.tmpl", map[string]any{
			"playlist": playlist,
		})
	}
}
