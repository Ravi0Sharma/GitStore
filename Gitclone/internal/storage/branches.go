package storage

import (
	"GitDb"
	"strings"
)

// ListBranches returns all branch names found in the repository
func ListBranches(root string, opts InitOptions) ([]string, error) {
	db, err := openDB(root, opts)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var branches []string
	
	// Scan for all refs/heads/* keys
	err = db.Scan(func(record GitDb.Record) error {
		if strings.HasPrefix(record.Key, "refs/heads/") {
			branchName := strings.TrimPrefix(record.Key, "refs/heads/")
			branches = append(branches, branchName)
		}
		return nil
	})

	return branches, err
}

