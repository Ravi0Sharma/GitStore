package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const RepoDir = ".gitclone"

type InitOptions struct {
	Bare bool
}

// InRepo checks whether the current folder already contains a gitclone repository.
func InRepo(root string, options InitOptions) bool {

	if options.Bare {
		// Bare repo: check for repository files directly in the root.
		if _, err := os.Stat(filepath.Join(root, "HEAD")); err == nil {
			return true
		}
		if _, err := os.Stat(filepath.Join(root, "objects")); err == nil {
			return true
		}
		return false
	}
	// For normal repos, check if the .gitclone directory exists.
	_, err := os.Stat(filepath.Join(root, RepoDir))
	return err == nil
}

// InitRepo initializes the current directory as a new repository.
func InitRepo(root string, options InitOptions) error {
	if InRepo(root, options) {
		return fmt.Errorf("repository already initialized")
	}

	// Create directory structure (config file and objects directory still needed)
	gitcloneStructure := map[string]any{
		"config":  "[core]\n\tbare = " + strconv.FormatBool(options.Bare) + "\n",
		"objects": map[string]any{},
	}
	var tree map[string]any
	if options.Bare {
		//write structure into the top level
		tree = gitcloneStructure
	} else {
		// nest everything under .gitclone/
		tree = map[string]any{
			RepoDir: gitcloneStructure,
		}
	}
	// Write the structure to disk
	if err := WriteFilesFromTree(root, tree); err != nil {
		return err
	}

	// Initialize DB with metadata keys
	db, err := openDB(root, options)
	if err != nil {
		return err
	}
	defer db.Close()

	// Initialize HEAD
	if err := db.Put("meta/HEAD", []byte("ref: refs/heads/master\n")); err != nil {
		return fmt.Errorf("failed to initialize HEAD: %w", err)
	}

	// Initialize NEXT_COMMIT_ID
	if err := db.Put("meta/NEXT_COMMIT_ID", []byte("0\n")); err != nil {
		return fmt.Errorf("failed to initialize NEXT_COMMIT_ID: %w", err)
	}

	// Initialize master branch ref (empty for new branch)
	if err := db.Put("refs/heads/master", []byte("")); err != nil {
		return fmt.Errorf("failed to initialize master ref: %w", err)
	}

	return nil
}

// WriteFilesFromTree writes a nested file/directory structure to disk.
func WriteFilesFromTree(root string, tree map[string]any) error {
	for name, val := range tree {
		path := filepath.Join(root, name)

		switch v := val.(type) {

		case string:
			// Leaf file with string contents
			if err := os.WriteFile(path, []byte(v), 0644); err != nil {
				return err
			}

		case map[string]any:
			// Directory: create dir and recurse
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
			if err := WriteFilesFromTree(path, v); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unsupported node type for %s", path)
		}
	}
	return nil
}
