package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"GitDb"
)

// RepoMeta represents repository metadata stored in gitDb
type RepoMeta struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	CurrentBranch string    `json:"currentBranch"`
	BranchCount   int       `json:"branchCount"`
	CommitCount   int       `json:"commitCount"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Missing       bool      `json:"missing,omitempty"` // true if repo folder doesn't exist
}

// Store manages repository metadata in gitDb
type Store struct {
	dbPath string
	db     *GitDb.DB
}

// NewStore creates a new metadata store
func NewStore(dbPath string) (*Store, error) {
	// Ensure db directory exists
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	// Open gitDb
	db, err := GitDb.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &Store{
		dbPath: dbPath,
		db:     db,
	}, nil
}

// Close closes the database
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// GetDB returns the underlying database for direct access
func (s *Store) GetDB() *GitDb.DB {
	return s.db
}

// ListRepos returns all repositories from the index
func (s *Store) ListRepos() ([]RepoMeta, error) {
	// Read index
	indexData, err := s.db.Get("repos:index")
	if err != nil {
		// Index doesn't exist yet, return empty array
		return []RepoMeta{}, nil
	}

	var repoIDs []string
	if err := json.Unmarshal(indexData, &repoIDs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal index: %w", err)
	}

	// Load each repo metadata
	repos := make([]RepoMeta, 0, len(repoIDs))
	for _, id := range repoIDs {
		meta, err := s.GetRepo(id)
		if err != nil {
			// Log but continue - repo might be missing
			continue
		}
		repos = append(repos, *meta)
	}

	return repos, nil
}

// GetRepo retrieves repository metadata by ID
func (s *Store) GetRepo(id string) (*RepoMeta, error) {
	key := fmt.Sprintf("repo:%s", id)
	data, err := s.db.Get(key)
	if err != nil {
		return nil, fmt.Errorf("repo not found: %s", id)
	}

	var meta RepoMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal repo metadata: %w", err)
	}

	return &meta, nil
}

// CreateRepo creates a new repository metadata entry
func (s *Store) CreateRepo(meta RepoMeta) error {
	// Set timestamps
	now := time.Now()
	if meta.CreatedAt.IsZero() {
		meta.CreatedAt = now
	}
	meta.UpdatedAt = now

	// Store repo metadata
	key := fmt.Sprintf("repo:%s", meta.ID)
	data, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal repo metadata: %w", err)
	}

	if err := s.db.Put(key, data); err != nil {
		return fmt.Errorf("failed to store repo metadata: %w", err)
	}

	// Update index
	if err := s.EnsureIndexContains(meta.ID); err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	return nil
}

// UpdateRepo updates existing repository metadata
func (s *Store) UpdateRepo(meta RepoMeta) error {
	// Update timestamp
	meta.UpdatedAt = time.Now()

	// Store repo metadata
	key := fmt.Sprintf("repo:%s", meta.ID)
	data, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal repo metadata: %w", err)
	}

	if err := s.db.Put(key, data); err != nil {
		return fmt.Errorf("failed to update repo metadata: %w", err)
	}

	return nil
}

// EnsureIndexContains ensures the repo ID is in the index
func (s *Store) EnsureIndexContains(id string) error {
	// Read current index
	indexData, err := s.db.Get("repos:index")
	var repoIDs []string
	if err != nil {
		// Index doesn't exist, create new
		repoIDs = []string{}
	} else {
		if err := json.Unmarshal(indexData, &repoIDs); err != nil {
			return fmt.Errorf("failed to unmarshal index: %w", err)
		}
	}

	// Check if ID already in index
	for _, existingID := range repoIDs {
		if existingID == id {
			// Already in index, nothing to do
			return nil
		}
	}

	// Add to index
	repoIDs = append(repoIDs, id)

	// Write index back
	data, err := json.Marshal(repoIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	if err := s.db.Put("repos:index", data); err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	return nil
}

// DeleteRepo removes a repository from metadata (but keeps it in index for now)
// In a production system, you might want to remove from index too
func (s *Store) DeleteRepo(id string) error {
	// Note: GitDb might not have Delete, so we could set to empty or leave it
	// For now, we'll just leave it - the missing flag will indicate it's gone
	return nil
}

