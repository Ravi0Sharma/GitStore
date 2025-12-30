package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"gitclone/internal/app/branches"
	"gitclone/internal/app/commits"
	"gitclone/internal/app/files"
	"gitclone/internal/infra/storage"
	"gitclone/internal/metadata"
	repostorage "gitclone/internal/storage"
)

// Server holds the server dependencies
type Server struct {
	repoBase  string
	metaStore *metadata.Store
	branchSvc *branches.Service
	commitSvc *commits.Service
	fileSvc   *files.Service
}

// NewServer creates a new server instance
func NewServer(repoBase string, metaStore *metadata.Store) *Server {
	return &Server{
		repoBase:  repoBase,
		metaStore: metaStore,
		branchSvc: branches.NewService(repoBase, metaStore),
		commitSvc: commits.NewService(repoBase, metaStore),
		fileSvc:   files.NewService(repoBase),
	}
}

// RepoBase returns the repository base path
func (s *Server) RepoBase() string {
	return s.repoBase
}

// MetaStore returns the metadata store
func (s *Server) MetaStore() *metadata.Store {
	return s.metaStore
}

// Helper functions for loading data

// LoadRepoSummary loads a repository summary
func (s *Server) LoadRepoSummary(repoPath, repoID string) (RepoListItem, error) {
	// Use services with RepoStore
	branches, _ := s.branchSvc.ListBranches(repoID)
	commits, _ := s.commitSvc.ListCommits(repoID, "", 100)

	currentBranch := ""
	if len(branches) > 0 {
		// Try to get current branch from metadata
		meta, err := s.metaStore.GetRepo(repoID)
		if err == nil {
			currentBranch = meta.CurrentBranch
		}
	}

	return RepoListItem{
		ID:            repoID,
		Name:          filepath.Base(repoID),
		CurrentBranch: currentBranch,
		BranchCount:   len(branches),
		CommitCount:   len(commits),
	}, nil
}

// LoadRepo loads a full repository with all details
func (s *Server) LoadRepo(repoPath, repoID string) (Repository, error) {
	// Use services with RepoStore
	branches, _ := s.branchSvc.ListBranches(repoID)
	commits, _ := s.commitSvc.ListCommits(repoID, "", 100)
	issues, _ := s.LoadIssues(repoID)

	// Get current branch from metadata
	currentBranch := ""
	meta, err := s.metaStore.GetRepo(repoID)
	if err == nil {
		currentBranch = meta.CurrentBranch
	}

	// Convert branches to HTTP types
	httpBranches := make([]Branch, len(branches))
	for i, b := range branches {
		httpBranches[i] = Branch{
			Name:      b.Name,
			CreatedAt: b.CreatedAt,
		}
	}

	// Convert commits to HTTP types
	httpCommits := make([]Commit, len(commits))
	for i, c := range commits {
		httpCommits[i] = Commit{
			Hash:    c.Hash,
			Message: c.Message,
			Author:  c.Author,
			Date:    c.Date,
		}
	}

	// Convert issues to []interface{} for Repository struct
	issuesInterface := make([]interface{}, len(issues))
	for i, issue := range issues {
		issuesInterface[i] = issue
	}

	return Repository{
		ID:            repoID,
		Name:          filepath.Base(repoID),
		CurrentBranch: currentBranch,
		Branches:      httpBranches,
		Commits:       httpCommits,
		Issues:        issuesInterface,
	}, nil
}

// LoadIssues loads all issues for a repository
func (s *Server) LoadIssues(repoID string) ([]Issue, error) {
	// Use metadata store's db directly
	db := s.metaStore.GetDB()
	if db == nil {
		return []Issue{}, nil
	}

	key := fmt.Sprintf("repo:%s:issues", repoID)
	data, err := db.Get(key)
	if err != nil {
		// No issues yet, return empty array
		return []Issue{}, nil
	}

	var issues []Issue
	if err := json.Unmarshal(data, &issues); err != nil {
		return nil, fmt.Errorf("failed to unmarshal issues: %w", err)
	}

	return issues, nil
}

// SaveIssue saves an issue to a repository
func (s *Server) SaveIssue(repoID string, issue Issue) error {
	// Load existing issues
	issues, err := s.LoadIssues(repoID)
	if err != nil {
		return err
	}

	// Add new issue
	issues = append(issues, issue)

	// Save back using metadata store's db
	db := s.metaStore.GetDB()
	if db == nil {
		return fmt.Errorf("database not available")
	}

	key := fmt.Sprintf("repo:%s:issues", repoID)
	data, err := json.Marshal(issues)
	if err != nil {
		return fmt.Errorf("failed to marshal issues: %w", err)
	}

	if err := db.Put(key, data); err != nil {
		return fmt.Errorf("failed to save issues: %w", err)
	}

	return nil
}

// IsAncestorFromStore checks if commitA is an ancestor of commitB using RepoStore
func (s *Server) IsAncestorFromStore(repoStore *storage.RepoStore, commitA, commitB int) bool {
	// If they're the same, it's trivially an ancestor
	if commitA == commitB {
		return true
	}

	// Walk backwards from commitB following parent pointers
	// If we reach commitA, then commitA is an ancestor of commitB
	visited := make(map[int]bool)
	queue := []int{commitB}
	maxDepth := 1000 // Safety limit to prevent infinite loops
	depth := 0

	for len(queue) > 0 && depth < maxDepth {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true
		depth++

		if current == commitA {
			return true
		}

		// Read commit and add parents to queue
		commit, err := repostorage.ReadCommitObjectFromStore(repoStore, current)
		if err != nil {
			// If we can't read the commit, stop searching
			break
		}

		if commit.Parent != nil {
			queue = append(queue, *commit.Parent)
		}
		// Note: We only follow Parent, not Parent2, for fast-forward detection
		// Parent2 would be from a previous merge, which breaks the linear history
	}

	return false
}

// RespondJSON is a helper to send JSON responses
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
