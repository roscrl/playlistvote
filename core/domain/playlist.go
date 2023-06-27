package domain

import (
	"context"
	"fmt"
	"image"
	"net/http"
	"sort"
	"strings"
	"time"
)

type PlaylistID string

type Playlist struct {
	ID          PlaylistID
	Name        string
	URI         string
	Description string
	Followers   int64

	LargestImageURL         string
	MiddleOrLargestImageURL string
	SmallestImageURL        string

	PlaylistMetadata

	Owner Owner

	Tracks []*Track
}

type PlaylistMetadata struct {
	Upvotes               int64
	AddedAt               time.Time
	CoverColorsCommonFour []string
	ArtistsCommonFour     []string
}

type Owner struct {
	URI         string
	DisplayName string
}

type Track struct {
	Name                  string
	URI                   string
	PreviewURL            string
	SmallestAlbumImageURL string
	LargestAlbumImageURL  string

	Album Album

	Artists []Artist
}

type Album struct {
	Name string
	URI  string
}

type Artist struct {
	Name string
	URI  string
}

func (p *Playlist) AttachMetadata(ctx context.Context, client *http.Client, upvotes int64, addedAt time.Time) error {
	artistsCommonFour := p.mostCommonFourArtists()

	colorsCommonFour, err := p.mostProminentFourCoverColors(ctx, client)
	if err != nil {
		return fmt.Errorf("fetching playlist %s prominent colors: %w", p.URI, err)
	}

	p.Upvotes = upvotes
	p.AddedAt = addedAt
	p.ArtistsCommonFour = artistsCommonFour
	p.CoverColorsCommonFour = colorsCommonFour

	return nil
}

func (p *Playlist) mostCommonFourArtists() []string {
	artistCount := make(map[string]int)

	for _, track := range p.Tracks {
		for _, artist := range track.Artists {
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

func (p *Playlist) mostProminentFourCoverColors(ctx context.Context, client *http.Client) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.SmallestImageURL, nil)
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
	if strings.HasPrefix(p.SmallestImageURL, "https://mosaic") {
		fourHexColors = ProminentFourColorsMosaic(img)
	} else {
		fourHexColors = ProminentFourColors(img)
	}

	return fourHexColors, nil
}
