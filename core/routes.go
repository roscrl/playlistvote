package core

import (
	"net/http"
	"regexp"

	"app/core/middleware"
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
	RouteHomeNew                = "/new"
	RoutePlaylistBase           = "/playlist"
	RoutePlaylistCreate         = "/playlist"
	RoutePlaylistsPaginationTop = "/playlists/top(.*)"
	RoutePlaylistsPaginationNew = "/playlists/new(.*)"

	RoutePlaylistView   = "/playlist/(.*)"
	RoutePlaylistUpvote = "/playlist/(.*)/upvote"

	RoutePlaylistsUpvotesSubscribe = "/playlists/subscribe/upvotes"
	RoutePlaylistsUpvotesStream    = "/playlists/stream/upvotes"

	RouteUp                  = "/up"
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
		newRoute(http.MethodGet, RouteHome, s.handleHomeTop()),
		newRoute(http.MethodGet, RouteHomeNew, s.handleHomeNew()),

		newRoute(http.MethodPost, RoutePlaylistsUpvotesSubscribe, s.handlePlaylistUpvotesSubscribe()),
		newRoute(http.MethodGet, RoutePlaylistsUpvotesStream, s.handlePlaylistsUpvotesStream()),

		newRoute(http.MethodPost, RoutePlaylistCreate, s.handlePlaylistCreate()),

		newRoute(http.MethodGet, RoutePlaylistsPaginationTop, s.handlePlaylistsPaginationTop()),
		newRoute(http.MethodGet, RoutePlaylistsPaginationNew, s.handlePlaylistsPaginationNew()),

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
		routes = append(routes, newRoute(http.MethodGet, path, middleware.BasicAuthAdmin(handler, s.Cfg.BasicDebugAuthUsername, s.Cfg.BasicDebugAuthPassword)))
	}

	instrumentRoutes(routes, s.APM)

	routerEntry := s.routing(routes)

	return middleware.RequestLogger(
		middleware.RequestPathToContext(
			middleware.RequestID(
				middleware.Recovery(
					middleware.CookieSession(
						middleware.RequestDuration(
							routerEntry, RouteAssetBase,
						),
					), noticeError,
				),
			), RouteAssetBase,
		),
	)
}

func getField(r *http.Request, index int) string {
	fields := r.Context().Value(contextKeyFields{}).([]string)

	return fields[index]
}
