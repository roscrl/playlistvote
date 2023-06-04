package main

import (
	"context"
	"net/http"
	"strings"
)

func (s *Server) handleRouting(routes []route) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var allow []string
		for _, route := range routes {
			matches := route.regex.FindStringSubmatch(req.URL.Path)
			if len(matches) > 0 {
				if req.Method != route.method {
					allow = append(allow, route.method)

					continue
				}
				ctx := context.WithValue(req.Context(), ctxKey{}, matches[1:])
				route.handler(w, req.WithContext(ctx))

				return
			}
		}
		if len(allow) > 0 {
			w.Header().Set("Allow", strings.Join(allow, ", "))
			http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)

			return
		}
		http.NotFound(w, req)
	})
}
