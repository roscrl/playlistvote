package sqlc

import (
	"context"
	"testing"

	"app/core/db"
	"github.com/matryer/is"
)

func TestNextTopPlaylist(t *testing.T) {
	is := is.New(t)

	sqliteDb := db.New(":memory:")
	db.RunMigrations(sqliteDb, db.MigrationsPath)
	qry := New(sqliteDb)

	table := []struct {
		id      string
		upvotes int64
	}{
		{"6", 4},
		{"5", 3},
		{"4", 2},
		{"3", 2},
		{"2", 2},
		{"1", 1},
		{"0", 1},
	}

	for _, tt := range table {
		_, err := qry.AddPlaylist(context.Background(), AddPlaylistParams{
			ID:      tt.id,
			Upvotes: tt.upvotes,
		})
		is.NoErr(err)
	}

	t.Run("return playlists pagination flow and is sorted by id and deduplicated if id cursor given", func(t *testing.T) {
		topPlaylist, err := qry.GetTopPlaylists(context.Background(), 1)
		is.NoErr(err)

		is.Equal(len(topPlaylist), 1)

		is.Equal(topPlaylist[0].ID, "6")
		is.Equal(topPlaylist[0].Upvotes, int64(4))

		topPlaylistsPagination, err := qry.NextTopPlaylists(context.Background(), NextTopPlaylistsParams{
			ID:      "6",
			Upvotes: 4,
			Limit:   2,
		})
		is.NoErr(err)

		is.Equal(len(topPlaylistsPagination), 2)

		is.Equal(topPlaylistsPagination[0].ID, "5")
		is.Equal(topPlaylistsPagination[0].Upvotes, int64(3))

		is.Equal(topPlaylistsPagination[1].ID, "4")
		is.Equal(topPlaylistsPagination[1].Upvotes, int64(2))

		topPlaylistsPagination, err = qry.NextTopPlaylists(context.Background(), NextTopPlaylistsParams{
			ID:      "4",
			Upvotes: 2,
			Limit:   2,
		})
		is.NoErr(err)

		is.Equal(len(topPlaylistsPagination), 2)

		is.Equal(topPlaylistsPagination[0].ID, "3")
		is.Equal(topPlaylistsPagination[0].Upvotes, int64(2))

		is.Equal(topPlaylistsPagination[1].ID, "2")
		is.Equal(topPlaylistsPagination[1].Upvotes, int64(2))

		topPlaylistsPagination, err = qry.NextTopPlaylists(context.Background(), NextTopPlaylistsParams{
			ID:      "2",
			Upvotes: 2,
			Limit:   2,
		})
		is.NoErr(err)

		is.Equal(len(topPlaylistsPagination), 2)

		is.Equal(topPlaylistsPagination[0].ID, "1")
		is.Equal(topPlaylistsPagination[0].Upvotes, int64(1))

		is.Equal(topPlaylistsPagination[1].ID, "0")
		is.Equal(topPlaylistsPagination[1].Upvotes, int64(1))

		topPlaylistsPagination, err = qry.NextTopPlaylists(context.Background(), NextTopPlaylistsParams{
			ID:      "0",
			Upvotes: 1,
			Limit:   2,
		})
		is.NoErr(err)

		is.Equal(len(topPlaylistsPagination), 0)
	})

	t.Run("return playlists even if id does not exist and is sorted by id but upvotes is given", func(t *testing.T) {
		topPlaylistsPagination, err := qry.NextTopPlaylists(context.Background(), NextTopPlaylistsParams{
			ID:      "does-not-exist",
			Upvotes: 2,
			Limit:   2,
		})
		is.NoErr(err)

		is.Equal(len(topPlaylistsPagination), 2)
		is.Equal(topPlaylistsPagination[0].ID, "4")
		is.Equal(topPlaylistsPagination[0].Upvotes, int64(2))

		is.Equal(topPlaylistsPagination[1].ID, "3")
		is.Equal(topPlaylistsPagination[1].Upvotes, int64(2))
	})
}
