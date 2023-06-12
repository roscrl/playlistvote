package spotify

import (
	"app/core/domain"
)

type TrackAPIResponse struct {
	Artists []struct {
		Name string `json:"name"`
		URI  string `json:"uri"`
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
		URI  string `json:"uri"`
	} `json:"album"`

	DurationMs int    `json:"duration_ms"`
	Name       string `json:"name"`
	PreviewURL string `json:"preview_url"`
	URI        string `json:"uri"`
}

func (t *TrackAPIResponse) toTrack() *domain.Track {
	smallestAlbumImageURL := func(t *TrackAPIResponse) string {
		if len(t.Album.Images) == 0 {
			return ""
		}

		return t.Album.Images[len(t.Album.Images)-1].URL
	}

	largestAlbumImageURL := func(t *TrackAPIResponse) string {
		if len(t.Album.Images) == 0 {
			return ""
		}

		return t.Album.Images[0].URL
	}

	artistsOnTrack := func(t *TrackAPIResponse) []domain.Artist {
		artists := make([]domain.Artist, len(t.Artists))

		for i, artist := range t.Artists {
			artists[i] = domain.Artist{
				Name: artist.Name,
				URI:  artist.URI,
			}
		}

		return artists
	}

	return &domain.Track{
		Name:                  t.Name,
		URI:                   t.URI,
		PreviewURL:            t.PreviewURL,
		SmallestAlbumImageURL: smallestAlbumImageURL(t),
		LargestAlbumImageURL:  largestAlbumImageURL(t),
		Album: domain.Album{
			Name: t.Album.Name,
			URI:  t.Album.URI,
		},
		Artists: artistsOnTrack(t),
	}
}
