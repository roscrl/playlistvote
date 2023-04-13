package main

import (
	"net/http"

	"github.com/newrelic/go-agent/v3/newrelic"
)

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

func startSegment(req *http.Request, name string) *newrelic.Segment {
	return newrelic.FromContext(req.Context()).StartSegment(name)
}
