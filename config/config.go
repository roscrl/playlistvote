package config

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const PathConfigDevBrowser = "config/.dev.browser.mock"

//go:embed .prod
var prodConfigFile embed.FS

type Server struct {
	Env                    Environment
	Port                   string
	SqliteDBPath           string
	Mocking                bool
	SpotifyClientID        string
	SpotifyClientSecret    string
	NewRelicLicense        string
	BasicDebugAuthUsername string
	BasicDebugAuthPassword string
}

func ProdEmbeddedConfig() *Server {
	cfg, err := LoadConfig(prodConfigFile, ".prod")
	if err != nil {
		log.Fatalf("failed to load embedded server config file: %v", err)
	}

	return cfg
}

func DevConfig() *Server {
	cfgPath := ".dev"

	cfg, err := LoadConfig(os.DirFS("./config"), cfgPath)
	if err != nil {
		log.Fatalf("error loading dev server config file: %v", err)
	}

	return cfg
}

func MockConfig() *Server {
	cfgPath := ".dev.mock"

	cfg, err := LoadConfig(os.DirFS("./config"), cfgPath)
	if err != nil {
		log.Fatalf("error loading mock server config file: %v", err)
	}

	return cfg
}

func CustomConfig(cfgPath string) *Server {
	absPath, err := filepath.Abs(cfgPath)
	if err != nil {
		log.Fatalf("error resolving server config path: %v", err)
	}

	cfg, err := LoadConfig(os.DirFS(filepath.Dir(absPath)), filepath.Base(absPath))
	if err != nil {
		log.Fatalf("error loading custom server config file: %v", err)
	}

	return cfg
}

func LoadConfig(fsys fs.FS, filePath string) (*Server, error) {
	file, err := fsys.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Server

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			// Skip comments and empty lines
			continue
		}

		parts := strings.SplitN(line, "=", 2) //nolint:gomnd
		if len(parts) != 2 {                  //nolint:gomnd
			return nil, fmt.Errorf("invalid line in .env file: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "ENVIRONMENT":
			env, err := parseEnvironment(value)
			if err != nil {
				return nil, fmt.Errorf("unknown environment in config dotfile file: %s = %s", key, value)
			}

			config.Env = env
		case "PORT":
			config.Port = value
		case "MOCK":
			config.Mocking = value == "true"
		case "SQLITE_DB_PATH":
			config.SqliteDBPath = value
		case "SPOTIFY_CLIENT_ID":
			config.SpotifyClientID = value
		case "SPOTIFY_CLIENT_SECRET":
			config.SpotifyClientSecret = value
		case "NEW_RELIC_LICENSE":
			config.NewRelicLicense = value
		case "BASIC_DEBUG_AUTH_USERNAME":
			config.BasicDebugAuthUsername = value
		case "BASIC_DEBUG_AUTH_PASSWORD":
			config.BasicDebugAuthPassword = value
		default:
			return nil, fmt.Errorf("unknown key in config dotfile file: %s", key)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &config, nil
}
