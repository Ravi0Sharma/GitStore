package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gitclone/internal/metadata"
	"gitclone/internal/storage"
)

// Server holds the server dependencies
type Server struct {
	repoBase  string
	metaStore *metadata.Store
}

// NewServer creates a new server instance
func NewServer(repoBase string, metaStore *metadata.Store) *Server {
	return &Server{
		repoBase:  repoBase,
		metaStore: metaStore,
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
	opts := storage.InitOptions{Bare: false}

	currentBranch, _ := storage.ReadHEADBranch(repoPath, opts)
	branches, _ := s.LoadBranches(repoPath)
	commits, _ := s.LoadCommits(repoPath, currentBranch, 100)

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
	opts := storage.InitOptions{Bare: false}

	currentBranch, _ := storage.ReadHEADBranch(repoPath, opts)
	branches, _ := s.LoadBranches(repoPath)
	commits, _ := s.LoadCommits(repoPath, currentBranch, 100)
	issues, _ := s.LoadIssues(repoID)

	// Convert issues to []interface{} for Repository struct
	issuesInterface := make([]interface{}, len(issues))
	for i, issue := range issues {
		issuesInterface[i] = issue
	}

	return Repository{
		ID:            repoID,
		Name:          filepath.Base(repoID),
		CurrentBranch: currentBranch,
		Branches:      branches,
		Commits:       commits,
		Issues:        issuesInterface,
	}, nil
}

// LoadBranches loads all branches for a repository
func (s *Server) LoadBranches(repoPath string) ([]Branch, error) {
	opts := storage.InitOptions{Bare: false}

	branchNames, err := storage.ListBranches(repoPath, opts)
	if err != nil {
		return nil, err
	}

	// Deduplicate branches by name (use map to track seen branches)
	seen := make(map[string]bool)
	uniqueNames := make([]string, 0, len(branchNames))
	for _, name := range branchNames {
		if !seen[name] {
			seen[name] = true
			uniqueNames = append(uniqueNames, name)
		}
	}

	branches := make([]Branch, 0, len(uniqueNames))
	for _, name := range uniqueNames {
		branches = append(branches, Branch{
			Name:      name,
			CreatedAt: time.Now().Format(time.RFC3339), // TODO: get actual creation time
		})
	}

	return branches, nil
}

// LoadCommits loads commits for a repository branch
func (s *Server) LoadCommits(repoPath string, branchName string, limit int) ([]Commit, error) {
	opts := storage.InitOptions{Bare: false}

	// Only GitClone repos are supported
	if !s.isGitCloneRepo(repoPath) {
		return []Commit{}, nil
	}

	// Use provided branch name, or default to current branch
	var targetBranch string
	if branchName != "" {
		targetBranch = branchName
	} else {
		var err error
		targetBranch, err = storage.ReadHEADBranch(repoPath, opts)
		if err != nil {
			return []Commit{}, nil
		}
	}

	// Get pushed commits for this branch
	pushedCommits, err := storage.GetPushedCommits(repoPath, opts, targetBranch)
	if err != nil {
		return []Commit{}, err
	}

	if len(pushedCommits) == 0 {
		return []Commit{}, nil
	}

	// Create a map for quick lookup
	pushedMap := make(map[int]bool)
	for _, id := range pushedCommits {
		pushedMap[id] = true
	}

	// Read commits using GitClone storage, but only include pushed ones
	tipPtr, err := storage.ReadHeadRefMaybe(repoPath, opts, targetBranch)
	if err != nil || tipPtr == nil {
		return []Commit{}, nil
	}

	var commits []Commit
	id := *tipPtr
	count := 0

	for count < limit {
		c, err := storage.ReadCommitObject(repoPath, opts, id)
		if err != nil {
			break
		}

		// Only include pushed commits
		if pushedMap[c.ID] {
			commits = append(commits, Commit{
				Hash:    fmt.Sprintf("%d", c.ID),
				Message: c.Message,
				Author:  "system", // TODO: get from commit
				Date:    time.Unix(c.Timestamp, 0).Format(time.RFC3339),
			})
			count++
		}

		if c.Parent == nil {
			break
		}
		id = *c.Parent
	}

	return commits, nil
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

// isGitCloneRepo checks if a repository is a GitClone repo (.gitclone/)
func (s *Server) isGitCloneRepo(repoPath string) bool {
	gitclonePath := filepath.Join(repoPath, storage.RepoDir)
	hasGitClone, err := os.Stat(gitclonePath)
	return err == nil && hasGitClone.IsDir()
}

// IsAncestor checks if commitA is an ancestor of commitB (i.e., commitA is reachable from commitB)
func (s *Server) IsAncestor(repoPath string, opts storage.InitOptions, commitA, commitB int) bool {
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
		commit, err := storage.ReadCommitObject(repoPath, opts, current)
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

