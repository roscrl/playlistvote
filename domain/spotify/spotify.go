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
	TokenEndpoint    = "https://accounts.spotify.com/api/token" //nolint:gosec
	PlaylistEndpoint = "https://api.spotify.com/v1/playlists"

	UserCopiedPlaylistPrefix = "https://open.spotify.com/playlist/"
	AlbumCopiedPrefix        = "https://open.spotify.com/album/"
	URIPlaylistPrefix        = "spotify:playlist:"
)

type Spotify struct {
	Client           *http.Client
	ClientID         string
	ClientSecret     string
	TokenEndpoint    string
	PlaylistEndpoint string
	Now              func() time.Time

	token              token
	tokenLifecycleOnce sync.Once
}

var (
	ErrPlaylistEmpty           = errors.New("playlist is empty")
	ErrPlaylistTooSmallTracks  = errors.New("playlist has too little tracks")
	ErrPlaylistTooSmallArtists = errors.New("playlist has too little artists")
	ErrPlaylistNoImage         = errors.New("playlist has no image")
	ErrPlaylistNotFound        = errors.New("playlist not found")
	ErrTooManyRequests         = errors.New("too many requests")
)

// StartTokenLifecycle starts the Spotify token lifecycle loop which will fetch a new token after the soft expiry date.
func (s *Spotify) StartTokenLifecycle(ctx context.Context) {
	s.tokenLifecycleOnce.Do(func() {
		s.token = token{
			Client:        s.Client,
			ClientID:      s.ClientID,
			ClientSecret:  s.ClientSecret,
			TokenEndpoint: s.TokenEndpoint,
			Now:           s.Now,
			done:          make(chan struct{}),
			firstInitDone: make(chan struct{}),
		}
		go s.token.startRefreshLoop(ctx)

		<-s.token.firstInitDone
	})
}

func (s *Spotify) StopTokenLifecycle() {
	s.token.stopRefreshLoop()
}

func (s *Spotify) Playlist(ctx context.Context, playlistID string) (*Playlist, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.PlaylistEndpoint+"/"+playlistID, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.token.AccessToken())

	qry := req.URL.Query()
	qry.Add("fields", `
		id, name, images, description, 
		owner(display_name, id, external_urls(spotify), uri), 
		followers, uri, external_urls, 
		tracks.items(
			track(name, duration_ms, preview_url, uri, artists(name, uri), album(name, images, external_urls(spotify), uri))
		)
	`)

	req.URL.RawQuery = qry.Encode()

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrPlaylistNotFound
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrTooManyRequests
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

func (s *Spotify) PlaylistMetadata(ctx context.Context, playlistID string) (*Playlist, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.PlaylistEndpoint+"/"+playlistID, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.token.AccessToken())

	q := req.URL.Query()
	q.Add("fields", "id, name, images, description, owner(display_name, id, external_urls(spotify), uri), followers, uri, external_urls, tracks.items(track(artists(name)))")
	req.URL.RawQuery = q.Encode()

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrPlaylistNotFound
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrTooManyRequests
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
