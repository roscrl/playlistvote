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

type Client struct {
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
	ErrPlaylistNotFound = errors.New("playlist not found")
	ErrTooManyRequests  = errors.New("too many requests")
)

// StartTokenLifecycle starts the Spotify token lifecycle loop which will fetch a new token after the soft expiry date.
func (sc *Client) StartTokenLifecycle(ctx context.Context) {
	sc.tokenLifecycleOnce.Do(func() {
		sc.token = token{
			Client:        sc.Client,
			ClientID:      sc.ClientID,
			ClientSecret:  sc.ClientSecret,
			TokenEndpoint: sc.TokenEndpoint,
			Now:           sc.Now,
			done:          make(chan struct{}),
			firstInitDone: make(chan struct{}),
		}
		go sc.token.startRefreshLoop(ctx)

		<-sc.token.firstInitDone
	})
}

func (sc *Client) StopTokenLifecycle() {
	sc.token.stopRefreshLoop()
}

func (sc *Client) Playlist(ctx context.Context, playlistID string) (*PlaylistAPIResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sc.PlaylistEndpoint+"/"+playlistID, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+sc.token.AccessToken())

	qry := req.URL.Query()
	qry.Add("fields", PlaylistAPIQuery)

	req.URL.RawQuery = qry.Encode()

	resp, err := sc.Client.Do(req)
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

	var playlist PlaylistAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&playlist); err != nil {
		return nil, err
	}

	return &playlist, nil
}
