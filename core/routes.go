package core

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
	RouteAssetBase = "/assets"
	RouteAsset     = "/assets/(.*)"

	RouteHome                   = "/"
	RoutePlaylistBase           = "/playlist"
	RoutePlaylistCreate         = "/playlist"
	RoutePlaylistsPaginationTop = "/playlist/top(.*)"
	RoutePlaylistView           = "/playlist/(.*)"
	RoutePlaylistUpvote         = "/playlist/(.*)/upvote"

	RouteUp = "/up"

	RouteProfileBaseRoute    = "/debug/pprof"
	RouteProfileAllocs       = "/debug/allocs"
	RouteProfileBlock        = "/debug/block"
	RouteProfileCmdline      = "/debug/cmdline"
	RouteProfileGoroutine    = "/debug/goroutine"
	RouteProfileHeap         = "/debug/heap"
	RouteProfileMutex        = "/debug/mutex"
	RouteProfileProfile      = "/debug/profile"
	RouteProfileThreadcreate = "/debug/threadcreate"
	RouteProfileSymbol       = "/debug/symbol"
	RouteProfileTrace        = "/debug/trace"
)

type contextKeyFields struct{}

func (s *Server) routes() http.Handler {
	newRoute := func(method, pattern string, handler http.HandlerFunc) route {
		return route{method, regexp.MustCompile("^" + pattern + "$"), handler}
	}

	routes := []route{
		newRoute(http.MethodGet, RouteAsset, http.StripPrefix(RouteAssetBase+"/", s.handleAssets()).ServeHTTP),
		newRoute(http.MethodGet, RouteHome, s.handleHome()),

		newRoute(http.MethodPost, RoutePlaylistCreate, s.handlePlaylistCreate()),
		newRoute(http.MethodGet, RoutePlaylistsPaginationTop, s.handlePlaylistsPaginationTop()),
		newRoute(http.MethodGet, RoutePlaylistView, s.handlePlaylistView()),
		newRoute(http.MethodPost, RoutePlaylistUpvote, s.handlePlaylistUpVote()),

		newRoute(http.MethodGet, RouteUp, s.handleUp()),
	}

	pprofRoutes := map[string]http.HandlerFunc{
		RouteProfileBaseRoute:    s.handleIndex(),
		RouteProfileAllocs:       s.handleAllocs(),
		RouteProfileBlock:        s.handleBlock(),
		RouteProfileCmdline:      s.handleCmdline(),
		RouteProfileGoroutine:    s.handleGoroutine(),
		RouteProfileHeap:         s.handleHeap(),
		RouteProfileMutex:        s.handleMutex(),
		RouteProfileProfile:      s.handleProfile(),
		RouteProfileThreadcreate: s.handleThreadcreate(),
		RouteProfileSymbol:       s.handleSymbol(),
		RouteProfileTrace:        s.handleTrace(),
	}

	for path, handler := range pprofRoutes {
		routes = append(routes, newRoute(http.MethodGet, path, basicAuthAdminRouteMiddleware(handler, s.Cfg.BasicDebugAuthUsername, s.Cfg.BasicDebugAuthPassword)))
	}

	instrumentRoutes(routes, s.APM)

	routerEntry := s.routing(routes)

	return recoveryMiddleware(requestLogger(requestID(requestDurationMiddleware(routerEntry))))
}

func getField(r *http.Request, index int) string {
	fields := r.Context().Value(contextKeyFields{}).([]string)

	return fields[index]
}
