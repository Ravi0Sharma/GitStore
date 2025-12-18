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
}

// CommitObjectPath returns the path to a commit object file.
func commitObjectPath(root string, options InitOptions, id int) string {
	base := root
	if !options.Bare {
		base = filepath.Join(root, RepoDir)
	}
	return filepath.Join(base, "objects", fmt.Sprintf("%d.json", id))
}
