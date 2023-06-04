package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	_ "github.com/mattn/go-sqlite3" // sqlite3 driver
)

const (
	PlaylistsTable = "playlists"
)

var MigrationsPath = ""

func init() {
	_, filename, _, _ := runtime.Caller(0)
	MigrationsPath = filepath.Dir(filename) + "/migrate"
}

func New(dbFilePath string) *sql.DB {
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		log.Fatal(err)
	}

	err = setPragmas(db)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func NewInMemory() *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}

	err = setPragmas(db)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func RunMigrations(db *sql.DB, migrationsPath string) {
	migrationsDir, err := os.ReadDir(migrationsPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range migrationsDir {
		if strings.HasSuffix(file.Name(), ".sql") {
			migration, err := os.ReadFile(migrationsPath + "/" + file.Name())
			if err != nil {
				log.Fatal(err)
			}

			_, err = db.Exec(string(migration))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func setPragmas(db *sql.DB) error {
	_, err := db.Exec("PRAGMA synchronous = OFF")
	if err != nil {
		return err
	}

	_, err = db.Exec("PRAGMA cache_size = 50000")
	if err != nil {
		return err
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return err
	}

	_, err = db.Exec("PRAGMA busy_timeout = 5000")
	if err != nil {
		return err
	}

	_, err = db.Exec("PRAGMA temp_store = MEMORY")
	if err != nil {
		return err
	}

	_, err = db.Exec("PRAGMA mmap_size = 300000000")
	if err != nil {
		return err
	}

	return nil
}
