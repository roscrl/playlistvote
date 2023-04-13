package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"app/services/spotify/sampler"
)

type Spotify struct {
	ClientID         string
	ClientSecret     string
	TokenEndpoint    string
	PlaylistEndpoint string
	Now              func() time.Time

	token token
}

var (
	PlaylistEmptyErr    = errors.New("playlist is empty")
	PlaylistTooSmallErr = errors.New("playlist has too little tracks")
	PlaylistNoImageErr  = errors.New("playlist has no image")
	PlaylistNotFound    = errors.New("playlist not found")
	TooManyRequestsErr  = errors.New("too many requests")

	initTokenOnce sync.Once
)

const (
	UserCopiedPlaylistPrefix = "https://open.spotify.com/playlist/"
)

func (s *Spotify) InitToken() {
	initTokenOnce.Do(func() {
		s.token = token{
			ClientID:      s.ClientID,
			ClientSecret:  s.ClientSecret,
			TokenEndpoint: s.TokenEndpoint,
			Now:           s.Now,
		}
		go s.token.startRefreshLoop()
	})
}

type Playlist struct {
	Description  string `json:"description"`
	ID           string `json:"id"`
	ExternalUrls struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Followers struct {
		Total int `json:"total"`
	} `json:"followers"`
	Images []struct {
		Height *int   `json:"height"`
		URL    string `json:"url"`
		Width  *int   `json:"width"`
	} `json:"images"`
	Name  string `json:"name"`
	Owner struct {
		DisplayName  string `json:"display_name"`
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		ID string `json:"id"`
	} `json:"owner"`
	Tracks struct {
		Items []struct {
			Track struct {
				Artists []struct {
					Name string `json:"name"`
				} `json:"artists"`
				Album struct {
					ExternalUrls struct {
						Spotify string `json:"spotify"`
					} `json:"external_urls"`
					Images []struct {
						Height *int   `json:"height"`
						URL    string `json:"url"`
						Width  *int   `json:"width"`
					} `json:"images"`
					Name string `json:"name"`
				} `json:"album"`
				DurationMs int    `json:"duration_ms"`
				Name       string `json:"name"`
				PreviewURL string `json:"preview_url"`
			} `json:"track"`
		} `json:"items"`
	} `json:"tracks"`
	URI string `json:"uri"`
}

func (p *Playlist) LargestImageURL() string {
	// Assuming Spotify API Images JSON field ordering does not change,
	// The first image is the largest from the Spotify API which is either a large 640x640 2x2 mosaic image or a user uploaded image
	return p.Images[0].URL
}

func (p *Playlist) MiddleOrLargestImageURL() string {
	if len(p.Images) > 1 {
		return p.Images[1].URL
	}
	return p.Images[0].URL
}

func (p *Playlist) SmallestImageURL() string {
	return p.Images[len(p.Images)-1].URL
}

func (p *Playlist) MostCommonFourArtists() []string {
	artistCount := make(map[string]int)
	for _, item := range p.Tracks.Items {
		for _, artist := range item.Track.Artists {
			artistCount[artist.Name]++
		}
	}

	type kv struct {
		Key   string
		Value int
	}

	var ss []kv
	for k, v := range artistCount {
		ss = append(ss, kv{k, v})
	}

	sortByValueDesc := func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	}
	sort.Slice(ss, sortByValueDesc)

	topCount := 4
	mostCommonArtists := make([]string, topCount)
	for i := 0; i < topCount; i++ {
		mostCommonArtists[i] = ss[i].Key
	}

	return mostCommonArtists
}

func (p *Playlist) ProminentFourCoverColors() ([]string, error) {
	smallestImageURL := p.SmallestImageURL()

	resp, err := http.Get(smallestImageURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %v", err)
	}

	var fourHexColors []string
	if strings.HasPrefix(smallestImageURL, "https://mosaic") {
		fourHexColors = sampler.ProminentFourColorsMosaic(img)
	} else {
		fourHexColors = sampler.ProminentFourColors(img)
	}

	return fourHexColors, nil
}

func (s *Spotify) Playlist(ctx context.Context, playlistId string) (*Playlist, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", s.PlaylistEndpoint+"/"+playlistId, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.token.AccessToken())

	q := req.URL.Query()
	q.Add("fields", "id, name, images, description, owner(display_name, id, external_urls(spotify)), followers, uri, external_urls, tracks.items(track(name, duration_ms, preview_url, artists(name), album(name, images, external_urls(spotify))))")
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
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

	if len(playlist.Tracks.Items) == 0 {
		return nil, PlaylistEmptyErr
	}

	if len(playlist.Images) == 0 {
		return nil, PlaylistNoImageErr
	}

	if len(playlist.Tracks.Items) < 4 {
		return nil, PlaylistTooSmallErr
	}

	return &playlist, nil
}
