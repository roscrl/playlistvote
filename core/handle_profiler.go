package core

import (
	"net/http"
	"net/http/pprof"
)

func (s *Server) handleIndex() http.HandlerFunc {
	return pprof.Index
}

func (s *Server) handleAllocs() http.HandlerFunc {
	return pprof.Handler("allocs").ServeHTTP
}

func (s *Server) handleBlock() http.HandlerFunc {
	return pprof.Handler("block").ServeHTTP
}

func (s *Server) handleCmdline() http.HandlerFunc {
	return pprof.Cmdline
}

func (s *Server) handleGoroutine() http.HandlerFunc {
	return pprof.Handler("goroutine").ServeHTTP
}

func (s *Server) handleHeap() http.HandlerFunc {
	return pprof.Handler("heap").ServeHTTP
}

func (s *Server) handleMutex() http.HandlerFunc {
	return pprof.Handler("mutex").ServeHTTP
}

func (s *Server) handleProfile() http.HandlerFunc {
	return pprof.Profile
}

func (s *Server) handleThreadcreate() http.HandlerFunc {
	return pprof.Handler("threadcreate").ServeHTTP
}

func (s *Server) handleSymbol() http.HandlerFunc {
	return pprof.Symbol
}

func (s *Server) handleTrace() http.HandlerFunc {
	return pprof.Trace
}
