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
SELECT id, upvotes
FROM playlists
ORDER BY upvotes DESC, id ASC
LIMIT ?;

--! Manually added to db/sqlc/query.sql.manual.go due to not working in sqlc
--! name: GetNextTopPlaylists :many
SELECT id, upvotes
FROM (SELECT id, upvotes
      FROM playlists
      WHERE upvotes <= ?1
        AND NOT (upvotes = ?1 AND id >= ?)
      ORDER BY upvotes DESC, id ASC)
LIMIT ?;

-- name: GetNewPlaylists :many
SELECT id, upvotes
FROM playlists
ORDER BY added_at DESC
LIMIT ?;

--! Manually added to db/sqlc/query.sql.manual.go due to not working in sqlc
--! name: GetNextNewPlaylists :many
-- SELECT id, upvotes
-- FROM (SELECT id, upvotes
--       FROM playlists
--       WHERE added_at <= ?
--         AND NOT (added_at = ? AND id >= ?)
--       ORDER BY upvotes DESC, id DESC)
-- LIMIT ?;

-- name: IncrementPlaylistUpvotes :one
UPDATE playlists
SET upvotes = upvotes + 1
WHERE id = ?
RETURNING upvotes;
