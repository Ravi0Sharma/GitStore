package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

const RepoDir = ".gitclone"

type InitOptions struct {
	Bare bool
}

// InRepo checks whether the current folder already contains a gitclone repository.
func inRepo(root string, options InitOptions) bool {

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
