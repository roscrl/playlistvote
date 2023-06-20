package views

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	TurboStreamMIME = "text/vnd.turbo-stream.html"
)

func TurboStreamRequest(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), TurboStreamMIME)
}

// SSEMessage constructs an SSE compatible message with a sequence, and
// line breaks from the output of a template
//
// This looks something like this:
//
//	event: message
//	id: 6
//	data: <turbo-stream action="replace" target="load">
//	data:     <template>
//	data:         <span id="load">04:20:13: 1.9</span>
//	data:     </template>
//	data: </turbo-stream>
func SSEMessage(w io.Writer, id int, event, message string) error {
	_, err := fmt.Fprintf(w, "event: %s\nid: %d\n", event, id)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(bytes.NewBufferString(message))
	for scanner.Scan() {
		_, err = fmt.Fprintf(w, "data: %s\n", scanner.Text())
		if err != nil {
			return err
		}
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	_, err = fmt.Fprint(w, "\n")
	if err != nil {
		return err
	}

	return nil
}
