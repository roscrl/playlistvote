//nolint:gosec
package main

import (
	"context"
	"log"

	"app/config"
	"app/core/db"
	"app/core/db/sqlc"
	"app/core/spotify/mock"
)

func main() {
	devCfg := config.DevConfig()
	playlists := mock.GenerateMockPlaylistsFile(devCfg.SpotifyClientID, devCfg.SpotifyClientSecret)

	mockCfg := config.MockConfig()
	database := db.New(mockCfg.SqliteDBPath)

	_, err := database.Exec("DELETE FROM " + db.PlaylistsTable)
	if err != nil {
		log.Fatal(err)
	}

	qry := sqlc.New(database)
	for _, playlist := range *playlists {
		_, err := qry.AddPlaylist(context.Background(), sqlc.AddPlaylistParams{
			ID:      playlist.ID,
			Upvotes: 1,
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}
