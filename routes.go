package main

import (
	"net/http"
	"regexp"
)

type route struct {
	method  string
	regex   *regexp.Regexp
	handler http.HandlerFunc
}

const (
	AssetBaseRoute = "/assets"
	AssetRoute     = "/assets/(.*)"

	HomeRoute           = "/"
	PlaylistBaseRoute   = "/playlist"
	PlaylistCreateRoute = "/playlist"
	PlaylistTopRoute    = "/playlist/top(.*)"
	PlaylistViewRoute   = "/playlist/(.*)"
	PlaylistUpvoteRoute = "/playlist/(.*)/upvote"

	ProfileBaseRoute      = "/debug/pprof"
	ProfileAllocsRoute    = "/debug/allocs"
	ProfileBlockRoute     = "/debug/block"
	ProfileCmdlineRoute   = "/debug/cmdline"
	ProfileGoroutineRoute = "/debug/goroutine"
	ProfileHeapRoute      = "/debug/heap"
	ProfileMutexRoute     = "/debug/mutex"
	ProfileProfileRoute   = "/debug/profile"
	ProfileThreadcreate   = "/debug/threadcreate"
	ProfileSymbolRoute    = "/debug/symbol"
	ProfileTraceRoute     = "/debug/trace"
)

type ctxKey struct{}

func (s *Server) routes() http.Handler {
	newRoute := func(method, pattern string, handler http.HandlerFunc) route {
		return route{method, regexp.MustCompile("^" + pattern + "$"), handler}
	}

	routes := []route{
		newRoute(http.MethodGet, AssetRoute, http.StripPrefix(AssetBaseRoute+"/", s.handleAssets()).ServeHTTP),
		newRoute(http.MethodGet, HomeRoute, s.handleHome()),

		newRoute(http.MethodPost, PlaylistCreateRoute, s.handlePostPlaylist()),
		newRoute(http.MethodGet, PlaylistTopRoute, s.handleTopPlaylistsAfterCursor()),
		newRoute(http.MethodGet, PlaylistViewRoute, s.handleGetPlaylist()),
		newRoute(http.MethodPost, PlaylistUpvoteRoute, s.handleUpVote()),
	}

	pprofRoutes := map[string]http.HandlerFunc{
		ProfileBaseRoute:      s.handleIndex(),
		ProfileAllocsRoute:    s.handleAllocs(),
		ProfileBlockRoute:     s.handleBlock(),
		ProfileCmdlineRoute:   s.handleCmdline(),
		ProfileGoroutineRoute: s.handleGoroutine(),
		ProfileHeapRoute:      s.handleHeap(),
		ProfileMutexRoute:     s.handleMutex(),
		ProfileProfileRoute:   s.handleProfile(),
		ProfileThreadcreate:   s.handleThreadcreate(),
		ProfileSymbolRoute:    s.handleSymbol(),
		ProfileTraceRoute:     s.handleTrace(),
	}

	for path, handler := range pprofRoutes {
		routes = append(routes, newRoute(http.MethodGet, path, basicAuthAdminRouteMiddleware(handler, s.cfg.BasicDebugAuthUsername, s.cfg.BasicDebugAuthPassword)))
	}

	instrumentRoutes(routes, s.apm)

	routerEntry := s.handleRouting(routes)

	return recoveryMiddleware(requestDurationMiddleware(routerEntry))
}

func getField(r *http.Request, index int) string {
	fields := r.Context().Value(ctxKey{}).([]string)

	return fields[index]
}
