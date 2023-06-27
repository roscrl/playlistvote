package sqlc

import "context"

const NextTopPlaylists = `-- name: NextTopPlaylists :many
SELECT id, upvotes, added_at
FROM (SELECT id, upvotes, added_at
      FROM playlists
      WHERE upvotes <= ?1
        AND NOT (upvotes = ?1 AND id >= ?)
      ORDER BY upvotes DESC, id DESC)
LIMIT ?;
`

type NextTopPlaylistsRow struct {
	ID      string
	Upvotes int64
	AddedAt int64
}

type NextTopPlaylistsParams struct {
	ID      string
	Upvotes int64
	Limit   int64
}

func (q *Queries) NextTopPlaylists(ctx context.Context, arg NextTopPlaylistsParams) ([]NextTopPlaylistsRow, error) {
	rows, err := q.db.QueryContext(ctx, NextTopPlaylists, arg.Upvotes, arg.ID, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []NextTopPlaylistsRow

	for rows.Next() {
		var i NextTopPlaylistsRow
		if err := rows.Scan(&i.ID, &i.Upvotes, &i.AddedAt); err != nil {
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

const NextNewPlaylists = `-- name: NextNewPlaylists :many
SELECT id, upvotes, added_at
FROM (SELECT id, upvotes, added_at
      FROM playlists
      WHERE added_at <= ?1
        AND NOT (added_at = ?1 AND id >= ?)
      ORDER BY added_at DESC, id DESC)
LIMIT ?;
`

type NextNewPlaylistsRow struct {
	ID      string
	Upvotes int64
	AddedAt int64
}

type NextNewPlaylistsParams struct {
	ID      string
	AddedAt int64
	Limit   int64
}

func (q *Queries) NextNewPlaylists(ctx context.Context, arg NextNewPlaylistsParams) ([]NextNewPlaylistsRow, error) {
	rows, err := q.db.QueryContext(ctx, NextNewPlaylists, arg.AddedAt, arg.ID, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []NextNewPlaylistsRow

	for rows.Next() {
		var i NextNewPlaylistsRow
		if err := rows.Scan(&i.ID, &i.Upvotes, &i.AddedAt); err != nil {
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
