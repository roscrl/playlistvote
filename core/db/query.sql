-- name: AddPlaylist :execresult
INSERT INTO playlists (id, upvotes)
VALUES (?, ?);

-- name: GetPlaylistUpvotes :one
SELECT upvotes
FROM playlists
WHERE id = ?;

-- name: PlaylistExists :one
SELECT EXISTS(SELECT 1 FROM playlists WHERE id = ?);

-- name: GetTopPlaylists :many
SELECT id, upvotes, added_at
FROM playlists
ORDER BY upvotes DESC, id DESC
LIMIT ?;

--! Manually added GetNextTopPlaylists to db/sqlc/query.sql.manual.go due to not working in sqlc

-- name: GetNewPlaylists :many
SELECT id, upvotes, added_at
FROM playlists
ORDER BY added_at DESC, id DESC
LIMIT ?;

--! Manually added GetNextNewPlaylists to db/sqlc/query.sql.manual.go due to not working in sqlc

-- name: IncrementPlaylistUpvotes :one
UPDATE playlists
SET upvotes = upvotes + 1
WHERE id = ?
RETURNING upvotes;
