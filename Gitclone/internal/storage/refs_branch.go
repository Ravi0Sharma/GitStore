package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func refHeadPath(root string, opts InitOptions, branch string) string {
	return filepath.Join(repoRoot(root, opts), "refs", "heads", branch)
}

// EnsureHeadRefExists creates refs/heads/<branch> if missing.
func EnsureHeadRefExists(root string, opts InitOptions, branch string) error {
	if branch == "" || strings.ContainsAny(branch, " \t\n") {
		return fmt.Errorf("invalid branch name")
	}

	p := refHeadPath(root, opts, branch)

	// Ensure refs/heads directory exists
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}

	// If file exists, do nothing
	if _, err := os.Stat(p); err == nil {
		return nil
	}

	// Create empty ref file
	return os.WriteFile(p, []byte(""), 0o644)
}

// WriteHeadRef writes commit ID into refs/heads/<branch>
func WriteHeadRef(root string, opts InitOptions, branch string, commitID int) error {
	if err := EnsureHeadRefExists(root, opts, branch); err != nil {
		return err
	}
	p := refHeadPath(root, opts, branch)
	return os.WriteFile(p, []byte(fmt.Sprintf("%d\n", commitID)), 0o644)
}

// ReadHeadRef reads commit ID from refs/heads/<branch>
func ReadHeadRef(root string, opts InitOptions, branch string) (int, error) {
	p := refHeadPath(root, opts, branch)
	b, err := os.ReadFile(p)
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(b))
	if s == "" {
		return 0, fmt.Errorf("branch has no commits")
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid commit id in branch ref: %q", s)
	}
	return n, nil
}

// ReadHeadRefMaybe reads commit ID from refs/heads/<branch>.
// Returns nil if branch has no commits (empty ref file).
func ReadHeadRefMaybe(root string, options InitOptions, branch string) (*int, error) {
	p := refHeadPath(root, options, branch)
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	s := strings.TrimSpace(string(b))
	if s == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil, fmt.Errorf("invalid commit id in branch ref: %q", s)
	}
	return &n, nil
}

func ReadHEADBranch(root string, opts InitOptions) (string, error) {
	b, err := os.ReadFile(headFile(root, opts))
	if err != nil {
		return "", err
	}

	head := strings.TrimSpace(string(b))
	const prefix = "ref: refs/heads/"
	if !strings.HasPrefix(head, prefix) {
		return "", fmt.Errorf("invalid HEAD format: %q", head)
	}

	branch := strings.TrimPrefix(head, prefix)
	if err := validateBranch(branch); err != nil {
		return "", err
	}
	return branch, nil
}
