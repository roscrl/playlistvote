package core

import (
	"context"
	"time"

	"app/core/spotify"
	spotifymock "app/core/spotify/mock"
)

func setupServices(srv *Server) {
	if srv.Cfg.Mocking {
		mockedServices(srv)
	} else {
		realServices(srv)
	}
}

func mockedServices(srv *Server) {
	srv.Log.Info("mocking services")

	mockServer := spotifymock.NewServer()

	srv.Spotify = &spotify.Client{
		Client:       srv.Client,
		ClientID:     srv.Cfg.SpotifyClientID,
		ClientSecret: srv.Cfg.SpotifyClientSecret,

		TokenEndpoint:    mockServer.TokenEndpoint,
		PlaylistEndpoint: mockServer.PlaylistEndpoint,

		Now: time.Now,
	}
	srv.Spotify.StartTokenLifecycle(context.Background())

	srv.Log.Info("services mocked")
}

func realServices(srv *Server) {
	srv.Log.Info("initializing services")

	srv.Spotify = &spotify.Client{
		Client:       srv.Client,
		ClientID:     srv.Cfg.SpotifyClientID,
		ClientSecret: srv.Cfg.SpotifyClientSecret,

		TokenEndpoint:    spotify.TokenEndpoint,
		PlaylistEndpoint: spotify.PlaylistEndpoint,

		Now: time.Now,
	}
	srv.Spotify.StartTokenLifecycle(context.Background())

	srv.Log.Info("services initialized")
}
