package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func requestDurationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, AssetBaseRoute) {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		next.ServeHTTP(w, r)

		elapsed := time.Since(start)
		log.Printf("%s %s took %s", r.Method, r.URL.Path, elapsed)
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovery := recover(); recovery != nil {
				var err error
				switch t := recovery.(type) {
				case string:
					err = fmt.Errorf(t)
				case error:
					err = t
				default:
					err = fmt.Errorf("unknown panic: %v", t)
				}
				log.Printf("panic: %s", err)
				noticeError(r, err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
