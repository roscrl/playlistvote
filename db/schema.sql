CREATE TABLE playlists
(
    id         TEXT PRIMARY KEY,
    upvotes    INTEGER NOT NULL DEFAULT 1,
    added_at   INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
) STRICT;