package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"GitDb"
)

// RepoStore represents a per-repository KV store for HEAD/refs/objects/index operations
type RepoStore struct {
	repoID   string
	repoPath string
	db       *GitDb.DB
}

// NewRepoStore opens or creates a per-repo KV store for the given repository
// The store is rooted at data/repos/<repoId>/.gitclone/db
func NewRepoStore(repoBase, repoID string) (*RepoStore, error) {
	// Resolve repo path: join repoBase with repoID and validate
	// Prevent directory traversal attacks
	if strings.Contains(repoID, "..") || strings.Contains(repoID, "/") || strings.Contains(repoID, "\\") {
		return nil, fmt.Errorf("invalid repo ID: contains illegal characters")
	}

	repoPath := filepath.Join(repoBase, repoID)
	
	// Validate that .gitclone directory exists
	gitclonePath := filepath.Join(repoPath, ".gitclone")
	if _, err := os.Stat(gitclonePath); err != nil {
		return nil, fmt.Errorf("repository not found or invalid: %w", err)
	}

	// Determine database path: data/repos/<repoId>/.gitclone/db
	dbDir := filepath.Join(repoPath, ".gitclone", "db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	// Open GitDb for this specific repo
	db, err := GitDb.Open(dbDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &RepoStore{
		repoID:   repoID,
		repoPath: repoPath,
		db:       db,
	}, nil
}

// Close closes the database connection
func (rs *RepoStore) Close() error {
	if rs.db != nil {
		return rs.db.Close()
	}
	return nil
}

// DB returns the underlying GitDb.DB for direct access
// This should only be used for HEAD/refs/objects/index operations
func (rs *RepoStore) DB() *GitDb.DB {
	return rs.db
}

// RepoID returns the repository ID
func (rs *RepoStore) RepoID() string {
	return rs.repoID
}

// RepoPath returns the absolute path to the repository
func (rs *RepoStore) RepoPath() string {
	return rs.repoPath
}

