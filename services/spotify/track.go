package spotify

import "strings"

type Track struct {
	Artists []ArtistWithUri `json:"artists"`
	Album   struct {
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

type ArtistWithUri struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

func (t *Track) SmallestAlbumImageURL() string {
	if len(t.Album.Images) == 0 {
		return ""
	}

	return t.Album.Images[len(t.Album.Images)-1].URL
}

func (t *Track) ArtistsCommaSeparated() string {
	var artists []string
	for _, artist := range t.Artists {
		artists = append(artists, artist.Name)
	}
	return strings.Join(artists, ", ")
}
