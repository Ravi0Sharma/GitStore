package storage

import (
	"encoding/json"
	"fmt"
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

// GetStagedFilesFromStore returns staged file paths using RepoStore
func GetStagedFilesFromStore(store *repostorage.RepoStore) ([]string, error) {
	entries, err := GetIndexEntriesFromStore(store)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(entries))
	for path, entry := range entries {
		if entry.BlobID != "" {
			paths = append(paths, path)
		}
	}

	return paths, nil
}

// GetIndexEntriesFromStore returns all staged entries using RepoStore
func GetIndexEntriesFromStore(store *repostorage.RepoStore) (map[string]IndexEntry, error) {
	db := store.DB()
	entries := make(map[string]IndexEntry)

	// Scan for all index/entries/* keys
	const indexEntriesPrefix = "index/entries/"
	err := db.Scan(func(record GitDb.Record) error {
		if strings.HasPrefix(record.Key, indexEntriesPrefix) {
			path := record.Key[len(indexEntriesPrefix):] // Remove "index/entries/" prefix

			var entry IndexEntry
			if err := json.Unmarshal(record.Value, &entry); err != nil {
				return nil // Skip invalid entries
			}

			// Only include entries with valid blobId
			if entry.BlobID != "" {
				entries[path] = entry
			}
		}
		return nil
	})

	return entries, err
}

// AddToIndexFromStore adds files to staging area using RepoStore
func AddToIndexFromStore(store *repostorage.RepoStore, path string) error {
	repoPath := store.RepoPath()
	options := InitOptions{Bare: false}
	return AddToIndex(repoPath, options, path)
}

// ClearIndexFromStore clears staging area using RepoStore
func ClearIndexFromStore(store *repostorage.RepoStore) error {
	repoPath := store.RepoPath()
	options := InitOptions{Bare: false}
	return ClearIndex(repoPath, options)
}

// HasStagedEntriesFromStore checks if there are staged entries using RepoStore
func HasStagedEntriesFromStore(store *repostorage.RepoStore) (bool, error) {
	entries, err := GetIndexEntriesFromStore(store)
	if err != nil {
		return false, err
	}

	// Check if any entries have valid blobId
	for _, entry := range entries {
		if entry.BlobID != "" {
			return true, nil
		}
	}

	return false, nil
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
// Since we can't truly delete in GitDb, we mark all entries as empty
func ClearIndexToBatch(batch *repostorage.WriteBatch, store *repostorage.RepoStore) error {
	// Get all current entries
	entries, err := GetIndexEntriesFromStore(store)
	if err != nil {
		return fmt.Errorf("failed to get index entries: %w", err)
	}

	// Mark all entries as cleared by writing empty entries
	for path := range entries {
		entryKey := fmt.Sprintf("index/entries/%s", path)
		emptyEntry := IndexEntry{BlobID: "", Mode: ""}
		entryData, err := json.Marshal(emptyEntry)
		if err != nil {
			return fmt.Errorf("failed to marshal empty entry: %w", err)
		}
		batch.Put(entryKey, entryData)
	}

	return nil
}

// EnsureHeadRefExistsFromStore ensures HEAD ref exists using RepoStore
func EnsureHeadRefExistsFromStore(store *repostorage.RepoStore, branch string) error {
	return EnsureHeadRefExists(store.RepoPath(), InitOptions{Bare: false}, branch)
}
