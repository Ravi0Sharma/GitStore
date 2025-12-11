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
	// Abort if already a repository
	if InRepo(root, options) {
		return fmt.Errorf("repository already initialized")
	}

	// Build a Git structure directory layout
	gitcloneStructure := map[string]any{
		"HEAD":    "ref: refs/heads/master\n",
		"config":  "[core]\n\tbare = " + strconv.FormatBool(options.Bare) + "\n",
		"objects": map[string]any{},
		"refs": map[string]any{
			"heads": map[string]any{},
		},
	}
	var tree map[string]any
	if options.Bare {
		// bare repo: write structure into the top level
		tree = gitcloneStructure
	} else {
		// non-bare repo: nest everything under .gitclone/
		tree = map[string]any{
			RepoDir: gitcloneStructure,
		}
	}
	// Write the structure to disk
	return WriteFilesFromTree(root, tree)
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
