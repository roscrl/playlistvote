package views

import (
	"net/http"
	"strings"
)

const (
	TurboStreamMIME = "text/vnd.turbo-stream.html"
)

func TurboStreamRequest(req *http.Request) bool {
	return strings.Contains(req.Header.Get("Accept"), TurboStreamMIME)
}
