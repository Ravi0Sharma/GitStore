package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Commit represents a single commit stored on disk.
type Commit struct {
	ID        int    `json:"id"`
	Message   string `json:"message"`
	Branch    string `json:"branch"`
	Timestamp int64  `json:"timestamp"`
	Parent    *int   `json:"parent,omitempty"`
	Parent2   *int   `json:"parent2,omitempty"`
}

// CommitObjectPath returns the path to a commit object file.
func commitObjectPath(root string, options InitOptions, id int) string {
	base := root
	if !options.Bare {
		base = filepath.Join(root, RepoDir)
	}
	return filepath.Join(base, "objects", fmt.Sprintf("%d.json", id))
}

// WriteCommitObject serializes a commit as JSON and writes it to disk.

func WriteCommitObject(root string, opts InitOptions, c Commit) error {
	path := commitObjectPath(root, opts, c.ID)

	// Ensure objects/ directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// Encode commit as JSON
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// Write commit file
	return os.WriteFile(path, data, 0644)
}

// ReadCommitObject loads and deserializes a commit from disk.
func ReadCommitObject(root string, opts InitOptions, id int) (Commit, error) {
	path := commitObjectPath(root, opts, id)

	// Read commit file
	data, err := os.ReadFile(path)
	if err != nil {
		return Commit{}, err
	}

	// Decode JSON into Commit struct
	var c Commit
	if err := json.Unmarshal(data, &c); err != nil {
		return Commit{}, err
	}

	return c, nil
}
