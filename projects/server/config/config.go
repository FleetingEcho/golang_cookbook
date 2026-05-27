package config

import (
	"net"
	"os"
	"path/filepath"
)

type Config struct {
	BindAddr    string
	DatabaseURL string
	UploadDir   string
	APIKey      string
}

func Load() *Config {
	cwd, _ := os.Getwd()
	defaultDB := filepath.Join(cwd, "data", "issue_tracker.db")
	defaultUpload := filepath.Join(cwd, "uploads")

	bind := os.Getenv("ISSUE_TRACKER_BIND_ADDR")
	if bind == "" {
		bind = "127.0.0.1:3001"
	}
	if _, err := net.ResolveTCPAddr("tcp", bind); err != nil {
		panic("ISSUE_TRACKER_BIND_ADDR must be a valid TCP address: " + err.Error())
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "sqlite://" + defaultDB
	}

	uploadDir := os.Getenv("ISSUE_TRACKER_UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = defaultUpload
	}

	apiKey := os.Getenv("ISSUE_TRACKER_API_KEY")
	if apiKey == "" {
		apiKey = "dev-secret"
	}

	return &Config{BindAddr: bind, DatabaseURL: dbURL, UploadDir: uploadDir, APIKey: apiKey}
}
