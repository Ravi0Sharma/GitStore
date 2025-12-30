package storage

import (
	"encoding/json"
	"fmt"
)

// PushCommit marks a commit as pushed to remote
func PushCommit(root string, options InitOptions, branch string, commitID int) error {
	db, err := openDB(root, options)
	if err != nil {
		return err
	}
	defer db.Close()

	// Read current remote commits for branch
	key := fmt.Sprintf("remote/%s/commits", branch)
	remoteData, err := db.Get(key)
	var pushedCommits []int
	if err != nil {
		// No remote commits yet, create new
		pushedCommits = []int{}
	} else {
		if err := json.Unmarshal(remoteData, &pushedCommits); err != nil {
			return fmt.Errorf("failed to unmarshal remote commits: %w", err)
		}
	}

	// Check if commit is already pushed
	for _, id := range pushedCommits {
		if id == commitID {
			// Already pushed
			return nil
		}
	}

	// Add commit to pushed list
	pushedCommits = append(pushedCommits, commitID)

	// Write back
	data, err := json.Marshal(pushedCommits)
	if err != nil {
		return fmt.Errorf("failed to marshal remote commits: %w", err)
	}

	return db.Put(key, data)
}

// GetPushedCommits returns the list of pushed commit IDs for a branch
func GetPushedCommits(root string, options InitOptions, branch string) ([]int, error) {
	db, err := openDB(root, options)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	key := fmt.Sprintf("remote/%s/commits", branch)
	remoteData, err := db.Get(key)
	if err != nil {
		// No pushed commits
		return []int{}, nil
	}

	var pushedCommits []int
	if err := json.Unmarshal(remoteData, &pushedCommits); err != nil {
		return nil, fmt.Errorf("failed to unmarshal remote commits: %w", err)
	}

	return pushedCommits, nil
}

// IsCommitPushed checks if a commit is pushed
func IsCommitPushed(root string, options InitOptions, branch string, commitID int) (bool, error) {
	pushedCommits, err := GetPushedCommits(root, options, branch)
	if err != nil {
		return false, err
	}

	for _, id := range pushedCommits {
		if id == commitID {
			return true, nil
		}
	}

	return false, nil
}
