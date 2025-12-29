package branches

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"GitDb"
	"gitclone/internal/infra/storage"
	"gitclone/internal/metadata"
	repostorage "gitclone/internal/storage"
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
	// Open per-repo store
	repoStore, err := storage.NewRepoStore(s.repoBase, repoID)
	if err != nil {
		return nil, err
	}
	defer repoStore.Close()

	// Debug: log repo info
	repoPath := repoStore.RepoPath()
	dbPath := filepath.Join(repoPath, ".gitclone", "db")
	log.Printf("DEBUG ListBranches: repoID=%s, repoBase=%s, repoPath=%s, dbPath=%s", 
		repoID, s.repoBase, repoPath, dbPath)

	branchNames, err := repostorage.ListBranchesFromStore(repoStore)
	if err != nil {
		return nil, err
	}

	// Debug: log found branches
	log.Printf("DEBUG ListBranches: found %d branches: %v", len(branchNames), branchNames)
	
	// Debug: scan DB directly to see all refs/heads/* keys
	db := repoStore.DB()
	allRefKeys := make([]string, 0)
	err = db.Scan(func(record GitDb.Record) error {
		if strings.HasPrefix(record.Key, "refs/heads/") {
			allRefKeys = append(allRefKeys, record.Key)
		}
		return nil
	})
	if err == nil {
		log.Printf("DEBUG ListBranches: all refs/heads/* keys in DB: %v", allRefKeys)
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

// Checkout switches to a branch, creating it if it doesn't exist atomically
func (s *Service) Checkout(repoID, branchName string) error {
	// Open per-repo store
	repoStore, err := storage.NewRepoStore(s.repoBase, repoID)
	if err != nil {
		return err
	}
	defer repoStore.Close()

	// Debug: log repo info
	repoPath := repoStore.RepoPath()
	dbPath := filepath.Join(repoPath, ".gitclone", "db")
	log.Printf("DEBUG Checkout: repoID=%s, repoBase=%s, repoPath=%s, dbPath=%s, branchName=%s", 
		repoID, s.repoBase, repoPath, dbPath, branchName)

	// Read current branch
	currentBranch, err := repostorage.ReadHEADBranchFromStore(repoStore)
	if err != nil {
		return fmt.Errorf("failed to read current branch: %w", err)
	}

	// Check if same branch
	if branchName == currentBranch {
		return nil // Already on this branch
	}

	// Check if target branch exists (before batch)
	targetTip, err := repostorage.ReadHeadRefMaybeFromStore(repoStore, branchName)
	if err != nil {
		return fmt.Errorf("failed to read target branch tip: %w", err)
	}

	// Create write batch for atomic operation
	batch := repoStore.NewWriteBatch()

	// Ensure target branch ref exists in batch (create empty ref if new)
	// This is critical: even if branch is new and repo is empty, we must create the ref
	if targetTip == nil {
		// Branch doesn't exist yet - create it
		// First, try to copy current branch's tip if it exists
		currentTip, err := repostorage.ReadHeadRefMaybeFromStore(repoStore, currentBranch)
		if err != nil {
			return fmt.Errorf("failed to read current branch tip: %w", err)
		}
		if currentTip != nil {
			// Copy current tip to new branch
			if err := repostorage.WriteHeadRefToBatch(batch, branchName, *currentTip); err != nil {
				return fmt.Errorf("failed to add branch copy to batch: %w", err)
			}
			log.Printf("DEBUG Checkout: creating new branch %s with tip from %s (commit %d)", branchName, currentBranch, *currentTip)
		} else {
			// Empty repo - create empty ref (branch exists but has no commits)
			key := "refs/heads/" + branchName
			batch.Put(key, []byte(""))
			log.Printf("DEBUG Checkout: creating new branch %s with empty ref (no commits yet)", branchName)
		}
	} else {
		log.Printf("DEBUG Checkout: branch %s already exists with tip %d", branchName, *targetTip)
	}

	// Update HEAD to point to target branch
	if err := repostorage.WriteHEADBranchToBatch(batch, branchName); err != nil {
		return fmt.Errorf("failed to add HEAD update to batch: %w", err)
	}

	// Commit batch atomically
	if err := batch.Commit(); err != nil {
		return fmt.Errorf("failed to commit checkout batch: %w", err)
	}

	log.Printf("DEBUG Checkout: batch committed successfully, branch %s should now exist", branchName)

	// Update metadata (using global store for repo registry)
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

