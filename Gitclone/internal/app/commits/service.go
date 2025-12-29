package commits

import (
	"fmt"
	"os"
	"time"

	"gitclone/internal/commands"
	"gitclone/internal/infra/storage"
	"gitclone/internal/metadata"
	repostorage "gitclone/internal/storage"
)

// Commit represents a git commit
type Commit struct {
	Hash    string
	Message string
	Author  string
	Date    string
}

// Service handles commit operations
type Service struct {
	repoBase  string
	metaStore *metadata.Store
}

// NewService creates a new commits service
func NewService(repoBase string, metaStore *metadata.Store) *Service {
	return &Service{
		repoBase:  repoBase,
		metaStore: metaStore,
	}
}

// ListCommits returns commits for a repository branch
func (s *Service) ListCommits(repoID, branchName string, limit int) ([]Commit, error) {
	// Open per-repo store
	repoStore, err := storage.NewRepoStore(s.repoBase, repoID)
	if err != nil {
		return nil, err
	}
	defer repoStore.Close()

	// Use provided branch name, or default to current branch
	var targetBranch string
	if branchName != "" {
		targetBranch = branchName
	} else {
		var err error
		targetBranch, err = repostorage.ReadHEADBranchFromStore(repoStore)
		if err != nil {
			return []Commit{}, nil
		}
	}

	// Get pushed commits for this branch
	pushedCommits, err := repostorage.GetPushedCommitsFromStore(repoStore, targetBranch)
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
	tipPtr, err := repostorage.ReadHeadRefMaybeFromStore(repoStore, targetBranch)
	if err != nil || tipPtr == nil {
		return []Commit{}, nil
	}

	var commits []Commit
	id := *tipPtr
	count := 0

	for count < limit {
		c, err := repostorage.ReadCommitObjectFromStore(repoStore, id)
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

// CreateCommit creates a new commit with the given message
func (s *Service) CreateCommit(repoID, message string) error {
	// Open per-repo store
	repoStore, err := storage.NewRepoStore(s.repoBase, repoID)
	if err != nil {
		return err
	}
	defer repoStore.Close()

	repoPath := repoStore.RepoPath()

	stagedFiles, err := repostorage.GetStagedFilesFromStore(repoStore)
	if err != nil {
		return fmt.Errorf("failed to check staged files: %w", err)
	}

	if len(stagedFiles) == 0 {
		return fmt.Errorf("nothing to commit. Stage changes first with 'git add <path>'")
	}

	oldDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(repoPath); err != nil {
		return err
	}

	commands.Commit([]string{"-m", message})

	// Clear staging area after successful commit
	if err := repostorage.ClearIndexFromStore(repoStore); err != nil {
		// Log but don't fail the operation
	}

	return nil
}

// PushCommits pushes commits to remote
// Returns the number of commits pushed, or 0 if already up to date
func (s *Service) PushCommits(repoID, branch string) (int, error) {
	// Open per-repo store
	repoStore, err := storage.NewRepoStore(s.repoBase, repoID)
	if err != nil {
		return 0, err
	}
	defer repoStore.Close()

	// Determine branch
	if branch == "" {
		currentBranch, err := repostorage.ReadHEADBranchFromStore(repoStore)
		if err != nil {
			branch = "main"
		} else {
			branch = currentBranch
		}
	}

	// Get current branch tip
	tipPtr, err := repostorage.ReadHeadRefMaybeFromStore(repoStore, branch)
	if err != nil || tipPtr == nil {
		return 0, fmt.Errorf("no commits to push")
	}

	// Get already pushed commits
	pushedCommits, err := repostorage.GetPushedCommitsFromStore(repoStore, branch)
	if err != nil {
		return 0, fmt.Errorf("failed to get pushed commits: %w", err)
	}

	// Find commits that need to be pushed (walk from tip to last pushed commit)
	var commitsToPush []int
	currentID := *tipPtr

	for {
		isPushed := false
		for _, id := range pushedCommits {
			if id == currentID {
				isPushed = true
				break
			}
		}

		if isPushed {
			break
		}

		commitsToPush = append(commitsToPush, currentID)

		c, err := repostorage.ReadCommitObjectFromStore(repoStore, currentID)
		if err != nil {
			break
		}

		if c.Parent == nil {
			break
		}
		currentID = *c.Parent
	}

	if len(commitsToPush) == 0 {
		return 0, nil // Already up to date
	}

	// Push commits (mark as pushed)
	for _, commitID := range commitsToPush {
		if err := repostorage.PushCommitFromStore(repoStore, branch, commitID); err != nil {
			return 0, fmt.Errorf("failed to push commit %d: %w", commitID, err)
		}
	}

	// Update metadata commit count (using global store for repo registry)
	meta, err := s.metaStore.GetRepo(repoID)
	if err == nil {
		commits, _ := s.ListCommits(repoID, branch, 100)
		meta.CommitCount = len(commits)
		meta.UpdatedAt = time.Now()
		if err := s.metaStore.UpdateRepo(*meta); err != nil {
			// Log but don't fail the operation
		}
	}

	return len(commitsToPush), nil
}

