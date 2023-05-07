package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const (
	PlaylistsTable = "playlists"
)

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
