package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func repoRoot(root string, opts InitOptions) string {
	if opts.Bare {
		return root
	}
	return filepath.Join(root, RepoDir)
}

func headFile(root string, opts InitOptions) string {
	return filepath.Join(repoRoot(root, opts), "HEAD")
}

func branchRefFile(root string, opts InitOptions, branch string) string {
	return filepath.Join(repoRoot(root, opts), "refs", "heads", branch)
}

// Validate branch name (minimal rules for now)
func validateBranch(branch string) error {
	if branch == "" {
		return fmt.Errorf("branch name cannot be empty")
	}
	if strings.ContainsAny(branch, " \t\n") {
		return fmt.Errorf("invalid branch name: contains whitespace")
	}
	if strings.Contains(branch, "..") || strings.Contains(branch, "~") || strings.Contains(branch, "^") || strings.Contains(branch, ":") {
		return fmt.Errorf("invalid branch name: contains illegal characters")
	}
	return nil
}

// WriteHEADBranch writes: "ref: refs/heads/<branch>\n" into HEAD.
func WriteHEADBranch(root string, opts InitOptions, branch string) error {
	if err := validateBranch(branch); err != nil {
		return err
	}
	content := "ref: refs/heads/" + branch + "\n"
	return os.WriteFile(headFile(root, opts), []byte(content), 0o644)
}

// EnsureBranchRefExists creates refs/heads/<branch> if missing.
func EnsureBranchRefExists(root string, opts InitOptions, branch string) error {
	if err := validateBranch(branch); err != nil {
		return err
	}

	refPath := branchRefFile(root, opts, branch)

	// Make sure parent directory exists
	if err := os.MkdirAll(filepath.Dir(refPath), 0o755); err != nil {
		return err
	}

	// If file exists, do nothing
	if _, err := os.Stat(refPath); err == nil {
		return nil
	}

	// Otherwise create empty ref file
	return os.WriteFile(refPath, []byte(""), 0o644)
}
