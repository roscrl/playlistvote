package main

import (
	"net/http"
)

func (s *Server) handleAssets() http.HandlerFunc {
	return http.FileServer(http.Dir("./views/assets/dist/")).ServeHTTP
}
