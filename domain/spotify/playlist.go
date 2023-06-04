package spotify

import (
	"context"
	"fmt"
	"image"
	"net/http"
	"sort"
	"strings"

	"app/domain/spotify/sampler"
	"golang.org/x/exp/slices"
)

const (
	MinPlaylistTracks  = 4
	MinPlaylistArtists = 4
)

type PlaylistMetadata struct {
	Upvotes           int64
	ColorsCommonFour  []string
	ArtistsCommonFour []string
	EagerLoadImage    bool
}

type Playlist struct {
	PlaylistMetadata

	Description  string `json:"description"`
	ID           string `json:"id"`
	ExternalUrls struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Followers struct {
		Total int64 `json:"total"`
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
		ID  string `json:"id"`
		URI string `json:"uri"`
	} `json:"owner"`
	Tracks struct {
		Items []struct {
			Track `json:"track"`
		} `json:"items"`
	} `json:"tracks"`
	URI string `json:"uri"`
}

func (p *Playlist) valid() error {
	if len(p.Tracks.Items) == 0 {
		return ErrPlaylistEmpty
	}

	if len(p.Images) == 0 {
		return ErrPlaylistNoImage
	}

	if len(p.Tracks.Items) < MinPlaylistTracks {
		return ErrPlaylistTooSmallTracks
	}

	if !p.HasEnoughArtists() {
		return ErrPlaylistTooSmallArtists
	}

	return nil
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

func (p *Playlist) HasEnoughArtists() bool {
	var uniqueArtists []string

	for _, item := range p.Tracks.Items {
		for _, artist := range item.Track.Artists {
			if !slices.Contains(uniqueArtists, artist.Name) {
				uniqueArtists = append(uniqueArtists, artist.Name)
				if len(uniqueArtists) >= MinPlaylistArtists {
					return true
				}
			}
		}
	}

	return false
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

	sortedArtistsSet := make([]kv, 0, len(artistCount))
	for k, v := range artistCount {
		sortedArtistsSet = append(sortedArtistsSet, kv{k, v})
	}

	sortByValueDesc := func(i, j int) bool {
		return sortedArtistsSet[i].Value > sortedArtistsSet[j].Value
	}
	sort.Slice(sortedArtistsSet, sortByValueDesc)

	topCount := 4

	mostCommonArtists := make([]string, topCount)
	for i := 0; i < topCount; i++ {
		mostCommonArtists[i] = sortedArtistsSet[i].Key
	}

	return mostCommonArtists
}

func (p *Playlist) MostProminentFourCoverColors(ctx context.Context, client *http.Client) ([]string, error) {
	smallestImageURL := p.SmallestImageURL()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, smallestImageURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %w", err)
	}

	var fourHexColors []string
	if strings.HasPrefix(smallestImageURL, "https://mosaic") {
		fourHexColors = sampler.ProminentFourColorsMosaic(img)
	} else {
		fourHexColors = sampler.ProminentFourColors(img)
	}

	return fourHexColors, nil
}
