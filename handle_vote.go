package main

import (
	"log"
	"net/http"
)

func (s *Server) handleUpVote() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		playlistID := getField(req, 0)
		log.Printf("Upvoting playlist %s", playlistID) // TODO hit db
	}
}
