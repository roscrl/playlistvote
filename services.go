package main

import (
	"time"

	spotifymock "app/services/spotify/mock"

	"app/services/spotify"
)

func setupServices(srv *Server) {
	if srv.cfg.Mocking {
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
	srv.spotify.StartTokenLifecycle()
}

func realServices(srv *Server) {
	srv.spotify = &spotify.Spotify{
		ClientID:     srv.cfg.SpotifyClientID,
		ClientSecret: srv.cfg.SpotifyClientSecret,

		TokenEndpoint:    spotify.TokenEndpoint,
		PlaylistEndpoint: spotify.PlaylistEndpoint,

		Now: time.Now,
	}
	srv.spotify.StartTokenLifecycle()
}
