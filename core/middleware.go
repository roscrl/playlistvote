package core

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"app/core/rlog"
	"golang.org/x/exp/slog"
)

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		textHandler := slog.NewTextHandler(os.Stdout, nil)
		requestContextHandler := rlog.ContextRequestHandler{Handler: textHandler}

		logger := slog.New(requestContextHandler)

		ctx = context.WithValue(ctx, rlog.ContextKeyRequestLogger{}, logger)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		bytes := make([]byte, 8) //nolint:gomnd // 8 bytes = 64 bits = 16 hex characters
		if _, err := rand.Read(bytes); err != nil {
			bytes = []byte("00000000")
		}

		requestID := fmt.Sprintf("%X", bytes)

		ctx = context.WithValue(ctx, rlog.ContextKeyRequestID{}, requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovery := recover(); recovery != nil {
				var err error
				switch panicType := recovery.(type) {
				case string:
					err = fmt.Errorf(panicType)
				case error:
					err = panicType
				default:
					err = fmt.Errorf("unknown panic: %v", panicType)
				}

				log := rlog.L(r.Context())
				log.ErrorCtx(r.Context(), "panic", "err", err)

				noticeError(r.Context(), err)

				var requestId string
				if rid, ok := r.Context().Value(rlog.ContextKeyRequestID{}).(string); ok {
					requestId = rid
				}

				http.Error(w, "internal server error "+requestId, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func requestDurationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, RouteAssetBase) {
			next.ServeHTTP(w, r)

			return
		}

		start := time.Now()

		next.ServeHTTP(w, r)

		elapsed := time.Since(start)

		log := rlog.L(r.Context())
		log.InfoCtx(r.Context(), "request duration", "path", r.URL.Path, "took", elapsed)
	})
}

func basicAuthAdminRouteMiddleware(next http.HandlerFunc, username, password string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		givenUsername, givenPassword, ok := r.BasicAuth()

		if !ok || givenUsername != username || givenPassword != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)

			return
		}

		next.ServeHTTP(w, r)
	}
}
