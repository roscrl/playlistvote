package main

import (
	"log"
	"net/http"

	"github.com/newrelic/go-agent/v3/newrelic"
)

func newAPM(environment, license string) *newrelic.Application {
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName("Playlist Vote "+environment),
		newrelic.ConfigLicense(license),
		newrelic.ConfigCodeLevelMetricsEnabled(true),
	)
	if err != nil {
		log.Fatal(err)
	}

	return app
}

func startSegment(req *http.Request, name string) *newrelic.Segment {
	return newrelic.FromContext(req.Context()).StartSegment(name)
}

func noticeError(req *http.Request, err error) {
	newrelic.FromContext(req.Context()).NoticeError(err)
}

func instrumentRoutes(routes []route, apm *newrelic.Application) {
	for i := range routes {
		_, handler := newrelic.WrapHandleFunc(apm, routes[i].regex.String(), routes[i].handler)
		routes[i].handler = handler
	}
}
