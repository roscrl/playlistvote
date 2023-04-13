-- name: AddPlaylist :execresult
INSERT INTO playlists (id, upvotes, added_at)
VALUES (?, ?, ?);

-- name: GetPlaylist :one
SELECT *
FROM playlists
WHERE id = ?;

-- name: PlaylistExists :one
SELECT EXISTS(SELECT 1 FROM playlists WHERE id = ?);

-- name: GetTop100PlaylistsByUpvotesPastWeek :many
SELECT *
FROM playlists
WHERE added_at >= strftime('%s', 'now', '-7 days') * 1000
ORDER BY upvotes DESC
LIMIT 100;

-- name: GetTop100PlaylistsByUpvotesPastMonth :many
SELECT *
FROM playlists
WHERE added_at >= strftime('%s', 'now', '-1 months') * 1000
ORDER BY upvotes DESC
LIMIT 100;

-- name: GetTop100PlaylistsByUpvotesPast3Months :many
SELECT *
FROM playlists
WHERE added_at >= strftime('%s', 'now', '-3 months') * 1000
ORDER BY upvotes DESC
LIMIT 100;

-- name: GetTop500PlaylistsByUpvotesAllTime :many
SELECT *
FROM playlists
ORDER BY upvotes DESC
LIMIT 500;

-- e.g upvotes for March 2023, replace ?1 with '2023-03'.
-- name: GetTop100PlaylistsByUpvotesInMonth :many
SELECT *
FROM playlists
WHERE strftime('%Y-%m', added_at / 1000, 'unixepoch') = ?1
ORDER BY upvotes DESC
LIMIT 100;

-- name: IncrementPlaylistUpvotes :execresult
UPDATE playlists
SET upvotes = upvotes + 1
WHERE id = ?;