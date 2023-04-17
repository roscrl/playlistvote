package spotify

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
)

func MockEndpoints(mockPlaylistFilePath string) (*httptest.Server, *httptest.Server) {
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

		f, err := os.ReadFile(mockPlaylistFilePath)
		if err != nil {
			panic(err)
		}
		w.Write(f)
	}))

	return tsToken, tsPlaylist
}

func MockPlaylists(mockPlaylistsFilePath string) []Playlist {
	file, err := os.ReadFile(mockPlaylistsFilePath)
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
