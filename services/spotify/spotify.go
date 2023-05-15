package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	TokenEndpoint    = "https://accounts.spotify.com/api/token"
	PlaylistEndpoint = "https://api.spotify.com/v1/playlists"

	UserCopiedPlaylistPrefix = "https://open.spotify.com/playlist/"
	AlbumCopiedPrefix        = "https://open.spotify.com/album/"
	URIPlaylistPrefix        = "spotify:playlist:"
)

type Spotify struct {
	ClientID         string
	ClientSecret     string
	TokenEndpoint    string
	PlaylistEndpoint string
	Now              func() time.Time

	token         token
	initTokenOnce sync.Once
}

var (
	PlaylistEmptyErr           = errors.New("playlist is empty")
	PlaylistTooSmallTracksErr  = errors.New("playlist has too little tracks")
	PlaylistTooSmallArtistsErr = errors.New("playlist has too little artists")
	PlaylistNoImageErr         = errors.New("playlist has no image")
	PlaylistNotFound           = errors.New("playlist not found")
	TooManyRequestsErr         = errors.New("too many requests")
)

func (s *Spotify) StartTokenLifecycle() {
	s.initTokenOnce.Do(func() {
		s.token = token{
			ClientID:      s.ClientID,
			ClientSecret:  s.ClientSecret,
			TokenEndpoint: s.TokenEndpoint,
			Now:           s.Now,
			done:          make(chan struct{}),
			firstInitDone: make(chan struct{}),
		}
		go s.token.startRefreshLoop()

		<-s.token.firstInitDone
	})
}

func (s *Spotify) StopTokenLifecycle() {
	s.token.stopRefreshLoop()
}

func (s *Spotify) Playlist(ctx context.Context, playlistId string) (*Playlist, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", s.PlaylistEndpoint+"/"+playlistId, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.token.AccessToken())

	q := req.URL.Query()
	q.Add("fields", "id, name, images, description, owner(display_name, id, external_urls(spotify), uri), followers, uri, external_urls, tracks.items(track(name, duration_ms, preview_url, uri, artists(name, uri), album(name, images, external_urls(spotify), uri)))")
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, PlaylistNotFound
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, TooManyRequestsErr
	}

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("reading response body: %w", err)
		}
		return nil, fmt.Errorf("spotify: %s %s, request: %s", resp.Status, body, req.URL)
	}

	var playlist Playlist
	if err := json.NewDecoder(resp.Body).Decode(&playlist); err != nil {
		return nil, err
	}

	if err := playlist.valid(); err != nil {
		return nil, err
	}

	return &playlist, nil
}

func (s *Spotify) PlaylistMetadata(ctx context.Context, playlistId string) (*Playlist, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", s.PlaylistEndpoint+"/"+playlistId, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.token.AccessToken())

	q := req.URL.Query()
	q.Add("fields", "id, name, images, description, owner(display_name, id, external_urls(spotify), uri), followers, uri, external_urls, tracks.items(track(artists(name)))")
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, PlaylistNotFound
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, TooManyRequestsErr
	}

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("reading response body: %w", err)
		}
		return nil, fmt.Errorf("spotify: %s %s, request: %s", resp.Status, body, req.URL)
	}

	var playlist Playlist
	if err := json.NewDecoder(resp.Body).Decode(&playlist); err != nil {
		return nil, err
	}

	if err := playlist.valid(); err != nil {
		return nil, err
	}

	return &playlist, nil
}
