package main

import (
	"net/http"

	"app/views"
)

func (s *Server) handleUpVote() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if !views.TurboStreamRequest(req) {
			http.Redirect(w, req, HomeRoute, http.StatusSeeOther)

			return
		}

		playlistID := getField(req, 0)

		seg := startSegment(req, "PlaylistUpvote")

		upvotes, err := s.qry.IncrementPlaylistUpvotes(req.Context(), playlistID)
		if err != nil {
			s.views.Render(w, "error.tmpl", map[string]any{
				"error": "Something went wrong on our side trying to upvote this playlist. Please try again later.",
			})

			return
		}

		seg.End()

		s.views.Stream(w, "playlist/_upvote_success.stream.tmpl", map[string]any{
			"playlist_id": playlistID,
			"upvotes":     upvotes,
		})
	}
}
