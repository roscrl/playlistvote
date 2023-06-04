package main

import (
	"context"
	"time"

	"app/domain/spotify"
	spotifymock "app/domain/spotify/mock"
)

func setupServices(srv *Server) {
	if srv.cfg.Mocking {
		mockedServices(srv)
	} else {
		realServices(srv)
	}
}

func mockedServices(srv *Server) {
	srv.log.Info("mocking domain")

	mockServer := spotifymock.NewServer()

	srv.spotify = &spotify.Spotify{
		Client:       srv.client,
		ClientID:     srv.cfg.SpotifyClientID,
		ClientSecret: srv.cfg.SpotifyClientSecret,

		TokenEndpoint:    mockServer.TokenEndpoint,
		PlaylistEndpoint: mockServer.PlaylistEndpoint,

		Now: time.Now,
	}
	srv.spotify.StartTokenLifecycle(context.Background())
}

func realServices(srv *Server) {
	srv.spotify = &spotify.Spotify{
		Client:       srv.client,
		ClientID:     srv.cfg.SpotifyClientID,
		ClientSecret: srv.cfg.SpotifyClientSecret,

		TokenEndpoint:    spotify.TokenEndpoint,
		PlaylistEndpoint: spotify.PlaylistEndpoint,

		Now: time.Now,
	}
	srv.spotify.StartTokenLifecycle(context.Background())
}
