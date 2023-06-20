package spotify

import (
	"errors"

	"app/core/domain"
	"golang.org/x/exp/slices"
)

const (
	MinPlaylistTracks  = 4
	MinPlaylistArtists = 4

	PlaylistAPIQuery = `
		id, name, description, images
		owner(id, display_name, external_urls(spotify), uri), 
		followers, uri, external_urls, 
		tracks.items(
			track(name, duration_ms, preview_url, uri, artists(name, uri), album(name, images, external_urls(spotify), uri))
		)
	`
)

var (
	ErrPlaylistEmpty           = errors.New("playlist is empty")
	ErrPlaylistTooSmallTracks  = errors.New("playlist has too little tracks")
	ErrPlaylistTooSmallArtists = errors.New("playlist has too little artists")
	ErrPlaylistNoImage         = errors.New("playlist has no image")
)

type PlaylistAPIResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
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
			TrackAPIResponse `json:"track"`
		} `json:"items"`
	} `json:"tracks"`
	URI string `json:"uri"`
}

func (p *PlaylistAPIResponse) ToPlaylist() (*domain.Playlist, error) {
	if err := p.validPlaylist(); err != nil {
		return nil, err
	}

	largestImageURL := func(p *PlaylistAPIResponse) string {
		// Assuming Spotify API Images JSON field ordering does not change,
		// The first image is the largest from the Spotify API which is either a large 640x640 2x2 mosaic image or a user uploaded image
		return p.Images[0].URL
	}

	middleOrLargestImageURL := func(p *PlaylistAPIResponse) string {
		if len(p.Images) > 1 {
			return p.Images[1].URL
		}

		return p.Images[0].URL
	}

	smallestImageURL := p.Images[len(p.Images)-1].URL

	tracksOnPlaylist := func(p *PlaylistAPIResponse) []*domain.Track {
		tracks := make([]*domain.Track, len(p.Tracks.Items))

		for i, item := range p.Tracks.Items {
			tracks[i] = item.TrackAPIResponse.toTrack()
		}

		return tracks
	}

	return &domain.Playlist{
		ID:          domain.PlaylistID(p.ID),
		Name:        p.Name,
		URI:         p.URI,
		Description: p.Description,

		LargestImageURL:         largestImageURL(p),
		MiddleOrLargestImageURL: middleOrLargestImageURL(p),
		SmallestImageURL:        smallestImageURL,

		Followers: p.Followers.Total,

		Owner: domain.Owner{
			DisplayName: p.Owner.DisplayName,
			URI:         p.Owner.URI,
		},

		Tracks: tracksOnPlaylist(p),
	}, nil
}

func (p *PlaylistAPIResponse) validPlaylist() error {
	if len(p.Tracks.Items) == 0 {
		return ErrPlaylistEmpty
	}

	if len(p.Images) == 0 {
		return ErrPlaylistNoImage
	}

	if len(p.Tracks.Items) < MinPlaylistTracks {
		return ErrPlaylistTooSmallTracks
	}

	hasEnoughArtists := func(p *PlaylistAPIResponse) bool {
		var uniqueArtists []string

		for _, item := range p.Tracks.Items {
			for _, artist := range item.TrackAPIResponse.Artists {
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

	if !hasEnoughArtists(p) {
		return ErrPlaylistTooSmallArtists
	}

	return nil
}
