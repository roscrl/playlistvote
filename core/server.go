//nolint:gomnd
package core

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"app/config"
	"app/core/broadcast"
	"app/core/db"
	"app/core/db/sqlc"
	"app/core/rlog"
	"app/core/session"
	"app/core/spotify"
	"app/core/views"
	"github.com/newrelic/go-agent/v3/newrelic"
	"golang.org/x/exp/slog"
)

type Server struct {
	Cfg   *config.Server
	Log   *slog.Logger
	DB    *sql.DB
	Qry   *sqlc.Queries
	Views *views.Views

	Client  *http.Client
	Spotify *spotify.Client

	SessionStore             *session.Store
	UpvoteBroadcasterService *broadcast.UpvoteBroadcasterService

	APM      *newrelic.Application
	Router   http.Handler
	Listener net.Listener
	Port     string

	HTTPServer *http.Server
}

func NewServer(cfg *config.Server) *Server {
	srv := &Server{}

	srv.Cfg = cfg
	srv.Log = rlog.NewDefaultLogger()
	srv.DB = db.New(cfg.SqliteDBPath)
	srv.Qry = sqlc.New(srv.DB)
	srv.Views = views.New(srv.Cfg.Env)

	srv.Client = &http.Client{
		Timeout: 10 * time.Second,
	}

	setupServices(srv)

	srv.APM = newAPM(srv.Cfg.Env.String(), srv.Cfg.NewRelicLicense)
	srv.Router = srv.routes()

	srv.HTTPServer = &http.Server{
		Handler:     srv.Router,
		ReadTimeout: 5 * time.Second,
		// WriteTimeout: 10 * time.Second, TODO this is causing issues with the websocket
		IdleTimeout: 15 * time.Second,
	}

	return srv
}

func (s *Server) Start() {
	log.Printf("running in %v", s.Cfg.Env)
	log.Printf("using db %v", s.Cfg.SqliteDBPath)

	listener, err := net.Listen("tcp", ":"+s.Cfg.Port)
	if err != nil {
		log.Fatal(err)
	}

	s.Listener = listener
	s.Port = fmt.Sprintf("%v", listener.Addr().(*net.TCPAddr).Port)

	go func() {
		err := s.HTTPServer.Serve(s.Listener)
		if err != nil {
			var opErr *net.OpError
			if errors.As(err, &opErr) && opErr.Op == "accept" {
				log.Println("server shut down")
			} else {
				log.Fatal(err)
			}
		}
	}()

	log.Printf("ready to handle requests at :%v", s.Port)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}

func (s *Server) Stop() {
	log.Println("server shutting down...")

	if err := s.Listener.Close(); err != nil {
		log.Fatalf("failed to shutdown: %v", err)
	}

	s.Spotify.StopTokenLifecycle()

	err := s.DB.Close()
	if err != nil {
		log.Fatalf("failed to close db connection: %v", err)
	}
}
