CREATE TABLE playlists (
    id       TEXT PRIMARY KEY,
    upvotes  INTEGER NOT NULL,
    added_at INTEGER NOT NULL
); -- playlists is a strict table but sqlc complains when adding

