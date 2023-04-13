package spotify

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
)

func MockEndpoints() (*httptest.Server, *httptest.Server) {
	tsToken := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		mockToken := `{
		  "access_token": "TQBC_KDHhdQfPm_jLG0vxDlK_f0J1Vm57gIjG9uFJlWU8ySKySX_FhQ3re3iuReIF2No-Kz8fJNxoEaVdONxEFD0TkpZNsKBJCcHbtCaBPa0MigvRB0",
		  "token_type": "Bearer",
		  "expires_in": 3600
        }`
		w.Write([]byte(mockToken))
	}))

	tsPlaylist := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		f, err := os.ReadFile("services/spotify/mock_playlist.json")
		if err != nil {
			panic(err)
		}
		w.Write(f)
	}))

	return tsToken, tsPlaylist
}

func MockPlaylists() []Playlist {
	file, err := os.ReadFile("services/spotify/mock_playlists.json")
	if err != nil {
		log.Fatal(err)
	}

	var playlists []Playlist
	err = json.Unmarshal(file, &playlists)
	if err != nil {
		log.Fatal(err)
	}

	return playlists
}
