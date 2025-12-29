package storage

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
)

// TreeEntry represents a single entry in a tree object
// Tree format: objects/tree/<treeId> -> JSON array of {path, blobId, mode, type}
type TreeEntry struct {
	Path string `json:"path"` // Relative path within tree
	BlobID string `json:"blobId"` // Blob ID for files, tree ID for directories
	Mode string `json:"mode"` // File mode
	Type string `json:"type"` // "blob" or "tree"
}

// BuildTreeFromIndex builds a tree object from the staging area index
// Returns the tree ID (which is just a sequential ID for now)
func BuildTreeFromIndex(root string, options InitOptions, treeID int) error {
	db, err := openDB(root, options)
	if err != nil {
		return err
	}
	defer db.Close()

	// Get all index entries
	entries, err := GetIndexEntries(root, options)
	if err != nil {
		return fmt.Errorf("failed to get index entries: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("nothing to commit. Stage changes first with 'gitclone add'")
	}

	// Build tree entries from index
	treeEntries := make([]TreeEntry, 0, len(entries))
	
	// Sort paths for consistent ordering
	paths := make([]string, 0, len(entries))
	for path := range entries {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		entry := entries[path]
		if entry.BlobID == "" {
			continue // Skip empty entries
		}

		// Normalize path (use forward slashes)
		normalizedPath := filepath.ToSlash(path)
		
		treeEntries = append(treeEntries, TreeEntry{
			Path:   normalizedPath,
			BlobID: entry.BlobID,
			Mode:   entry.Mode,
			Type:   "blob",
		})
	}

	// Serialize tree
	treeData, err := json.MarshalIndent(treeEntries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tree: %w", err)
	}

	// Store tree object: objects/tree/<treeId>
	treeKey := fmt.Sprintf("objects/tree/%d", treeID)
	return db.Put(treeKey, treeData)
}

// ReadTree reads a tree object from storage
func ReadTree(root string, options InitOptions, treeID int) ([]TreeEntry, error) {
	db, err := openDB(root, options)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	treeKey := fmt.Sprintf("objects/tree/%d", treeID)
	data, err := db.Get(treeKey)
	if err != nil {
		return nil, fmt.Errorf("tree not found: %d", treeID)
	}

	var entries []TreeEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tree: %w", err)
	}

	return entries, nil
}

// GetBlobContent retrieves blob content by blob ID
func GetBlobContent(root string, options InitOptions, blobID string) ([]byte, error) {
	db, err := openDB(root, options)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	blobKey := fmt.Sprintf("objects/blob/%s", blobID)
	return db.Get(blobKey)
}

