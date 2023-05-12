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

func (s *Server) handleGetPlaylist() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		playlistURI := getField(req, 0)
		playlistID := strings.TrimPrefix(playlistURI, spotify.URIPlaylistPrefix)

		{
			playlistExists, err := s.qry.PlaylistExists(req.Context(), playlistID)
			if err != nil {
				log.Printf("failed to check if playlist %s exists: %v", playlistID, err)
				s.views.RenderError(w, "")
				return
			}

			if playlistExists == 0 {
				log.Printf("playlist %s does not exist", playlistID)
				s.views.RenderError(w, "playlist does not exist on our side. add it on the home page")
				return
			}
		}

		playlist, err := s.spotify.Playlist(req.Context(), playlistID)
		if err != nil {
			log.Printf("failed to get playlist %s: %v", playlistID, err)
			s.views.RenderError(w, "")
			return
		}

		playlist.ColorsCommonFour, err = playlist.ProminentFourCoverColors()
		if err != nil {
			log.Printf("failed to query common four colors for playlist %s: %v", playlistID, err)
			s.views.RenderError(w, "")
			return
		}

		upvotes, err := s.qry.GetPlaylistUpvotes(req.Context(), playlistID)
		if err != nil {
			log.Printf("failed to query for playlist %s upvotes: %v", playlistID, err)
			s.views.RenderError(w, "")
			return
		}
		playlist.Upvotes = upvotes

		w.Header().Set("Cache-Control", "public, max-age=3600")
		s.views.Render(w, "playlist/view.tmpl", map[string]any{
			"playlist": playlist,
		})
	}
}

func (s *Server) handlePostPlaylist() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if !views.TurboStreamRequest(req) {
			http.Redirect(w, req, HomeRoute, http.StatusSeeOther)
			return
		}

		if err := req.ParseForm(); err != nil {
			log.Printf("failed to parse form: %v", err)
			s.views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
				"error": "Failed to parse form",
			})
			return
		}

		log.Printf("given playlist link or id: %s", req.Form.Get("playlist_link_or_id"))
		playlistLinkOrID := req.Form.Get("playlist_link_or_id")

		if strings.HasPrefix(playlistLinkOrID, spotify.AlbumCopiedPrefix) {
			s.views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
				"playlist_input": playlistLinkOrID,
				"error":          "Looks like you copied a album link, try again with a playlist link",
			})
			return
		}

		playlistID := strings.TrimPrefix(playlistLinkOrID, spotify.UserCopiedPlaylistPrefix)
		playlistID = strings.Split(playlistID, "?")[0]
		log.Printf("parsed playlist id: %s", playlistID)

		{
			playlistExists, err := s.qry.PlaylistExists(req.Context(), playlistID)
			if err != nil {
				log.Printf("failed to check if playlist %s exists: %v", playlistID, err)
				s.views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
					"playlist_id":    playlistID,
					"playlist_input": playlistLinkOrID,
					"error":          "Oops, something went wrong on checking if playlist already exists",
				})
				return
			}

			if playlistExists == 1 {
				log.Printf("playlist %s already exists", playlistID)
				http.Redirect(w, req, PlaylistBaseRoute+"/"+playlistID, http.StatusSeeOther)
				return
			}
		}

		log.Printf("fetching new playlist %s", playlistID)
		seg := startSegment(req, "SpotifyGetPlaylist")
		playlist, err := s.spotify.Playlist(req.Context(), playlistID)
		seg.End()
		if err != nil {
			if errors.Is(err, spotify.PlaylistNotFound) {
				log.Printf("playlist %s is empty", playlistID)
				s.views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
					"playlist_id":    playlistID,
					"playlist_input": playlistLinkOrID,
					"error":          playlistID + " not found in Spotify, double check the link and try again",
				})
				return
			} else if errors.Is(err, spotify.PlaylistEmptyErr) {
				log.Printf("playlist %s is empty", playlistID)
				s.views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
					"playlist_id":    playlistID,
					"playlist_input": playlistLinkOrID,
					"error":          playlistID + " is an empty playlist! Add some songs and try again",
				})
				return
			} else if errors.Is(err, spotify.TooManyRequestsErr) {
				log.Printf("too many requests for playlist %s", playlistID)
				s.views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
					"playlist_id":    playlistID,
					"playlist_input": playlistLinkOrID,
					"error":          "Too many requests for the Spotify API, ping @spotify on Twitter (kindly!) so they increase the rate limit for Playlist Vote!",
				})
				return
			} else {
				log.Printf("failed to fetch playlist %s playlist: %v", playlistID, err)
				s.views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
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
			s.views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
				"playlist_id":    playlistID,
				"playlist_input": playlistLinkOrID,
				"error":          "Oops, something went wrong inserting your playlist to our database, try again later!",
			})
			return
		}

		playlist.ColorsCommonFour, err = playlist.ProminentFourCoverColors()
		if err != nil {
			log.Printf("failed to query common four colors for playlist %s: %v", playlistID, err)
			s.views.RenderError(w, "")
			return
		}

		playlist.ArtistsCommonFour = playlist.MostCommonFourArtists()

		log.Printf("added new playlist %s to db", playlistID)
		s.views.Stream(w, "playlist/_new.stream.tmpl", map[string]any{
			"playlist": playlist,
		})
	}
}
