package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// repoRoot returns the directory that contains HEAD.
// For non-bare repos it's: <root>/.gitclone
// For bare repos it's: <root>
func repoRoot(root string, opts InitOptions) string {
	if opts.Bare {
		return root
	}
	return filepath.Join(root, RepoDir)
}

func headPath(root string, opts InitOptions) string {
	return filepath.Join(repoRoot(root, opts), "HEAD")
}

// WriteHEADBranch persists the current branch into HEAD.
// Example content: "ref: refs/heads/master\n"
func WriteHEADBranch(root string, opts InitOptions, branch string) error {
	if branch == "" {
		return fmt.Errorf("branch name cannot be empty")
	}
	// Minimal validation (keep it simple for now)
	if strings.ContainsAny(branch, " \t\n") {
		return fmt.Errorf("invalid branch name: contains whitespace")
	}

	content := "ref: refs/heads/" + branch + "\n"
	return os.WriteFile(headPath(root, opts), []byte(content), 0o644)
}

// ReadHEADBranch reads HEAD and returns the current branch name if HEAD is a branch ref.
func ReadHEADBranch(root string, opts InitOptions) (string, error) {
	b, err := os.ReadFile(headPath(root, opts))
	if err != nil {
		return "", err
	}
	return ParseHEAD(string(b))
}

// ParseHEAD validates HEAD content and extracts the branch name.
// Expected format: "ref: refs/heads/<branch>\n"
func ParseHEAD(head string) (string, error) {
	head = strings.TrimSpace(head)

	const prefix = "ref: refs/heads/"
	if !strings.HasPrefix(head, prefix) {
		return "", fmt.Errorf("invalid HEAD format: %q", head)
	}

	branch := strings.TrimPrefix(head, prefix)
	if branch == "" {
		return "", fmt.Errorf("invalid HEAD: missing branch name")
	}
	if strings.ContainsAny(branch, " \t\n") {
		return "", fmt.Errorf("invalid HEAD: branch contains whitespace")
	}

	return branch, nil
}
