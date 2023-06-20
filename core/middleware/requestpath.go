package middleware

import (
	"context"
	"net/http"
	"strings"

	"app/core/contextkey"
)

func RequestPathToContext(next http.Handler, ignorePath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasPrefix(path, ignorePath) {
			next.ServeHTTP(w, r)

			return
		}

		ctx := context.WithValue(r.Context(), contextkey.RequestPath{}, path)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
