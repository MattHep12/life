package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const databaseEnvironmentVariable = "LIFE_DB_PATH"

// DefaultPath returns the database used by Life and its import commands.
// Existing repository-local databases keep working, while new installations
// use the operating system's per-user configuration directory.
func DefaultPath() (string, error) {
	if path := os.Getenv(databaseEnvironmentVariable); path != "" {
		return filepath.Abs(path)
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("find working directory: %w", err)
	}
	legacyPath := filepath.Join(workingDirectory, "life.db")
	if _, err := os.Stat(legacyPath); err == nil {
		return legacyPath, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("inspect existing database: %w", err)
	}

	configDirectory, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("find user data directory: %w", err)
	}
	dataDirectory := filepath.Join(configDirectory, "Life")
	if err := os.MkdirAll(dataDirectory, 0o700); err != nil {
		return "", fmt.Errorf("create user data directory: %w", err)
	}
	return filepath.Join(dataDirectory, "life.db"), nil
}
