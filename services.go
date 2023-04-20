package main

import (
	"app/services/spotify"
	"log"
	"net/http"
	"time"
)

func setupServices(srv *Server, mocking bool) {
	http.DefaultClient.Timeout = 10 * time.Second
	if mocking {
		log.Println("mocking enabled")
		tsToken, tsPlaylist := spotify.MockEndpoints("services/spotify/mock_playlist.json")
		srv.spotify = &spotify.Spotify{
			ClientID:     srv.cfg.SpotifyClientID,
			ClientSecret: srv.cfg.SpotifyClientSecret,

			TokenEndpoint:    tsToken.URL,
			PlaylistEndpoint: tsPlaylist.URL,

			Now: time.Now,
		}
		srv.spotify.InitTokenLifecycle()
	} else {
		srv.spotify = &spotify.Spotify{
			ClientID:     srv.cfg.SpotifyClientID,
			ClientSecret: srv.cfg.SpotifyClientSecret,

			TokenEndpoint:    "https://accounts.spotify.com/api/token",
			PlaylistEndpoint: "https://api.spotify.com/v1/playlists",

			Now: time.Now,
		}
		srv.spotify.InitTokenLifecycle()
	}
}
