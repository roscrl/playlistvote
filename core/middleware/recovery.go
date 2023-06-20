package middleware

import (
	"context"
	"fmt"
	"net/http"

	"app/core/contextkey"
	"app/core/rlog"
)

func Recovery(next http.Handler, noticeError func(context.Context, error)) http.Handler {
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

				var requestID string
				if rid, ok := r.Context().Value(contextkey.RequestID{}).(string); ok {
					requestID = rid
				}

				http.Error(w, "internal server error "+requestID, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
