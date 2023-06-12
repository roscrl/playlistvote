package core

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"app/core/db/sqlc"
	"app/core/spotify"
	"app/core/views"
)

func (s *Server) handlePlaylistView() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		playlistURI := getField(req, 0)
		playlistID := strings.TrimPrefix(playlistURI, spotify.URIPlaylistPrefix)

		{
			playlistExists, err := s.Qry.PlaylistExists(req.Context(), playlistID)
			if err != nil {
				log.Printf("failed to check if playlist %s exists: %v", playlistID, err)
				s.Views.RenderStandardError(w)

				return
			}

			if playlistExists == 0 {
				log.Printf("playlist %s does not exist", playlistID)
				s.Views.RenderError(w, "playlist does not exist on our side. add it on the home page")

				return
			}
		}

		playlistAPIResponse, err := s.Spotify.Playlist(req.Context(), playlistID)
		if err != nil {
			log.Printf("failed to get playlist %s: %v", playlistID, err)
			s.Views.RenderStandardError(w)

			return
		}

		upvotes, err := s.Qry.GetPlaylistUpvotes(req.Context(), playlistID)
		if err != nil {
			log.Printf("failed to query for playlist %s upvotes: %v", playlistID, err)
			s.Views.RenderError(w, "")

			return
		}

		playlist, err := playlistAPIResponse.ToPlaylist()
		if err != nil {
			log.Printf("transforming playlist %s to playlist: %v", playlistID, err)
			s.Views.RenderStandardError(w)

			return
		}

		err = playlist.AttachMetadata(req.Context(), s.Client, upvotes)
		if err != nil {
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=3600")
		s.Views.Render(w, "playlist/view.tmpl", map[string]any{
			"playlist": playlist,
		})
	}
}

func (s *Server) handlePlaylistCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if !views.TurboStreamRequest(req) {
			http.Redirect(w, req, RouteHome, http.StatusSeeOther)

			return
		}

		if err := req.ParseForm(); err != nil {
			log.Printf("failed to parse form: %v", err)
			s.Views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
				"error": "Failed to parse form",
			})

			return
		}

		log.Printf("given playlist link or id: %s", req.Form.Get("playlist_link_or_id"))
		playlistLinkOrID := req.Form.Get("playlist_link_or_id")

		if strings.HasPrefix(playlistLinkOrID, spotify.AlbumCopiedPrefix) {
			s.Views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
				"playlist_input": playlistLinkOrID,
				"error":          "Looks like you copied a album link, try again with a playlist link",
			})

			return
		}

		playlistID := strings.TrimPrefix(playlistLinkOrID, spotify.UserCopiedPlaylistPrefix)
		playlistID = strings.Split(playlistID, "?")[0]
		log.Printf("parsed playlist id: %s", playlistID)

		{
			playlistExists, err := s.Qry.PlaylistExists(req.Context(), playlistID)
			if err != nil {
				log.Printf("failed to check if playlist %s exists: %v", playlistID, err)
				s.Views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
					"playlist_id":    playlistID,
					"playlist_input": playlistLinkOrID,
					"error":          "Oops, something went wrong on checking if playlist already exists",
				})

				return
			}

			if playlistExists == 1 {
				log.Printf("playlist %s already exists", playlistID)
				http.Redirect(w, req, RoutePlaylistBase+"/"+playlistID, http.StatusSeeOther)

				return
			}
		}

		log.Printf("fetching new playlist %s", playlistID)

		seg := startSegment(req, "SpotifyPlaylistGet")
		playlistAPIResponse, err := s.Spotify.Playlist(req.Context(), playlistID)

		seg.End()

		if err != nil {
			if errors.Is(err, spotify.ErrPlaylistNotFound) {
				log.Printf("playlist %s is empty", playlistID)
				s.Views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
					"playlist_id":    playlistID,
					"playlist_input": playlistLinkOrID,
					"error":          playlistID + " not found in Spotify, double check the link and try again",
				})
			} else if errors.Is(err, spotify.ErrTooManyRequests) {
				log.Printf("too many requests for playlist %s", playlistID)
				s.Views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
					"playlist_id":    playlistID,
					"playlist_input": playlistLinkOrID,
					"error":          "Too many requests for the Spotify API, ping @spotify on Twitter (kindly!) so they increase the rate limit for PlaylistAPIResponse Vote!",
				})
			} else {
				log.Printf("failed to fetch playlist %s playlist: %v", playlistID, err)
				s.Views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
					"playlist_id":    playlistID,
					"playlist_input": playlistLinkOrID,
					"error":          "Oops, something went wrong on fetching the playlist from Spotify",
				})
			}

			return
		}

		playlist, err := playlistAPIResponse.ToPlaylist()
		if err != nil {
			if errors.Is(err, spotify.ErrPlaylistEmpty) {
				log.Printf("playlist %s is empty", playlistID)
				s.Views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
					"playlist_id":    playlistID,
					"playlist_input": playlistLinkOrID,
					"error":          playlistID + " is an empty playlist! Add some songs and try again",
				})
			} else {
				log.Printf("failed to fetch playlist %s playlist: %v", playlistID, err)
				s.Views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
					"playlist_id":    playlistID,
					"playlist_input": playlistLinkOrID,
					"error":          "Oops, something went wrong handling your playlist, try again later! Make sure you have at least 4 tracks with 4 different artists in your playlist!",
				})
			}

			return
		}

		seg = startSegment(req, "DBPlaylistAdd")
		_, err = s.Qry.AddPlaylist(req.Context(), sqlc.AddPlaylistParams{
			ID:      playlistID,
			Upvotes: 1,
		})

		seg.End()

		if err != nil {
			log.Printf("failed to add playlist %s playlist: %v", playlistID, err)
			s.Views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
				"playlist_id":    playlistID,
				"playlist_input": playlistLinkOrID,
				"error":          "Oops, something went wrong inserting your playlist to our database, try again later!",
			})

			return
		}

		err = playlist.AttachMetadata(req.Context(), s.Client, 1)
		if err != nil {
			log.Printf("failed to attach metadata to playlist %s: %v", playlistID, err)
			s.Views.RenderStandardError(w)

			return
		}

		log.Printf("added new playlist %s to db", playlistID)
		s.Views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
			"playlist": playlist,
		})
	}
}
