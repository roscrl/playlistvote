//nolint:gomnd
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"app/config"
	"app/db"
	"app/db/sqlc"
	"app/domain/spotify"
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

	client  *http.Client
	spotify *spotify.Spotify

	apm      *newrelic.Application
	router   http.Handler
	listener net.Listener
	port     string

	httpServer *http.Server
}

func NewServer(cfg *config.Server) *Server {
	srv := &Server{}

	srv.cfg = cfg
	srv.log = slog.New(slog.NewTextHandler(os.Stdout, nil))
	srv.db = db.New(cfg.SqliteDBPath)
	srv.qry = sqlc.New(srv.db)
	srv.views = views.New(srv.cfg.Env)

	srv.client = &http.Client{
		Timeout: 10 * time.Second,
	}

	setupServices(srv)

	srv.apm = newAPM(srv.cfg.Env.String(), srv.cfg.NewRelicLicense)
	srv.router = srv.routes()

	srv.httpServer = &http.Server{
		Handler:      srv.router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	return srv
}

func (s *Server) Start() {
	log.Printf("running in %v", s.cfg.Env)
	log.Printf("using db %v", s.cfg.SqliteDBPath)

	listener, err := net.Listen("tcp", ":"+s.cfg.Port)
	if err != nil {
		log.Fatal(err)
	}

	s.listener = listener
	s.port = fmt.Sprintf("%v", listener.Addr().(*net.TCPAddr).Port)

	go func() {
		err := s.httpServer.Serve(s.listener)
		if err != nil {
			var opErr *net.OpError
			if errors.As(err, &opErr) && opErr.Op == "accept" {
				log.Println("server shut down")
			} else {
				log.Fatal(err)
			}
		}
	}()

	log.Printf("ready to handle requests at :%v", s.port)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

func (s *Server) Stop() {
	log.Println("server shutting down...")

	if err := s.listener.Close(); err != nil {
		log.Fatalf("failed to shutdown: %v", err)
	}

	s.spotify.StopTokenLifecycle()

	err := s.db.Close()
	if err != nil {
		log.Fatalf("failed to close db connection: %v", err)
	}
}
