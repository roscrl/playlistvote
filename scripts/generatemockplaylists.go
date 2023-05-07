package main

import (
	"context"
	"log"
	"time"

	"app/config"
	"app/db"
	"app/db/sqlc"
	"app/services/spotify/mock"
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
			AddedAt: time.Now().Unix(),
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}
