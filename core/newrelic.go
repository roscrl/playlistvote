package core

import (
	"context"
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

func startSegment(r *http.Request, name string) *newrelic.Segment {
	return newrelic.FromContext(r.Context()).StartSegment(name)
}

func noticeError(ctx context.Context, err error) {
	newrelic.FromContext(ctx).NoticeError(err)
}

func instrumentRoutes(routes []route, apm *newrelic.Application) {
	for i := range routes {
		_, handler := newrelic.WrapHandleFunc(apm, routes[i].regex.String(), routes[i].handler)
		routes[i].handler = handler
	}
}
