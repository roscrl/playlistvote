package main

import (
	"flag"
	"log"

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

	log.Printf("running in %v", cfg.Env)
	log.Printf("using db %v", cfg.SqliteDBPath)

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
