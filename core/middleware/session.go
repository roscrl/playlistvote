package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"app/core/contextkey"
	"app/core/rlog"
)

const CookieSessionName = "session_id"

func CookieSession(next http.Handler) http.Handler {
	const CookieBytesLength = 20

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := rlog.L(r.Context())

		var sessionIDCookie *http.Cookie

		sessionIDCookie, err := r.Cookie(CookieSessionName)
		if err != nil {
			log.InfoCtx(r.Context(), "session cookie does not exist, creating new session")

			randomBytes := make([]byte, CookieBytesLength)

			_, err := rand.Read(randomBytes)
			if err != nil {
				randomBytes = []byte("00000000")
			}

			sessionID := base64.URLEncoding.EncodeToString(randomBytes)

			sessionIDCookie = &http.Cookie{
				Name:  CookieSessionName,
				Value: sessionID,
			}

			http.SetCookie(w, sessionIDCookie)

			r = r.WithContext(context.WithValue(r.Context(), contextkey.SessionNew{}, true))
		}

		r = r.WithContext(context.WithValue(r.Context(), contextkey.SessionID{}, sessionIDCookie.Value))

		next.ServeHTTP(w, r)
	})
}
