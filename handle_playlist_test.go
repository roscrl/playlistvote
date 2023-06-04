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
	req := httptest.NewRequest(http.MethodGet, "/playlist/top?after=6C6IZB8oY1VgSSjerp9edG", nil)
	req.Header.Set("Accept", views.TurboStreamMIME)

	server.ServeHTTP(w, req)
	is.Equal(w.Result().StatusCode, http.StatusOK)
}
