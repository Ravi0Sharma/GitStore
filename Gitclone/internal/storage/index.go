package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AddToIndex adds files to the staging area (index)
func AddToIndex(root string, options InitOptions, paths []string) error {
	db, err := openDB(root, options)
	if err != nil {
		return err
	}
	defer db.Close()

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

	for _, path := range paths {
		// Normalize path
		normalizedPath := filepath.Clean(path)
		if !seen[normalizedPath] {
			// Check if file exists
			fullPath := filepath.Join(root, normalizedPath)
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

// GetStagedFiles returns the list of staged files
func GetStagedFiles(root string, options InitOptions) ([]string, error) {
	db, err := openDB(root, options)
	if err != nil {
		return nil, err
	}
	defer db.Close()

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

// ClearIndex clears the staging area
func ClearIndex(root string, options InitOptions) error {
	db, err := openDB(root, options)
	if err != nil {
		return err
	}
	defer db.Close()

	// Write empty array
	data, err := json.Marshal([]string{})
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	return db.Put("index/staged", data)
}
