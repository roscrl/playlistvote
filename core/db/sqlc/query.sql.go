// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: query.sql

package sqlc

import (
	"context"
	"database/sql"
)

const addPlaylist = `-- name: AddPlaylist :execresult
INSERT INTO playlists (id, upvotes)
VALUES (?, ?)
`

type AddPlaylistParams struct {
	ID      string
	Upvotes int64
}

func (q *Queries) AddPlaylist(ctx context.Context, arg AddPlaylistParams) (sql.Result, error) {
	return q.db.ExecContext(ctx, addPlaylist, arg.ID, arg.Upvotes)
}

const getNewPlaylists = `-- name: GetNewPlaylists :many

SELECT id, upvotes
FROM playlists
ORDER BY added_at DESC
LIMIT ?
`

type GetNewPlaylistsRow struct {
	ID      string
	Upvotes int64
}

// ! Manually added to db/sqlc/query.sql.manual.go due to not working in sqlc
// ! name: GetNextTopPlaylists :many
func (q *Queries) GetNewPlaylists(ctx context.Context, limit int64) ([]GetNewPlaylistsRow, error) {
	rows, err := q.db.QueryContext(ctx, getNewPlaylists, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetNewPlaylistsRow
	for rows.Next() {
		var i GetNewPlaylistsRow
		if err := rows.Scan(&i.ID, &i.Upvotes); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPlaylistUpvotes = `-- name: GetPlaylistUpvotes :one
SELECT upvotes
FROM playlists
WHERE id = ?
`

func (q *Queries) GetPlaylistUpvotes(ctx context.Context, id string) (int64, error) {
	row := q.db.QueryRowContext(ctx, getPlaylistUpvotes, id)
	var upvotes int64
	err := row.Scan(&upvotes)
	return upvotes, err
}

const getTopPlaylists = `-- name: GetTopPlaylists :many
SELECT id, upvotes
FROM playlists
ORDER BY upvotes DESC, id DESC
LIMIT ?
`

type GetTopPlaylistsRow struct {
	ID      string
	Upvotes int64
}

func (q *Queries) GetTopPlaylists(ctx context.Context, limit int64) ([]GetTopPlaylistsRow, error) {
	rows, err := q.db.QueryContext(ctx, getTopPlaylists, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTopPlaylistsRow
	for rows.Next() {
		var i GetTopPlaylistsRow
		if err := rows.Scan(&i.ID, &i.Upvotes); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const incrementPlaylistUpvotes = `-- name: IncrementPlaylistUpvotes :one

UPDATE playlists
SET upvotes = upvotes + 1
WHERE id = ?
RETURNING upvotes
`

// ! TODO
// ! Manually added to db/sqlc/query.sql.manual.go due to not working in sqlc
// ! name: GetNextNewPlaylists :many
func (q *Queries) IncrementPlaylistUpvotes(ctx context.Context, id string) (int64, error) {
	row := q.db.QueryRowContext(ctx, incrementPlaylistUpvotes, id)
	var upvotes int64
	err := row.Scan(&upvotes)
	return upvotes, err
}

const playlistExists = `-- name: PlaylistExists :one
SELECT EXISTS(SELECT 1 FROM playlists WHERE id = ?)
`

func (q *Queries) PlaylistExists(ctx context.Context, id string) (int64, error) {
	row := q.db.QueryRowContext(ctx, playlistExists, id)
	var column_1 int64
	err := row.Scan(&column_1)
	return column_1, err
}
