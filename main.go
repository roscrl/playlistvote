package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"app/db"
	"app/db/sqlc"

	"github.com/newrelic/go-agent/v3/newrelic"

	"app/config"
	"app/services/spotify"
	"app/views"
)

type Server struct {
	env   config.Environment
	db    *sql.DB
	qry   *sqlc.Queries
	views *views.Views

	spotify *spotify.Spotify

	apm    *newrelic.Application
	router http.Handler
}

func NewServer(env config.Environment, sqliteDBPath string, mock bool) *Server {
	srv := &Server{}

	srv.env = env
	srv.db = database(sqliteDBPath)
	srv.qry = sqlc.New(srv.db)
	srv.views = views.New(srv.env)

	setupServices(srv, mock)

	srv.apm = apm()
	srv.router = srv.routes()

	return srv
}

func setupServices(srv *Server, mock bool) {
	http.DefaultClient.Timeout = 10 * time.Second
	if mock {
		log.Println("mocking enabled")
		tsToken, tsPlaylist := spotify.MockEndpoints()
		srv.spotify = &spotify.Spotify{
			ClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
			ClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),

			TokenEndpoint:    tsToken.URL,
			PlaylistEndpoint: tsPlaylist.URL,

			Now: time.Now,
		}
		srv.spotify.InitToken()
	} else {
		srv.spotify = &spotify.Spotify{
			ClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
			ClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),

			TokenEndpoint:    "https://accounts.spotify.com/api/token",
			PlaylistEndpoint: "https://api.spotify.com/v1/playlists",

			Now: time.Now,
		}
		srv.spotify.InitToken()
	}
}

func database(sqliteDBPath string) *sql.DB {
	db, err := db.New(sqliteDBPath)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func apm() *newrelic.Application {
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName("Playlist Vote"),
		newrelic.ConfigLicense(os.Getenv("NEW_RELIC_LICENSE")),
		newrelic.ConfigAppLogForwardingEnabled(true),
		newrelic.ConfigCodeLevelMetricsEnabled(true),
	)
	if err != nil {
		log.Fatal(err)
	}
	return app
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.router)
}

func (s *Server) Close() error {
	return s.db.Close()
}

func main() {
	var mock bool
	var sqliteDBPath string

	flag.BoolVar(&mock, "mock", false, "enable mocking of services")
	flag.StringVar(&sqliteDBPath, "sqliteDBPath", "./db/playlist-vote.db", "path to sqlite db")

	flag.Parse()

	err := config.LoadEnv(".env")
	if err != nil {
		log.Fatal("error loading .env file")
	}

	var srv *Server
	if os.Getenv("ENV") == "prod" {
		srv = NewServer(config.PROD, sqliteDBPath, mock)
	} else {
		srv = NewServer(config.DEV, sqliteDBPath, mock)
	}

	defer func() {
		err := srv.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	err = srv.Start(":3000")
	if err != nil {
		log.Fatal(err)
	}
}
