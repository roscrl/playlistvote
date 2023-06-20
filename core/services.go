package core

import (
	"context"
	"time"

	"app/core/broadcast"
	"app/core/session"
	"app/core/spotify"
	spotifymock "app/core/spotify/mock"
)

func setupServices(s *Server) {
	s.SessionStore = session.NewStore()
	s.UpvoteBroadcasterService = broadcast.NewUpvoteBroadcasterService()
	s.UpvoteBroadcasterService.Start(s.SessionStore)

	if s.Cfg.Mocking {
		mockedServices(s)
	} else {
		realServices(s)
	}
}

func mockedServices(s *Server) {
	s.Log.Info("mocking services")

	mockServer := spotifymock.NewServer()

	s.Spotify = &spotify.Client{
		Client:       s.Client,
		ClientID:     s.Cfg.SpotifyClientID,
		ClientSecret: s.Cfg.SpotifyClientSecret,

		TokenEndpoint:    mockServer.TokenEndpoint,
		PlaylistEndpoint: mockServer.PlaylistEndpoint,

		Now: time.Now,
	}
	s.Spotify.StartTokenLifecycle(context.Background())

	s.Log.Info("services mocked")
}

func realServices(s *Server) {
	s.Log.Info("initializing services")

	s.Spotify = &spotify.Client{
		Client:       s.Client,
		ClientID:     s.Cfg.SpotifyClientID,
		ClientSecret: s.Cfg.SpotifyClientSecret,

		TokenEndpoint:    spotify.TokenEndpoint,
		PlaylistEndpoint: spotify.PlaylistEndpoint,

		Now: time.Now,
	}
	s.Spotify.StartTokenLifecycle(context.Background())

	s.Log.Info("services initialized")
}
