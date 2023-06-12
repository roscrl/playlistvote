package spotify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestClientToken(t *testing.T) {
	is := is.New(t)

	tokenCallEndpointCount := 0

	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		tokenCallEndpointCount++

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
		  "access_token": "TQBC_KDHhdQfPm_jLG0vxDlK_f0J1Vm57gIjG9uFJlWU8ySKySX_FhQ3re3iuReIF2No-Kz8fJNxoEaVdONxEFD0TkpZNsKBJCcHbtCaBPa0MigvRB0",
		  "token_type": "Bearer",
		  "expires_in": 3600}
      `))
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	s := Client{
		Client:       http.DefaultClient,
		ClientID:     "unused_client_id",
		ClientSecret: "unused_client_secret",

		TokenEndpoint:    ts.URL + "/token",
		PlaylistEndpoint: ts.URL + "/playlist",

		Now: time.Now,
	}

	s.StartTokenLifecycle(context.Background())

	is.Equal(tokenCallEndpointCount, 1)
}
