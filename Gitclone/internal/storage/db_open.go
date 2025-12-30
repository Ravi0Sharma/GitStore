package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"GitDb"
)

// dbPath returns the path to the database directory
func dbPath(root string, options InitOptions) string {
	repoRootPath := repoRoot(root, options)
	return filepath.Join(repoRootPath, "db")
}

// openDB opens the database
func openDB(root string, options InitOptions) (*GitDb.DB, error) {
	dbDir := dbPath(root, options)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}
	// Note: GitDb.Open currently ignores the path parameter
	db, err := GitDb.Open(dbDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return db, nil
}

