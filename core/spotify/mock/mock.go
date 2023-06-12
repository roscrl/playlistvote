//nolint:gosec,gomnd
package mock

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"app/core/spotify"
)

const (
	IdsPath       = "core/spotify/mock/ids.json"
	CoversPath    = "core/spotify/mock/covers"
	PlaylistsPath = "core/spotify/mock/playlists.json"
	TokenPath     = "core/spotify/mock/token.json"
)

type SpotifyServer struct {
	Server            *httptest.Server
	TokenEndpoint     string
	PlaylistEndpoint  string
	PlaylistsEndpoint string
}

func NewServer() *SpotifyServer {
	playlistFile, err := os.ReadFile(PlaylistsPath)
	if err != nil {
		log.Fatal(err)
	}

	var playlists []spotify.PlaylistAPIResponse
	if err := json.Unmarshal(playlistFile, &playlists); err != nil {
		log.Fatal(err)
	}

	testServer := httptest.NewUnstartedServer(nil)
	testServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.HasPrefix(r.URL.Path, "/playlist") {
			w.WriteHeader(http.StatusOK)

			randomIndex := rand.Intn(len(playlists))
			randomPlaylist := playlists[randomIndex]
			for i := range randomPlaylist.Images {
				randomPlaylist.Images[i].URL = testServer.URL + "/cover"
			}

			randomPlaylistBytes, err := json.Marshal(randomPlaylist)
			if err != nil {
				log.Fatal(err)
			}

			_, _ = w.Write(randomPlaylistBytes)

			return
		}

		switch r.URL.Path {
		case "/token":
			w.WriteHeader(http.StatusOK)

			f, err := os.ReadFile(TokenPath)
			if err != nil {
				log.Fatal(err)
			}
			_, _ = w.Write(f)
		case "/cover":
			w.WriteHeader(http.StatusOK)

			files, err := os.ReadDir(CoversPath)
			if err != nil {
				log.Fatal(err)
			}

			randomIndex := rand.Intn(len(files))
			randomFile := files[randomIndex]

			f, err := os.ReadFile(CoversPath + "/" + randomFile.Name())
			if err != nil {
				log.Fatal(err)
			}

			_, _ = w.Write(f)
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error": "Not Found"}`))
		}
	})

	testServer.Start() // leaks but in mock mode it's fine

	return &SpotifyServer{
		Server:           testServer,
		TokenEndpoint:    testServer.URL + "/token",
		PlaylistEndpoint: testServer.URL + "/playlist",
	}
}

func GenerateMockPlaylistsFile(spotifyClientID string, spotifyClientSecret string) *[]spotify.PlaylistAPIResponse {
	spotifyClient := &spotify.Client{
		Client:       &http.Client{Timeout: 5 * time.Second},
		ClientID:     spotifyClientID,
		ClientSecret: spotifyClientSecret,

		TokenEndpoint:    spotify.TokenEndpoint,
		PlaylistEndpoint: spotify.PlaylistEndpoint,

		Now: time.Now,
	}
	spotifyClient.StartTokenLifecycle(context.Background())

	idsFile, err := os.ReadFile(IdsPath)
	if err != nil {
		log.Fatal(err)
	}

	ids := struct {
		Ids []string `json:"ids"`
	}{}
	if err := json.Unmarshal(idsFile, &ids); err != nil {
		log.Fatal(err)
	}

	playlists := make([]spotify.PlaylistAPIResponse, 0, len(ids.Ids))

	for _, id := range ids.Ids {
		playlist, err := spotifyClient.Playlist(context.Background(), id)
		if err != nil {
			log.Fatal(err)
		}

		playlists = append(playlists, *playlist)
	}

	playlistsBytes, err := json.MarshalIndent(playlists, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile(PlaylistsPath, playlistsBytes, 0o600); err != nil {
		log.Fatal(err)
	}

	return &playlists
}
