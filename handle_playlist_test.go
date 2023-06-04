package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"app/config"
	"app/views"
	"github.com/matryer/is"
)

func TestHandlePlaylist(t *testing.T) {
	is, server := is.New(t), NewServer(config.MockConfig())

	w := httptest.NewRecorder()
	// set turbo header to test that it is passed through
	req := httptest.NewRequest("GET", "/playlist/top?after=6C6IZB8oY1VgSSjerp9edG", nil)
	req.Header.Set("Accept", views.TurboStreamMIME)

	server.ServeHTTP(w, req) // integration test like (middlewares included)
	is.Equal(w.Result().StatusCode, http.StatusOK)

	// server.handleHome()(w, req) // unit test like (no middlewares)
	// is.Equal(w.Result().StatusCode, http.StatusOK) nocheckin
}
