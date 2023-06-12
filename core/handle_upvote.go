package core

import (
	"net/http"

	"app/core/views"
)

func (s *Server) handlePlaylistUpVote() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if !views.TurboStreamRequest(req) {
			http.Redirect(w, req, RouteHome, http.StatusSeeOther)

			return
		}

		playlistID := getField(req, 0)

		seg := startSegment(req, "PlaylistUpvote")

		upvotes, err := s.Qry.IncrementPlaylistUpvotes(req.Context(), playlistID)
		if err != nil {
			s.Views.Render(w, "error.tmpl", map[string]any{
				"error": "Something went wrong on our side trying to upvote this playlist. Please try again later.",
			})

			return
		}

		seg.End()

		s.Views.Stream(w, "playlist/_upvote_success.stream.tmpl", map[string]any{
			"playlist_id": playlistID,
			"upvotes":     upvotes,
		})
	}
}
