package core

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"app/config"
	"app/core/db"
	"app/core/views"
	_ "app/testing"
	"github.com/matryer/is"
)

func TestHandleTopPlaylistsHome(t *testing.T) {
	is, server := is.New(t), NewServer(config.MockConfig())

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	server.ServeHTTP(w, r) // integration test like (middlewares included)
	is.Equal(w.Result().StatusCode, http.StatusOK)

	server.handleHomeTop()(w, r) // unit test like (no middlewares)
	is.Equal(w.Result().StatusCode, http.StatusOK)
}

func TestHandleTopPlaylistsAfterCursor(t *testing.T) {
	mc := config.MockConfig()
	mc.SqliteDBPath = ":memory:"

	is, server := is.New(t), NewServer(mc)

	db.RunMigrations(server.DB, db.MigrationsPath)
	db.SeedTestData(server.DB)

	w := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodGet, "/playlists/top?after=6-10", nil)
	req.Header.Set("Accept", views.TurboStreamMIME)

	server.ServeHTTP(w, req)

	server.handlePlaylistsPaginationTop()(w, req)
	is.Equal(w.Result().StatusCode, http.StatusOK)
}
