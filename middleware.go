package main

import (
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
