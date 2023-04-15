package main

import (
	"app/views"
	"net/http"
)

func (s *Server) handleUpVote() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if !views.TurboStreamRequest(req) {
			http.Redirect(w, req, HomeRoute, http.StatusSeeOther)
			return
		}

		playlistID := getField(req, 0)

		_, err := s.qry.IncrementPlaylistUpvotes(req.Context(), playlistID)
		if err != nil {
			s.views.Render(w, "error.tmpl", map[string]any{})
			return
		}

		s.views.RenderStream(w, "playlist/_upvote_success.stream.tmpl", map[string]any{
			"playlist_id": playlistID,
		})
	}
}
