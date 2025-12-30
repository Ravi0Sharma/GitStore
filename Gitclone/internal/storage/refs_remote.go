package storage

import (
	"fmt"
	"strconv"
	"strings"
)

// ReadRemoteRef reads commit ID from refs/remotes/origin/<branch>
// Returns nil if branch has no remote ref (not pushed yet)
func ReadRemoteRef(root string, opts InitOptions, branch string) (*int, error) {
	db, err := openDB(root, opts)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	key := "refs/remotes/origin/" + branch
	b, err := db.Get(key)
	if err != nil {
		// Remote ref doesn't exist - branch hasn't been pushed yet
		return nil, nil
	}
	
	s := strings.TrimSpace(string(b))
	if s == "" {
		return nil, nil
	}
	
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil, fmt.Errorf("invalid commit id in remote ref: %q", s)
	}
	return &n, nil
}

// WriteRemoteRef writes commit ID into refs/remotes/origin/<branch>
func WriteRemoteRef(root string, opts InitOptions, branch string, commitID int) error {
	db, err := openDB(root, opts)
	if err != nil {
		return err
	}
	defer db.Close()

	key := "refs/remotes/origin/" + branch
	return db.Put(key, []byte(fmt.Sprintf("%d\n", commitID)))
}

// ReadRemoteRefMaybe reads commit ID from refs/remotes/origin/<branch>
// Returns nil if branch has no remote ref (not pushed yet)
func ReadRemoteRefMaybe(root string, options InitOptions, branch string) (*int, error) {
	return ReadRemoteRef(root, options, branch)
}


