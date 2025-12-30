package commits

import (
	"encoding/json"
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

	// Read from remote ref (refs/remotes/origin/<branch>) - this is the pushed state
	// If branch hasn't been pushed yet, return empty list
	log.Printf("DEBUG ListCommits: repoID=%s, branchName=%s, targetBranch=%s", repoID, branchName, targetBranch)
	
	// Debug: check both local and remote refs
	headTip, _ := repostorage.ReadHeadRefMaybeFromStore(repoStore, targetBranch)
	if headTip != nil {
		log.Printf("DEBUG ListCommits: refs/heads/%s = %d", targetBranch, *headTip)
	} else {
		log.Printf("DEBUG ListCommits: refs/heads/%s = (empty)", targetBranch)
	}
	
	tipPtr, err := repostorage.ReadRemoteRefFromStore(repoStore, targetBranch)
	if err != nil {
		log.Printf("DEBUG ListCommits: error reading remote ref: %v", err)
		return []Commit{}, err
	}
	if tipPtr == nil {
		// Branch hasn't been pushed yet - no commits to show
		log.Printf("DEBUG ListCommits: refs/remotes/origin/%s = (empty) - returning empty commits list", targetBranch)
		return []Commit{}, nil
	}
	
	log.Printf("DEBUG ListCommits: refs/remotes/origin/%s = %d - will walk from this commit", targetBranch, *tipPtr)

	// Walk commit history from remote ref tip
	var commits []Commit
	id := *tipPtr
	count := 0

	for count < limit {
		c, err := repostorage.ReadCommitObjectFromStore(repoStore, id)
		if err != nil {
			break
		}

		// All commits from remote ref are pushed commits
		commits = append(commits, Commit{
			Hash:    fmt.Sprintf("%d", c.ID),
			Message: c.Message,
			Author:  "system", // TODO: get from commit
			Date:    time.Unix(c.Timestamp, 0).Format(time.RFC3339),
		})
		count++

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

	// Debug: log repo info - verify DB path matches StageFiles
	repoPath := repoStore.RepoPath()
	dbPath := filepath.Join(repoPath, ".gitclone", "db")
	log.Printf("DEBUG CreateCommit: repoID=%s, repoBase=%s, repoPath=%s, dbPath=%s", 
		repoID, s.repoBase, repoPath, dbPath)
	
	// Verify RepoStore DB path matches expected
	actualDBPath := filepath.Join(repoStore.RepoPath(), ".gitclone", "db")
	log.Printf("DEBUG CreateCommit: RepoStore.RepoPath()=%s, actualDBPath=%s", repoStore.RepoPath(), actualDBPath)

	// Check if there are staged entries
	entries, err := repostorage.GetIndexEntriesFromStore(repoStore)
	if err != nil {
		log.Printf("DEBUG CreateCommit: error getting index entries: %v", err)
		return fmt.Errorf("failed to check staged entries: %w", err)
	}
	
	stagedCount := len(entries)
	log.Printf("DEBUG CreateCommit: staged entries count: %d", stagedCount)
	
	// Debug: scan DB directly to see all index/entries/* keys
	db := repoStore.DB()
	allIndexKeys := make([]string, 0)
	const indexEntriesPrefix = "index/entries/"
	err = db.Scan(func(record GitDb.Record) error {
		if strings.HasPrefix(record.Key, indexEntriesPrefix) {
			allIndexKeys = append(allIndexKeys, record.Key)
			// Try to unmarshal to check blobId
			var entry repostorage.IndexEntry
			if err := json.Unmarshal(record.Value, &entry); err == nil {
				log.Printf("DEBUG CreateCommit: found key=%s, blobId=%q, mode=%q", 
					record.Key, entry.BlobID, entry.Mode)
			} else {
				log.Printf("DEBUG CreateCommit: found key=%s, unmarshal error: %v", record.Key, err)
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("DEBUG CreateCommit: error scanning DB: %v", err)
	}
	log.Printf("DEBUG CreateCommit: total index/entries/* keys in DB: %d", len(allIndexKeys))
	
	// Log a few entry keys for debugging
	if stagedCount > 0 {
		listed := 0
		for p := range entries {
			log.Printf("DEBUG CreateCommit: staged path: %s", p)
			listed++
			if listed >= 5 {
				break
			}
		}
	} else {
		log.Printf("DEBUG CreateCommit: WARNING - no staged entries found!")
		// Try to scan DB directly to see what keys exist
		db := repoStore.DB()
		var allKeys []string
		var indexKeys []string
		_ = db.Scan(func(record GitDb.Record) error {
			allKeys = append(allKeys, record.Key)
			if strings.HasPrefix(record.Key, "index/entries/") {
				indexKeys = append(indexKeys, record.Key)
			}
			return nil
		})
		log.Printf("DEBUG CreateCommit: total keys in DB: %d, index/entries/* keys: %d", len(allKeys), len(indexKeys))
		if len(indexKeys) > 0 {
			for i, k := range indexKeys {
				if i < 5 {
					log.Printf("DEBUG CreateCommit: found index key: %s", k)
				}
			}
		}
	}
	
	hasStaged := stagedCount > 0
	if !hasStaged {
		return fmt.Errorf("Nothing to commit. Stage changes first with 'git add <path>' or 'gitclone add <path>'")
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

	log.Printf("DEBUG PushCommits: repoID=%s, branch=%s", repoID, branch)

	// Get current branch tip (refs/heads/<branch>)
	headTipPtr, err := repostorage.ReadHeadRefMaybeFromStore(repoStore, branch)
	if err != nil || headTipPtr == nil {
		return 0, fmt.Errorf("no commits to push")
	}
	headTip := *headTipPtr
	log.Printf("DEBUG PushCommits: refs/heads/%s = %d", branch, headTip)

	// Get current remote ref (refs/remotes/origin/<branch>)
	remoteTipPtr, err := repostorage.ReadRemoteRefFromStore(repoStore, branch)
	if err != nil {
		return 0, fmt.Errorf("failed to get remote ref: %w", err)
	}
	if remoteTipPtr != nil {
		log.Printf("DEBUG PushCommits: refs/remotes/origin/%s = %d", branch, *remoteTipPtr)
	} else {
		log.Printf("DEBUG PushCommits: refs/remotes/origin/%s = (empty)", branch)
	}

	// If remote ref doesn't exist or is behind, push all commits from head to remote
	// Push sets: refs/remotes/origin/<branch> = refs/heads/<branch>

	// Check if already up to date
	if remoteTipPtr != nil && *remoteTipPtr == headTip {
		log.Printf("DEBUG PushCommits: already up to date, no push needed")
		return 0, nil // Already up to date
	}

	// Count commits to push (walk from head tip to remote tip or root)
	var commitsToPush []int
	currentID := headTip

	for {
		// Stop if we've reached the remote tip (if it exists)
		if remoteTipPtr != nil && currentID == *remoteTipPtr {
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

	// Push: set remote ref to head ref (atomic write)
	batch := repoStore.NewWriteBatch()
	if err := repostorage.WriteRemoteRefToBatch(batch, branch, headTip); err != nil {
		return 0, fmt.Errorf("failed to add remote ref to batch: %w", err)
	}
	if err := batch.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit push: %w", err)
	}
	log.Printf("DEBUG PushCommits: pushed %d commits, updated refs/remotes/origin/%s to %d", len(commitsToPush), branch, headTip)

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

