package views

import (
	"net/http"
	"strings"
)

const (
	TurboStreamMIME = "text/vnd.turbo-stream.html"
)

func TurboStreamRequest(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), TurboStreamMIME)
}
