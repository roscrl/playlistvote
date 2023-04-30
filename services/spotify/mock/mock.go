package mock

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"app/services/spotify"
)

const (
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

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.HasPrefix(r.URL.Path, "/playlist") {
			w.WriteHeader(http.StatusOK)

			randomIndex := rand.Intn(len(playlists))
			randomPlaylist := playlists[randomIndex]

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
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Not Found"}`))
		}
	}))

	return &SpotifyServer{
		Server:           ts,
		TokenEndpoint:    ts.URL + "/token",
		PlaylistEndpoint: ts.URL + "/playlist",
	}
}
