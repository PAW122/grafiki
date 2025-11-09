package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func openDatabase(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("file:%s?_busy_timeout=5000&_pragma=foreign_keys(ON)", filepath.ToSlash(path))
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	if _, err := db.Exec(`PRAGMA journal_mode = WAL;`); err != nil {
		db.Close()
		return nil, err
	}

	if err := migrateDatabase(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func migrateDatabase(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS folders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		slug TEXT NOT NULL UNIQUE,
		path TEXT NOT NULL UNIQUE,
		visibility TEXT NOT NULL DEFAULT 'private',
		shared_token TEXT UNIQUE,
		shared_views INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_folders_slug ON folders(slug);
	CREATE INDEX IF NOT EXISTS idx_folders_visibility ON folders(visibility);

	INSERT INTO folders (name, slug, path, visibility)
	SELECT 'Domyslny', 'domyslny', '', 'public'
	WHERE NOT EXISTS (SELECT 1 FROM folders WHERE path = '');
	`

	_, err := db.Exec(schema)
	return err
}
