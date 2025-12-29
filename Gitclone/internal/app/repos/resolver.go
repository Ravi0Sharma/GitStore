package repos

import (
	"fmt"
	"os"
	"path/filepath"

	"gitclone/internal/storage"
)

// ResolveRepoPath resolves a repository ID to an absolute path and validates
// that the repository exists and contains a .gitclone/ directory.
// Returns the absolute path to the repository root, or an error if validation fails.
func ResolveRepoPath(repoBase string, repoID string) (string, error) {
	// Validate repoID doesn't contain path traversal attempts
	if repoID == "" {
		return "", fmt.Errorf("repository ID cannot be empty")
	}
	if filepath.IsAbs(repoID) {
		return "", fmt.Errorf("repository ID must be relative, not absolute")
	}
	if containsPathTraversal(repoID) {
		return "", fmt.Errorf("repository ID contains invalid path characters")
	}

	// Construct absolute path
	repoPath := filepath.Join(repoBase, repoID)
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path for repo %s: %w", repoID, err)
	}

	// Validate that the resolved path is within repoBase (prevent directory traversal)
	repoBaseAbs, err := filepath.Abs(repoBase)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path for repo base: %w", err)
	}

	relPath, err := filepath.Rel(repoBaseAbs, absPath)
	if err != nil {
		return "", fmt.Errorf("repository path outside base directory: %w", err)
	}

	// Check for path traversal (should not contain "..")
	if containsPathTraversal(relPath) {
		return "", fmt.Errorf("repository path outside base directory")
	}

	// Validate repository directory exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("repository not found: %s", repoID)
		}
		return "", fmt.Errorf("failed to stat repository: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("repository path is not a directory: %s", repoID)
	}

	// Validate that .gitclone/ directory exists
	gitclonePath := filepath.Join(absPath, storage.RepoDir)
	gitcloneInfo, err := os.Stat(gitclonePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("repository does not contain .gitclone/ directory: %s", repoID)
		}
		return "", fmt.Errorf("failed to stat .gitclone directory: %w", err)
	}
	if !gitcloneInfo.IsDir() {
		return "", fmt.Errorf(".gitclone is not a directory: %s", repoID)
	}

	return absPath, nil
}

// containsPathTraversal checks if a path contains path traversal sequences
func containsPathTraversal(path string) bool {
	// Check for common path traversal patterns
	return path == ".." || 
		len(path) >= 3 && path[:3] == "../" ||
		len(path) >= 3 && path[len(path)-3:] == "/.." ||
		contains(path, "/../") ||
		contains(path, "\\..\\") ||
		contains(path, "..\\")
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

