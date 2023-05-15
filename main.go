package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/exp/slog"

	"app/config"
)

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

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	srv.Start()

	<-stop

	srv.Stop()
}
