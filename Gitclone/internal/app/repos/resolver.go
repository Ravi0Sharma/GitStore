package repos

import (
	"fmt"
	"os"
	"path/filepath"

	"gitclone/internal/storage"
)

// ResolveRepoPath resolves a repository ID to an absolute path and validates
// that the repository exists and contains a .gitclone/ directory.
// Returns the absolute path to the repository root on success, or an error
// if the repository doesn't exist or is invalid.
func ResolveRepoPath(repoBase, repoID string) (string, error) {
	// Construct absolute path
	repoPath := filepath.Join(repoBase, repoID)
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path for repo %s: %w", repoID, err)
	}

	// Validate that the directory exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("repository not found: %s", repoID)
		}
		return "", fmt.Errorf("failed to stat repository path: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("repository path is not a directory: %s", repoID)
	}

	// Validate that it contains .gitclone/
	if !storage.InRepo(absPath, storage.InitOptions{Bare: false}) {
		return "", fmt.Errorf("repository does not contain .gitclone/: %s", repoID)
	}

	return absPath, nil
}

