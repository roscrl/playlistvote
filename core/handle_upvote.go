package core

import (
	"net/http"

	"app/core/views"
)

func (s *Server) handlePlaylistUpVote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !views.TurboStreamRequest(r) {
			http.Redirect(w, r, RouteHome, http.StatusSeeOther)

			return
		}

		playlistID := getField(r, 0)

		seg := startSegment(r, "PlaylistUpvote")

		upvotes, err := s.Qry.IncrementPlaylistUpvotes(r.Context(), playlistID)
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
