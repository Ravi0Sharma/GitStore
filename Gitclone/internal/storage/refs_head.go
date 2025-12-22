package storage

import (
	"fmt"
	"path/filepath"
	"strings"
)

func repoRoot(root string, opts InitOptions) string {
	if opts.Bare {
		return root
	}
	return filepath.Join(root, RepoDir)
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
	db, err := openDB(root, opts)
	if err != nil {
		return err
	}
	defer db.Close()

	content := "ref: refs/heads/" + branch + "\n"
	return db.Put("meta/HEAD", []byte(content))
}

// EnsureBranchRefExists creates refs/heads/<branch> if missing.
func EnsureBranchRefExists(root string, opts InitOptions, branch string) error {
	if err := validateBranch(branch); err != nil {
		return err
	}

	db, err := openDB(root, opts)
	if err != nil {
		return err
	}
	defer db.Close()

	key := "refs/heads/" + branch
	// Check if key exists
	_, err = db.Get(key)
	if err == nil {
		// Key exists, do nothing
		return nil
	}

	//create empty ref
	return db.Put(key, []byte(""))
}
