package core

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sort"

	"app/core/domain"
	"app/core/middleware"
	"app/core/rlog"
	"app/core/session"
	"app/core/views"
)

func (s *Server) handlePlaylistUpVote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := rlog.L(r.Context())

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

		sessionID, err := r.Cookie(middleware.CookieSessionName)
		if err != nil {
			log.ErrorCtx(r.Context(), "error getting session cookie", "err", err)
			http.Error(w, "error", http.StatusInternalServerError)

			return
		}

		s.UpvoteBroadcasterService.Notify(&domain.PlaylistIDUpvoteMessage{
			PlaylistID: domain.PlaylistID(playlistID),
			Upvotes:    upvotes,
			UpvotedBy:  domain.SessionID(sessionID.Value),
		})

		s.Views.Stream(w, "playlist/_upvote_success.stream.tmpl", map[string]any{
			"playlist_id": playlistID,
			"upvotes":     upvotes,
		})
	}
}

func (s *Server) handlePlaylistsUpvotesStream() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := rlog.L(r.Context())

		log.InfoCtx(r.Context(), "subscribing to playlists upvotes stream")

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		sessionID, err := r.Cookie(middleware.CookieSessionName)
		if err != nil {
			log.ErrorCtx(r.Context(), "error getting session cookie", "err", err)
			http.Error(w, "error", http.StatusInternalServerError)

			return
		}

		log.InfoCtx(r.Context(), "session id is set, getting session")

		userSession, err := s.SessionStore.GetSession(domain.SessionID(sessionID.Value))
		if err != nil {
			log.InfoCtx(r.Context(), "no user session exists in memory, adding new session with channel to listen to relevant playlist id upvotes")

			newSession := &session.User{
				ID:                                 domain.SessionID(sessionID.Value),
				PlaylistIDUpvoteChannel:            make(chan *domain.PlaylistIDUpvoteMessage),
				PlaylistIDsToListenToUpvoteChanges: []domain.PlaylistID{},
			}

			s.SessionStore.AddSession(newSession)
			userSession = newSession
		}

		{
			flusher, _ := w.(http.Flusher)
			seq := 0

			log.InfoCtx(r.Context(), "sending connected event")

			err := views.SSEMessage(w, seq, "connected", "connected")
			if err != nil {
				log.ErrorCtx(r.Context(), "error sending connected event", "err", err)
				http.Error(w, "error", http.StatusInternalServerError)

				return
			}

			flusher.Flush()
			seq++

			log.InfoCtx(r.Context(), "listening for changes on users playlist id upvote channel")

			for {
				select {
				case <-r.Context().Done():
					log.InfoCtx(r.Context(), "streaming ended due to end of request listener")

					return
				case playlistIDUpvoteMessage := <-userSession.PlaylistIDUpvoteChannel:
					if playlistIDUpvoteMessage.UpvotedBy == domain.SessionID(sessionID.Value) {
						log.InfoCtx(r.Context(), "not sending self upvote turbo stream", "playlist_id", playlistIDUpvoteMessage.PlaylistID, "upvotes", playlistIDUpvoteMessage.Upvotes)

						continue
					}

					buf := new(bytes.Buffer)
					s.Views.Render(buf, "playlist/_upvote.stream.tmpl", map[string]any{
						"playlist_id": playlistIDUpvoteMessage.PlaylistID,
						"upvotes":     playlistIDUpvoteMessage.Upvotes,
					})

					log.InfoCtx(r.Context(), "sending new upvote turbo stream to listening user", "playlist_id", playlistIDUpvoteMessage.PlaylistID, "upvotes", playlistIDUpvoteMessage.Upvotes)

					err := views.SSEMessage(w, seq, "message", buf.String())
					if err != nil {
						log.ErrorCtx(r.Context(), "error sending upvote turbo stream", "err", err)
						http.Error(w, "error", http.StatusInternalServerError)

						return
					}

					flusher.Flush()
					seq++
				}
			}
		}
	}
}

func (s *Server) handlePlaylistUpvotesSubscribe() http.HandlerFunc {
	const MaxPlaylistSubscriptions = 30

	type PlaylistUpvoteSubscribeRequest struct {
		PlaylistIDs []string `json:"playlist_ids"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log := rlog.L(r.Context())

		var playlistSubscribeRequest PlaylistUpvoteSubscribeRequest
		if err := json.NewDecoder(r.Body).Decode(&playlistSubscribeRequest); err != nil {
			log.ErrorCtx(r.Context(), "error decoding playlist subscribe request", "err", err)
			http.Error(w, "error", http.StatusInternalServerError)

			return
		}

		log.InfoCtx(r.Context(), "user would like to subscribe to upvote events on playlists", "amount", len(playlistSubscribeRequest.PlaylistIDs))

		sessionID, err := r.Cookie(middleware.CookieSessionName)
		if err != nil {
			log.ErrorCtx(r.Context(), "error getting session cookie", "err", err)
			http.Error(w, "error", http.StatusInternalServerError)

			return
		}

		log.InfoCtx(r.Context(), "got session id")

		userSession, err := s.SessionStore.GetSession(domain.SessionID(sessionID.Value))
		if err != nil {
			log.ErrorCtx(r.Context(), "error getting session", "err", err)
			http.Error(w, "error", http.StatusInternalServerError)

			return
		}

		log.InfoCtx(r.Context(), "got users session, subscribing to upvote changes", "playlist_ids", playlistSubscribeRequest.PlaylistIDs)

		userSession.Lock.Lock()
		defer userSession.Lock.Unlock()

		for _, playlistID := range playlistSubscribeRequest.PlaylistIDs {
			for i, id := range userSession.PlaylistIDsToListenToUpvoteChanges {
				if id == domain.PlaylistID(playlistID) {
					userSession.PlaylistIDsToListenToUpvoteChanges = append(userSession.PlaylistIDsToListenToUpvoteChanges[:i], userSession.PlaylistIDsToListenToUpvoteChanges[i+1:]...)

					break
				}
			}
		}

		for _, playlistID := range playlistSubscribeRequest.PlaylistIDs {
			userSession.PlaylistIDsToListenToUpvoteChanges = append(userSession.PlaylistIDsToListenToUpvoteChanges, domain.PlaylistID(playlistID))
		}

		if len(userSession.PlaylistIDsToListenToUpvoteChanges) > MaxPlaylistSubscriptions {
			userSession.PlaylistIDsToListenToUpvoteChanges = userSession.PlaylistIDsToListenToUpvoteChanges[len(userSession.PlaylistIDsToListenToUpvoteChanges)-MaxPlaylistSubscriptions:]
		}

		sort.Slice(userSession.PlaylistIDsToListenToUpvoteChanges, func(i, j int) bool {
			return userSession.PlaylistIDsToListenToUpvoteChanges[i] < userSession.PlaylistIDsToListenToUpvoteChanges[j]
		})

		log.InfoCtx(r.Context(), "user is now subscribed to upvote changes on given playlists", "amount", len(userSession.PlaylistIDsToListenToUpvoteChanges))
	}
}
