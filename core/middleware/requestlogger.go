package middleware

import (
	"context"
	"net/http"
	"os"

	"app/core/contextkey"
	"app/core/rlog"
	"golang.org/x/exp/slog"
)

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		textHandler := slog.NewTextHandler(os.Stdout, nil)
		requestContextHandler := rlog.ContextRequestHandler{Handler: textHandler}

		logger := slog.New(requestContextHandler)

		ctx = context.WithValue(ctx, contextkey.RequestLogger{}, logger)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
