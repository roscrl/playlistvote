package main

import (
	"context"
	"net/http"
	"regexp"
	"strings"
)

type route struct {
	method  string
	regex   *regexp.Regexp
	handler http.HandlerFunc
}

const (
	AssetRoute          = "/assets/(.*)"
	HomeRoute           = "/"
	PlaylistBaseRoute   = "/playlists"
	PlaylistCreateRoute = "/playlists"
	PlaylistViewRoute   = "/playlists/(.*)"
	PlaylistUpvoteRoute = "/playlists/(.*)/upvote"
)

type ctxKey struct{}

func (s *Server) routes() http.Handler {
	newRoute := func(method, pattern string, handler http.HandlerFunc) route {
		return route{method, regexp.MustCompile("^" + pattern + "$"), handler}
	}

	routes := []route{
		newRoute("GET", AssetRoute, http.StripPrefix("/assets/", s.handleAssets()).ServeHTTP),
		newRoute("GET", HomeRoute, s.handleHome()),
		newRoute("POST", PlaylistCreateRoute, s.handlePostPlaylist()),
		newRoute("GET", PlaylistViewRoute, s.handleGetPlaylist()),
		newRoute("POST", PlaylistUpvoteRoute, s.handleUpVote()),
	}

	instrumentRoutes(routes, s.apm)

	routerEntry := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var allow []string
		for _, route := range routes {
			matches := route.regex.FindStringSubmatch(r.URL.Path)
			if len(matches) > 0 {
				if r.Method != route.method {
					allow = append(allow, route.method)
					continue
				}
				ctx := context.WithValue(r.Context(), ctxKey{}, matches[1:])
				route.handler(w, r.WithContext(ctx))
				return
			}
		}
		if len(allow) > 0 {
			w.Header().Set("Allow", strings.Join(allow, ", "))
			http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
			return
		}
		http.NotFound(w, r)
	})

	return requestDurationMiddleware(routerEntry)
}

func getField(r *http.Request, index int) string {
	fields := r.Context().Value(ctxKey{}).([]string)
	return fields[index]
}

