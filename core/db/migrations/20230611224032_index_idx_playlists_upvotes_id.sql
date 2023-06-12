DROP INDEX idx_playlist_upvotes_id;

CREATE INDEX idx_playlist_upvotes_id ON playlists(upvotes DESC, id DESC);

PRAGMA USER_VERSION = 5;