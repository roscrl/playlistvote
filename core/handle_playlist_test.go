package core

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"app/config"
	"app/core/views"
	"github.com/matryer/is"
)

func TestHandlePlaylist(t *testing.T) {
	is, server := is.New(t), NewServer(config.MockConfig())

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept", views.TurboStreamMIME)

	server.ServeHTTP(w, r)
	is.Equal(w.Result().StatusCode, http.StatusOK)
}
