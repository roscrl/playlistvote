package main

import (
	"database/sql"
	"net/http"
	"os"

	"app/config"
	"app/db"
	"app/db/sqlc"
	"app/services/spotify"
	"app/views"

	"github.com/newrelic/go-agent/v3/newrelic"
	"golang.org/x/exp/slog"
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
	srv.log = slog.New(slog.NewTextHandler(os.Stdout, nil))
	srv.db = db.New(cfg.SqliteDBPath)
	srv.qry = sqlc.New(srv.db)
	srv.views = views.New(srv.cfg.Env)

	setupServices(srv)

	srv.apm = newAPM(srv.cfg.Env.String(), srv.cfg.NewRelicLicense)
	srv.router = srv.routes()

	return srv
}

func (s *Server) Start() error {
	return http.ListenAndServe(":"+s.cfg.Port, s.router)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

func (s *Server) Stop() error {
	return s.db.Close()
}
