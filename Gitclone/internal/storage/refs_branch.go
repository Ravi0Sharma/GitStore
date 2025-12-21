package storage

import (
	"fmt"
	"strconv"
	"strings"
)

// EnsureHeadRefExists creates refs/heads/<branch> if missing.
func EnsureHeadRefExists(root string, opts InitOptions, branch string) error {
	if branch == "" || strings.ContainsAny(branch, " \t\n") {
		return fmt.Errorf("invalid branch name")
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

	// Key doesn't exist, create empty ref
	return db.Put(key, []byte(""))
}

// WriteHeadRef writes commit ID into refs/heads/<branch>
func WriteHeadRef(root string, opts InitOptions, branch string, commitID int) error {
	if err := EnsureHeadRefExists(root, opts, branch); err != nil {
		return err
	}
	db, err := openDB(root, opts)
	if err != nil {
		return err
	}
	defer db.Close()

	key := "refs/heads/" + branch
	return db.Put(key, []byte(fmt.Sprintf("%d\n", commitID)))
}

// ReadHeadRef reads commit ID from refs/heads/<branch>
func ReadHeadRef(root string, opts InitOptions, branch string) (int, error) {
	db, err := openDB(root, opts)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	key := "refs/heads/" + branch
	b, err := db.Get(key)
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
	db, err := openDB(root, options)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	key := "refs/heads/" + branch
	b, err := db.Get(key)
	if err != nil {
		// Note: Assuming key not found means branch doesn't exist (returns nil)
		// GitDb.Get returns error for missing keys, which we treat as nil
		return nil, nil
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
	db, err := openDB(root, opts)
	if err != nil {
		return "", err
	}
	defer db.Close()

	b, err := db.Get("meta/HEAD")
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
