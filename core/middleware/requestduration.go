package middleware

import (
	"net/http"
	"strings"
	"time"

	"app/core/rlog"
)

func RequestDuration(next http.Handler, ignorePath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, ignorePath) {
			next.ServeHTTP(w, r)

			return
		}

		start := time.Now()

		next.ServeHTTP(w, r)

		elapsed := time.Since(start)

		log := rlog.L(r.Context())
		log.InfoCtx(r.Context(), "request duration", "took", elapsed)
	})
}
