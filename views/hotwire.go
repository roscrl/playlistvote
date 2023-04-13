package views

import (
	"net/http"
	"strings"
)

const (
	TurboStreamMIME = "text/vnd.turbo-stream.html"
)

func TurboStreamRequest(req *http.Request) bool {
	if strings.Contains(req.Header.Get("Accept"), TurboStreamMIME) {
		return true
	}
	return false
}
