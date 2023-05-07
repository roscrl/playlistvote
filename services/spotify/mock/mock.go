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

	"app/services/spotify"
)

const (
	IdsPath       = "services/spotify/mock/ids.json"
	CoversPath    = "services/spotify/mock/covers"
	PlaylistsPath = "services/spotify/mock/playlists.json"
	TokenPath     = "services/spotify/mock/token.json"
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

	var playlists []spotify.Playlist
	if err := json.Unmarshal(playlistFile, &playlists); err != nil {
		log.Fatal(err)
	}

	ts := httptest.NewUnstartedServer(nil)
	ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.HasPrefix(r.URL.Path, "/playlist") {
			w.WriteHeader(http.StatusOK)

			randomIndex := rand.Intn(len(playlists))
			randomPlaylist := playlists[randomIndex]
			for i := range randomPlaylist.Images {
				randomPlaylist.Images[i].URL = ts.URL + "/cover"
			}

			randomPlaylistBytes, err := json.Marshal(randomPlaylist)
			if err != nil {
				log.Fatal(err)
			}

			w.Write(randomPlaylistBytes)
			return
		}

		switch r.URL.Path {
		case "/token":
			w.WriteHeader(http.StatusOK)

			f, err := os.ReadFile(TokenPath)
			if err != nil {
				log.Fatal(err)
			}
			w.Write(f)
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

			w.Write(f)
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Not Found"}`))
		}
	})

	ts.Start() // leaks but in mock mode it's fine

	return &SpotifyServer{
		Server:           ts,
		TokenEndpoint:    ts.URL + "/token",
		PlaylistEndpoint: ts.URL + "/playlist",
	}
}

func GenerateMockPlaylistsFile(spotifyClientID string, spotifyClientSecret string) *[]spotify.Playlist {
	spotifyClient := &spotify.Spotify{
		ClientID:     spotifyClientID,
		ClientSecret: spotifyClientSecret,

		TokenEndpoint:    spotify.TokenEndpoint,
		PlaylistEndpoint: spotify.PlaylistEndpoint,

		Now: time.Now,
	}
	spotifyClient.InitTokenLifecycle()

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

	var playlists []spotify.Playlist
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

	if err := os.WriteFile(PlaylistsPath, playlistsBytes, 0o644); err != nil {
		log.Fatal(err)
	}

	return &playlists
}
