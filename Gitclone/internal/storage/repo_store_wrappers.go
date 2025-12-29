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

			// Since GitDb is append-only, Scan() iterates through all entries in order
			// For the same key, later entries overwrite earlier ones in the map
			// We only include entries with valid blobId (skip cleared entries with empty blobId)
			// But we need to check: if we've already seen this path with a valid blobId,
			// and now we see it with an empty blobId, we should remove it from the map
			if entry.BlobID != "" {
				entries[path] = entry
			} else {
				// Empty blobId means cleared - remove from map if it exists
				delete(entries, path)
			}
		}
		return nil
	})

	return entries, err
}

// AddToIndexFromStore adds files to staging area using RepoStore
// This uses the RepoStore's DB directly to ensure consistency with other operations
func AddToIndexFromStore(store *repostorage.RepoStore, path string) error {
	repoPath := store.RepoPath()
	db := store.DB()

	// Normalize path
	normalizedPath := filepath.Clean(path)
	if normalizedPath == "." {
		// Stage all files in repo (except .gitclone)
		return addAllFilesToIndexFromStore(repoPath, db)
	}

	// Stage single file or directory
	fullPath := filepath.Join(repoPath, normalizedPath)
	info, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("file not found: %s", normalizedPath)
	}

	if info.IsDir() {
		// Recursively add all files in directory
		return addDirectoryToIndexFromStore(repoPath, normalizedPath, db)
	}

	// Add single file
	return addFileToIndex(repoPath, normalizedPath, db)
}

// addAllFilesToIndexFromStore stages all files in repo using provided DB
func addAllFilesToIndexFromStore(root string, db *GitDb.DB) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .gitclone directory
		if info.IsDir() && filepath.Base(path) == ".gitclone" {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		// Normalize: remove leading ./ and convert to forward slashes
		relPath = filepath.Clean(relPath)
		relPath = filepath.ToSlash(relPath)
		// Remove leading ./ if present
		if strings.HasPrefix(relPath, "./") {
			relPath = relPath[2:]
		}

		return addFileToIndex(root, relPath, db)
	})
}

// addDirectoryToIndexFromStore recursively stages all files in a directory using provided DB
func addDirectoryToIndexFromStore(root, relPath string, db *GitDb.DB) error {
	fullPath := filepath.Join(root, relPath)
	return filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .gitclone directory
		if info.IsDir() && filepath.Base(path) == RepoDir {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		// Normalize path separators to forward slashes
		rel = filepath.ToSlash(rel)

		return addFileToIndex(root, rel, db)
	})
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

// ReadRemoteRefFromStore reads commit ID from refs/remotes/origin/<branch> using RepoStore
// Returns nil if branch has no remote ref (not pushed yet)
func ReadRemoteRefFromStore(store *repostorage.RepoStore, branch string) (*int, error) {
	db := store.DB()
	key := "refs/remotes/origin/" + branch
	data, err := db.Get(key)
	if err != nil {
		// Remote ref doesn't exist - branch hasn't been pushed yet
		return nil, nil
	}
	
	content := strings.TrimSpace(string(data))
	if content == "" {
		return nil, nil
	}
	
	commitID, err := strconv.Atoi(content)
	if err != nil {
		return nil, fmt.Errorf("invalid commit ID in remote ref: %w", err)
	}
	
	return &commitID, nil
}

// WriteRemoteRefFromStore writes commit ID into refs/remotes/origin/<branch> using RepoStore
func WriteRemoteRefFromStore(store *repostorage.RepoStore, branch string, commitID int) error {
	db := store.DB()
	key := "refs/remotes/origin/" + branch
	return db.Put(key, []byte(fmt.Sprintf("%d\n", commitID)))
}

// WriteRemoteRefToBatch writes remote ref to a batch
func WriteRemoteRefToBatch(batch *repostorage.WriteBatch, branch string, commitID int) error {
	key := "refs/remotes/origin/" + branch
	batch.Put(key, []byte(fmt.Sprintf("%d\n", commitID)))
	return nil
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
