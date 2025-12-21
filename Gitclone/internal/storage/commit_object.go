package storage

import (
	"encoding/json"
	"fmt"
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

// WriteCommitObject serializes a commit as JSON and writes it to the database.
func WriteCommitObject(root string, options InitOptions, commit Commit) error {
	db, err := openDB(root, options)
	if err != nil {
		return err
	}
	defer db.Close()

	// Encode commit as JSON
	data, err := json.MarshalIndent(commit, "", "  ")
	if err != nil {
		return err
	}

	// Write commit to DB with key "objects/<id>"
	key := fmt.Sprintf("objects/%d", commit.ID)
	return db.Put(key, data)
}

// ReadCommitObject loads and deserializes a commit from the database.
func ReadCommitObject(root string, opts InitOptions, id int) (Commit, error) {
	db, err := openDB(root, opts)
	if err != nil {
		return Commit{}, err
	}
	defer db.Close()

	// Read commit from DB
	key := fmt.Sprintf("objects/%d", id)
	data, err := db.Get(key)
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
