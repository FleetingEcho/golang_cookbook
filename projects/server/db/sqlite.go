package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

func Connect(databaseURL string) (*sql.DB, error) {
	path := strings.TrimPrefix(databaseURL, "sqlite://")
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}
	slog.Info("connected to sqlite", "path", path)
	return db, nil
}

func runMigrations(db *sql.DB) error {
	data, err := os.ReadFile("migrations/0001_init.sql")
	if err != nil {
		data, err = os.ReadFile("../migrations/0001_init.sql")
		if err != nil {
			return fmt.Errorf("read migration file: %w", err)
		}
	}
	_, err = db.Exec(string(data))
	return err
}
