package commits

import (
	"fmt"
	"time"

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

// CreateCommit creates a new commit with the given message atomically
func (s *Service) CreateCommit(repoID, message string) error {
	// Open per-repo store
	repoStore, err := storage.NewRepoStore(s.repoBase, repoID)
	if err != nil {
		return err
	}
	defer repoStore.Close()

	// Check if there are staged entries
	hasStaged, err := repostorage.HasStagedEntriesFromStore(repoStore)
	if err != nil {
		return fmt.Errorf("failed to check staged entries: %w", err)
	}

	if !hasStaged {
		return fmt.Errorf("nothing to commit. Stage changes first with 'gitclone add'")
	}

	// Get current branch
	currentBranch, err := repostorage.ReadHEADBranchFromStore(repoStore)
	if err != nil {
		return fmt.Errorf("failed to read current branch: %w", err)
	}

	// Get current branch tip for parent
	parentPtr, err := repostorage.ReadHeadRefMaybeFromStore(repoStore, currentBranch)
	if err != nil {
		return fmt.Errorf("failed to read branch tip: %w", err)
	}

	// Allocate commit ID (this needs to be done before batch)
	// For now, we'll read it directly - in a real system this should be atomic too
	commitID, err := repostorage.NextCommitIDFromStore(repoStore)
	if err != nil {
		return fmt.Errorf("failed to allocate commit ID: %w", err)
	}

	// Create commit object
	commit := repostorage.Commit{
		ID:        commitID,
		Message:   message,
		Branch:    currentBranch,
		Timestamp: time.Now().Unix(),
		Parent:    parentPtr,
	}

	// Create write batch for atomic operation
	batch := repoStore.NewWriteBatch()

	// Add all writes to batch:
	// 1. Commit object
	if err := repostorage.WriteCommitObjectToBatch(batch, commit); err != nil {
		return fmt.Errorf("failed to add commit to batch: %w", err)
	}

	// 2. Update branch ref
	if err := repostorage.WriteHeadRefToBatch(batch, currentBranch, commitID); err != nil {
		return fmt.Errorf("failed to add ref update to batch: %w", err)
	}

	// 3. Clear index
	if err := repostorage.ClearIndexToBatch(batch, repoStore); err != nil {
		return fmt.Errorf("failed to add index clear to batch: %w", err)
	}

	// Commit batch atomically
	if err := batch.Commit(); err != nil {
		return fmt.Errorf("failed to commit batch: %w", err)
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

