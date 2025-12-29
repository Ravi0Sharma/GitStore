package branches

import (
	"os"
	"time"

	"gitclone/internal/app/repos"
	"gitclone/internal/commands"
	"gitclone/internal/metadata"
	"gitclone/internal/storage"
)

// Branch represents a git branch
type Branch struct {
	Name      string
	CreatedAt string
}

// Service handles branch operations
type Service struct {
	repoBase  string
	metaStore *metadata.Store
}

// NewService creates a new branches service
func NewService(repoBase string, metaStore *metadata.Store) *Service {
	return &Service{
		repoBase:  repoBase,
		metaStore: metaStore,
	}
}

// ListBranches returns all branches for a repository
func (s *Service) ListBranches(repoID string) ([]Branch, error) {
	repoPath, err := repos.ResolveRepoPath(s.repoBase, repoID)
	if err != nil {
		return nil, err
	}

	opts := storage.InitOptions{Bare: false}
	branchNames, err := storage.ListBranches(repoPath, opts)
	if err != nil {
		return nil, err
	}

	// Deduplicate branches by name
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

// Checkout switches to a branch, creating it if it doesn't exist
func (s *Service) Checkout(repoID, branchName string) error {
	repoPath, err := repos.ResolveRepoPath(s.repoBase, repoID)
	if err != nil {
		return err
	}

	oldDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(repoPath); err != nil {
		return err
	}

	commands.Checkout([]string{branchName})

	// Update metadata
	meta, err := s.metaStore.GetRepo(repoID)
	if err == nil {
		// Reload branch info
		branches, _ := s.ListBranches(repoID)
		meta.CurrentBranch = branchName
		meta.BranchCount = len(branches)
		meta.UpdatedAt = time.Now()
		if err := s.metaStore.UpdateRepo(*meta); err != nil {
			// Log but don't fail the operation
		}
	}

	return nil
}

