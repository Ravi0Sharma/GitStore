package storage

import (
	"fmt"
	"strconv"
	"strings"
)

func NextCommitID(root string, options InitOptions) (int, error) {
	db, err := openDB(root, options)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	// Read current value
	b, err := db.Get("meta/NEXT_COMMIT_ID")
	if err != nil {
		return 0, err
	}

	curStr := strings.TrimSpace(string(b))
	cur, err := strconv.Atoi(curStr)
	if err != nil {
		return 0, fmt.Errorf("invalid NEXT_COMMIT_ID: %q", curStr)
	}

	// Write incremented value
	if err := db.Put("meta/NEXT_COMMIT_ID", []byte(fmt.Sprintf("%d\n", cur+1))); err != nil {
		return 0, err
	}

	return cur, nil
}
