package main

import (
	spotifymock "app/services/spotify/mock"
	"time"

	"app/services/spotify"
)

func setupServices(srv *Server, mocking bool) {
	if mocking {
		mockedServices(srv)
	} else {
		realServices(srv)
	}
}

func mockedServices(srv *Server) {
	srv.log.Info("mocking services")
	ms := spotifymock.NewServer()

	srv.spotify = &spotify.Spotify{
		ClientID:     srv.cfg.SpotifyClientID,
		ClientSecret: srv.cfg.SpotifyClientSecret,

		TokenEndpoint:    ms.TokenEndpoint,
		PlaylistEndpoint: ms.PlaylistEndpoint,

		Now: time.Now,
	}
	srv.spotify.InitTokenLifecycle()
}

func realServices(srv *Server) {
	srv.spotify = &spotify.Spotify{
		ClientID:     srv.cfg.SpotifyClientID,
		ClientSecret: srv.cfg.SpotifyClientSecret,

		TokenEndpoint:    spotify.TokenEndpoint,
		PlaylistEndpoint: spotify.PlaylistEndpoint,

		Now: time.Now,
	}
	srv.spotify.InitTokenLifecycle()
}
