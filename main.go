package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/exp/slog"

	"app/db"
	"app/db/sqlc"

	"github.com/newrelic/go-agent/v3/newrelic"

	"app/config"
	"app/services/spotify"
	"app/views"
)

type Server struct {
	cfg   *config.Server
	log   *slog.Logger
	db    *sql.DB
	qry   *sqlc.Queries
	views *views.Views

	spotify *spotify.Spotify

	apm    *newrelic.Application
	router http.Handler
}

func NewServer(cfg *config.Server) *Server {
	srv := &Server{}

	srv.cfg = cfg
	srv.log = slog.New(slog.NewTextHandler(os.Stdout))
	srv.db = newDB(cfg.SqliteDBPath)
	srv.qry = sqlc.New(srv.db)
	srv.views = views.New(srv.cfg.Env)

	setupServices(srv, cfg.Mocking)

	srv.apm = newApm(srv.cfg.Env.String(), srv.cfg.NewRelicLicense)
	srv.router = srv.routes()

	return srv
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

func (s *Server) Start() error {
	return http.ListenAndServe(":"+s.cfg.Port, s.router)
}

func (s *Server) Stop() error {
	log.Println("closing server")
	return s.db.Close()
}

func newDB(sqliteDBPath string) *sql.DB {
	db, err := db.New(sqliteDBPath)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func setupServices(srv *Server, mocking bool) {
	http.DefaultClient.Timeout = 10 * time.Second
	if mocking {
		log.Println("mocking enabled")
		tsToken, tsPlaylist := spotify.MockEndpoints("services/spotify/mock_playlist.json")
		srv.spotify = &spotify.Spotify{
			ClientID:     srv.cfg.SpotifyClientID,
			ClientSecret: srv.cfg.SpotifyClientSecret,

			TokenEndpoint:    tsToken.URL,
			PlaylistEndpoint: tsPlaylist.URL,

			Now: time.Now,
		}
		srv.spotify.InitTokenLifecycle()
	} else {
		srv.spotify = &spotify.Spotify{
			ClientID:     srv.cfg.SpotifyClientID,
			ClientSecret: srv.cfg.SpotifyClientSecret,

			TokenEndpoint:    "https://accounts.spotify.com/api/token",
			PlaylistEndpoint: "https://api.spotify.com/v1/playlists",

			Now: time.Now,
		}
		srv.spotify.InitTokenLifecycle()
	}
}

func newApm(environment, license string) *newrelic.Application {
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName("Playlist Vote "+environment),
		newrelic.ConfigLicense(license),
		newrelic.ConfigAppLogForwardingEnabled(true),
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

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "USE_EMBEDDED_PROD_CONFIG", "file path to server config file otherwise use the embedded prod config")
	flag.Parse()

	var cfg *config.Server
	if configPath == "USE_EMBEDDED_PROD_CONFIG" {
		cfg = config.ProdEmbeddedConfig()
	} else {
		cfg = config.CustomConfig(configPath)
	}

	srv := NewServer(cfg)

	slog.SetDefault(srv.log)

	err := srv.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = srv.Stop()
		if err != nil {
			log.Fatal(err)
		}
	}()
}
