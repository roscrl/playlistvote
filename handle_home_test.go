package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"app/config"

	"github.com/matryer/is"
)

// TODO rm in favor of browsertests/??
func TestHandleHome(t *testing.T) {
	is, server := is.New(t), NewServer(config.DevConfig(true))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	server.ServeHTTP(w, req) // integration test like (middlewares included)
	is.Equal(w.Result().StatusCode, http.StatusOK)

	server.handleHome()(w, req) // unit test like (no middlewares)
	is.Equal(w.Result().StatusCode, http.StatusOK)
}
