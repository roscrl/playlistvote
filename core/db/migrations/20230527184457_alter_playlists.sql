BEGIN TRANSACTION;

CREATE TABLE playlists_new
(
    id         TEXT PRIMARY KEY,
    upvotes    INTEGER NOT NULL DEFAULT 1,
    added_at   INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
) STRICT;

INSERT INTO playlists_new (id, upvotes)
SELECT id, upvotes
FROM playlists;

DROP TABLE playlists;

ALTER TABLE playlists_new
    RENAME TO playlists;

CREATE INDEX idx_playlist_added_at ON playlists(added_at);
CREATE INDEX idx_playlist_upvotes ON playlists(upvotes DESC);
CREATE INDEX idx_playlist_upvotes_id ON playlists(upvotes DESC, id ASC);

PRAGMA user_version = 4;

COMMIT;
