package core

import (
	"net/http"
)

func (s *Server) handleUp() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}
