package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"GitDb"
	repostorage "gitclone/internal/infra/storage"
)

// ListBranchesFromStore lists branches using a RepoStore
func ListBranchesFromStore(store *repostorage.RepoStore) ([]string, error) {
	db := store.DB()
	var branches []string
	
	// Scan for all refs/heads/* keys
	err := db.Scan(func(record GitDb.Record) error {
		if strings.HasPrefix(record.Key, "refs/heads/") {
			branchName := strings.TrimPrefix(record.Key, "refs/heads/")
			branches = append(branches, branchName)
		}
		return nil
	})

	return branches, err
}

// ReadHEADBranchFromStore reads the current branch from HEAD using RepoStore
func ReadHEADBranchFromStore(store *repostorage.RepoStore) (string, error) {
	db := store.DB()
	data, err := db.Get("meta/HEAD")
	if err != nil {
		return "", err
	}
	
	// Parse "ref: refs/heads/<branch>\n"
	content := string(data)
	if !strings.HasPrefix(content, "ref: refs/heads/") {
		return "", fmt.Errorf("invalid HEAD format")
	}
	
	branch := strings.TrimPrefix(content, "ref: refs/heads/")
	branch = strings.TrimSuffix(branch, "\n")
	return branch, nil
}

// ReadHeadRefMaybeFromStore reads commit ID from refs/heads/<branch> using RepoStore
func ReadHeadRefMaybeFromStore(store *repostorage.RepoStore, branch string) (*int, error) {
	db := store.DB()
	key := "refs/heads/" + branch
	data, err := db.Get(key)
	if err != nil {
		return nil, nil
	}
	
	// Parse commit ID
	content := strings.TrimSpace(string(data))
	if content == "" {
		return nil, nil
	}
	
	commitID, err := strconv.Atoi(content)
	if err != nil {
		return nil, fmt.Errorf("invalid commit ID: %w", err)
	}
	
	return &commitID, nil
}

// GetStagedFilesFromStore returns staged files using RepoStore
func GetStagedFilesFromStore(store *repostorage.RepoStore) ([]string, error) {
	db := store.DB()
	indexData, err := db.Get("index/staged")
	if err != nil {
		// No staged files
		return []string{}, nil
	}

	var stagedFiles []string
	if err := json.Unmarshal(indexData, &stagedFiles); err != nil {
		return nil, fmt.Errorf("failed to unmarshal index: %w", err)
	}

	return stagedFiles, nil
}

// AddToIndexFromStore adds files to staging area using RepoStore
func AddToIndexFromStore(store *repostorage.RepoStore, paths []string) error {
	db := store.DB()
	
	// Read current index
	indexData, err := db.Get("index/staged")
	var stagedFiles []string
	if err != nil {
		// Index doesn't exist, create new
		stagedFiles = []string{}
	} else {
		if err := json.Unmarshal(indexData, &stagedFiles); err != nil {
			return fmt.Errorf("failed to unmarshal index: %w", err)
		}
	}

	// Add new paths (deduplicate)
	seen := make(map[string]bool)
	for _, path := range stagedFiles {
		seen[path] = true
	}

	repoPath := store.RepoPath()
	for _, path := range paths {
		// Normalize path
		normalizedPath := filepath.Clean(path)
		if !seen[normalizedPath] {
			// Check if file exists
			fullPath := filepath.Join(repoPath, normalizedPath)
			if _, err := os.Stat(fullPath); err == nil {
				stagedFiles = append(stagedFiles, normalizedPath)
				seen[normalizedPath] = true
			}
		}
	}

	// Write index back
	data, err := json.Marshal(stagedFiles)
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	return db.Put("index/staged", data)
}

// ClearIndexFromStore clears staging area using RepoStore
func ClearIndexFromStore(store *repostorage.RepoStore) error {
	db := store.DB()
	
	// Write empty array
	data, err := json.Marshal([]string{})
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	return db.Put("index/staged", data)
}

// GetPushedCommitsFromStore returns pushed commits for a branch using RepoStore
func GetPushedCommitsFromStore(store *repostorage.RepoStore, branch string) ([]int, error) {
	db := store.DB()
	key := fmt.Sprintf("remote/%s/commits", branch)
	data, err := db.Get(key)
	if err != nil {
		// No pushed commits yet
		return []int{}, nil
	}

	var commitIDs []int
	if err := json.Unmarshal(data, &commitIDs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pushed commits: %w", err)
	}

	return commitIDs, nil
}

// PushCommitFromStore marks a commit as pushed using RepoStore
func PushCommitFromStore(store *repostorage.RepoStore, branch string, commitID int) error {
	db := store.DB()
	
	// Read current remote commits for branch
	key := fmt.Sprintf("remote/%s/commits", branch)
	remoteData, err := db.Get(key)
	var pushedCommits []int
	if err != nil {
		// No remote commits yet, create new
		pushedCommits = []int{}
	} else {
		if err := json.Unmarshal(remoteData, &pushedCommits); err != nil {
			return fmt.Errorf("failed to unmarshal remote commits: %w", err)
		}
	}

	// Check if commit is already pushed
	for _, id := range pushedCommits {
		if id == commitID {
			// Already pushed
			return nil
		}
	}

	// Add commit to pushed list
	pushedCommits = append(pushedCommits, commitID)

	// Write back
	data, err := json.Marshal(pushedCommits)
	if err != nil {
		return fmt.Errorf("failed to marshal remote commits: %w", err)
	}

	return db.Put(key, data)
}

// ReadCommitObjectFromStore reads a commit object using RepoStore
func ReadCommitObjectFromStore(store *repostorage.RepoStore, commitID int) (Commit, error) {
	db := store.DB()
	
	// Read commit from DB
	key := fmt.Sprintf("objects/%d", commitID)
	data, err := db.Get(key)
	if err != nil {
		return Commit{}, err
	}

	// Decode JSON into Commit struct
	var c Commit
	if err := json.Unmarshal(data, &c); err != nil {
		return Commit{}, err
	}

	return c, nil
}

// WriteCommitObjectToBatch writes a commit object to a batch
func WriteCommitObjectToBatch(batch *repostorage.WriteBatch, commit Commit) error {
	data, err := json.MarshalIndent(commit, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal commit: %w", err)
	}
	
	key := fmt.Sprintf("objects/%d", commit.ID)
	batch.Put(key, data)
	return nil
}

// NextCommitIDFromStore gets and increments the next commit ID
func NextCommitIDFromStore(store *repostorage.RepoStore) (int, error) {
	db := store.DB()
	
	// Read current value
	b, err := db.Get("meta/NEXT_COMMIT_ID")
	if err != nil {
		return 0, err
	}

	curStr := strings.TrimSpace(string(b))
	cur, err := strconv.Atoi(curStr)
	if err != nil {
		return 0, fmt.Errorf("invalid NEXT_COMMIT_ID: %q", curStr)
	}

	// Write incremented value
	if err := db.Put("meta/NEXT_COMMIT_ID", []byte(fmt.Sprintf("%d\n", cur+1))); err != nil {
		return 0, err
	}

	return cur, nil
}

// WriteHeadRefToBatch writes a branch ref to a batch
func WriteHeadRefToBatch(batch *repostorage.WriteBatch, branch string, commitID int) error {
	// Ensure ref exists first (check if it exists, if not add empty)
	key := "refs/heads/" + branch
	batch.Put(key, []byte(fmt.Sprintf("%d\n", commitID)))
	return nil
}

// WriteHEADBranchToBatch writes HEAD to a batch
func WriteHEADBranchToBatch(batch *repostorage.WriteBatch, branch string) error {
	content := "ref: refs/heads/" + branch + "\n"
	batch.Put("meta/HEAD", []byte(content))
	return nil
}

// ClearIndexToBatch clears the index in a batch
func ClearIndexToBatch(batch *repostorage.WriteBatch) error {
	data, err := json.Marshal([]string{})
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}
	batch.Put("index/staged", data)
	return nil
}

// EnsureHeadRefExistsFromStore ensures HEAD ref exists using RepoStore
func EnsureHeadRefExistsFromStore(store *repostorage.RepoStore, branch string) error {
	return EnsureHeadRefExists(store.RepoPath(), InitOptions{Bare: false}, branch)
}
