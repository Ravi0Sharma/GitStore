package storage

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"GitDb"
)

// IndexEntry represents a single entry in the staging area
// Stored as: index/entries/<path> -> {blobId, mode}
type IndexEntry struct {
	BlobID string `json:"blobId"` // SHA1 hash of file content (or simple ID for now)
	Mode   string `json:"mode"`   // File mode: "100644" for regular files, "100755" for executables, "040000" for directories
}

// AddToIndex stages files to the index
// Stores entries as index/entries/<path> -> {blobId, mode}
func AddToIndex(root string, options InitOptions, path string) error {
	db, err := openDB(root, options)
	if err != nil {
		return err
	}
	// Note: We defer Close() to ensure writes are persisted
	// GitDb.Put() writes directly to file AND updates in-memory log
	// Close() writes the in-memory log to file (which should be identical)
	defer db.Close()

	// Normalize path
	normalizedPath := filepath.Clean(path)
	if normalizedPath == "." {
		// Stage all files in repo (except .gitclone)
		return addAllFilesToIndex(root, options, db)
	}

	// Stage single file or directory
	fullPath := filepath.Join(root, normalizedPath)
	info, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("file not found: %s", normalizedPath)
	}

	if info.IsDir() {
		// Recursively add all files in directory
		return addDirectoryToIndex(root, normalizedPath, options, db)
	}

	// Add single file
	return addFileToIndex(root, normalizedPath, db)
}

// addFileToIndex stages a single file
func addFileToIndex(root, relPath string, db *GitDb.DB) error {
	fullPath := filepath.Join(root, relPath)

	// Read file content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Compute blob ID (simple SHA1 hash for now)
	hash := sha1.Sum(content)
	blobID := fmt.Sprintf("%x", hash)

	// Determine file mode
	mode := "100644" // Regular file
	info, err := os.Stat(fullPath)
	if err == nil && info.Mode()&0111 != 0 {
		mode = "100755" // Executable
	}

	// Create index entry
	entry := IndexEntry{
		BlobID: blobID,
		Mode:   mode,
	}

	// Store blob object
	blobKey := fmt.Sprintf("objects/blob/%s", blobID)
	if err := db.Put(blobKey, content); err != nil {
		return fmt.Errorf("failed to store blob: %w", err)
	}

	// Normalize path separators to forward slashes for consistency
	normalizedRelPath := filepath.ToSlash(relPath)

	// Store index entry: index/entries/<path> -> {blobId, mode}
	entryKey := fmt.Sprintf("index/entries/%s", normalizedRelPath)
	entryData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	return db.Put(entryKey, entryData)
}

// addDirectoryToIndex recursively stages all files in a directory
func addDirectoryToIndex(root, relPath string, options InitOptions, db *GitDb.DB) error {
	fullPath := filepath.Join(root, relPath)

	return filepath.Walk(fullPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip .gitclone directory
		if info.IsDir() && info.Name() == RepoDir {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil // Continue walking
		}

		// Get relative path from repo root
		fileRelPath, err := filepath.Rel(root, filePath)
		if err != nil {
			return nil // Skip if we can't get relative path
		}

		// Normalize path separators
		fileRelPath = filepath.ToSlash(fileRelPath)

		// Add file to index
		return addFileToIndex(root, fileRelPath, db)
	})
}

// addAllFilesToIndex stages all files in the repository
func addAllFilesToIndex(root string, options InitOptions, db *GitDb.DB) error {
	return filepath.Walk(root, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip .gitclone directory
		if info.IsDir() && info.Name() == RepoDir {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(root, filePath)
		if err != nil {
			return nil
		}

		// Normalize path separators
		relPath = filepath.ToSlash(relPath)

		return addFileToIndex(root, relPath, db)
	})
}

// GetIndexEntries returns all staged entries from the index
func GetIndexEntries(root string, options InitOptions) (map[string]IndexEntry, error) {
	db, err := openDB(root, options)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	entries := make(map[string]IndexEntry)

	// Scan for all index/entries/* keys
	// GitDb.Scan() iterates through the in-memory log, which is rebuilt from file on Open()
	// Since GitDb.Put() writes directly to the log file, and Open() reads the file,
	// we should see all writes from AddToIndex() even if it closed the DB
	const indexEntriesPrefix = "index/entries/"
	err = db.Scan(func(record GitDb.Record) error {
		// Check if key starts with "index/entries/" using strings.HasPrefix
		if strings.HasPrefix(record.Key, indexEntriesPrefix) {
			path := record.Key[len(indexEntriesPrefix):] // Remove "index/entries/" prefix

			var entry IndexEntry
			if err := json.Unmarshal(record.Value, &entry); err != nil {
				// Skip invalid entries but don't fail
				return nil
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

// ClearIndex clears all entries from the staging area
func ClearIndex(root string, options InitOptions) error {
	db, err := openDB(root, options)
	if err != nil {
		return err
	}
	defer db.Close()

	// Get all index entry keys (including those with empty blobId to find all keys)
	// We need to scan for ALL keys with the prefix, not just valid entries
	const indexEntriesPrefix = "index/entries/"
	
	// Scan for ALL keys with the prefix (including already-cleared entries)
	// We need to find all keys, not just valid entries, to clear them all
	allPaths := make(map[string]bool)
	err = db.Scan(func(record GitDb.Record) error {
		if strings.HasPrefix(record.Key, indexEntriesPrefix) {
			path := record.Key[len(indexEntriesPrefix):]
			allPaths[path] = true
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Clear by writing empty entries for all paths found
	// Since GitDb is append-only, we can't truly delete, but we can mark entries as cleared
	// by writing entries with empty blobId, which GetIndexEntries() will filter out
	// GitDb's index will point to the latest (empty) entry for each key
	for path := range allPaths {
		entryKey := fmt.Sprintf("index/entries/%s", path)
		// Write empty entry to effectively "delete" it
		emptyEntry := IndexEntry{BlobID: "", Mode: ""}
		emptyEntryData, err := json.Marshal(emptyEntry)
		if err != nil {
			return fmt.Errorf("failed to marshal empty entry: %w", err)
		}
		if err := db.Put(entryKey, emptyEntryData); err != nil {
			return fmt.Errorf("failed to clear entry %s: %w", path, err)
		}
	}

	return nil
}

// HasStagedEntries checks if there are any staged entries
func HasStagedEntries(root string, options InitOptions) (bool, error) {
	entries, err := GetIndexEntries(root, options)
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

