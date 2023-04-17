package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"app/config"

	"github.com/matryer/is"
)

func TestHandleHome(t *testing.T) {
	is, server := is.New(t), NewServer(config.DevConfig(true))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	server.ServeHTTP(w, req)    // integration test like (middlewares included)
	server.handleHome()(w, req) // unit test like (no middlewares)

	is.Equal(w.Result().StatusCode, http.StatusOK)
}
