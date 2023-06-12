package core

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
)

// Key to use when setting the request ID.
type ctxKeyRequestID int

// RequestIDKey is the key that holds the unique request ID in a request context.
const RequestIDKey ctxKeyRequestID = 0

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		bytes := make([]byte, 8) //nolint:gomnd // 8 bytes = 64 bits = 16 hex characters
		if _, err := rand.Read(bytes); err != nil {
			bytes = []byte("00000000")
		}

		requestID := fmt.Sprintf("%X", bytes)

		ctx = context.WithValue(ctx, RequestIDKey, requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
