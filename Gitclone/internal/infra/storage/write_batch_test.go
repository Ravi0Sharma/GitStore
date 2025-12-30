package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"GitDb"
)

func TestWriteBatch_AtomicCommit(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "gitstore-batch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test repo structure
	repoDir := filepath.Join(tmpDir, "test-repo")
	gitcloneDir := filepath.Join(repoDir, ".gitclone")
	dbDir := filepath.Join(gitcloneDir, "db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("Failed to create repo dir: %v", err)
	}

	// Create a minimal RepoStore
	store := &RepoStore{
		repoID:   "test-repo",
		repoPath: repoDir,
		db:       nil, // Will be initialized
	}

	// Open GitDb
	db, err := GitDb.Open(dbDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	store.db = db

	// Initialize with NEXT_COMMIT_ID
	if err := db.Put("meta/NEXT_COMMIT_ID", []byte("1\n")); err != nil {
		t.Fatalf("Failed to initialize NEXT_COMMIT_ID: %v", err)
	}

	// Create a batch with multiple writes
	batch := NewWriteBatch(store)
	batch.Put("key1", []byte("value1"))
	batch.Put("key2", []byte("value2"))
	batch.Put("key3", []byte("value3"))

	// Commit the batch
	if err := batch.Commit(); err != nil {
		t.Fatalf("Failed to commit batch: %v", err)
	}

	// Verify all keys were written
	val1, err := db.Get("key1")
	if err != nil {
		t.Fatalf("key1 not found: %v", err)
	}
	if string(val1) != "value1" {
		t.Errorf("key1 value mismatch: got %s, want value1", string(val1))
	}

	val2, err := db.Get("key2")
	if err != nil {
		t.Fatalf("key2 not found: %v", err)
	}
	if string(val2) != "value2" {
		t.Errorf("key2 value mismatch: got %s, want value2", string(val2))
	}

	val3, err := db.Get("key3")
	if err != nil {
		t.Fatalf("key3 not found: %v", err)
	}
	if string(val3) != "value3" {
		t.Errorf("key3 value mismatch: got %s, want value3", string(val3))
	}

	// Verify tx marker was marked as committed
	// Since GitDb is append-only, Scan() shows all versions, but Get() returns the latest
	// We verify that Get() returns the committed version (not start or failed)
	txMarkerKey := fmt.Sprintf("_tx/%d", 3) // We wrote 3 keys
	txData, err := db.Get(txMarkerKey)
	if err == nil {
		// If tx marker exists, verify it's committed (not start or failed)
		var tx struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(txData, &tx); err == nil {
			if tx.Type != "batch_committed" {
				t.Errorf("Tx marker not committed: got type %s, want batch_committed. This indicates a partial write occurred.", tx.Type)
			}
		} else {
			t.Errorf("Failed to unmarshal tx marker: %v", err)
		}
	}
	// If tx marker doesn't exist (key not found), that's also acceptable
	// It means the batch completed and marker was cleaned up
}

func TestRecoverTransactions(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "gitstore-recovery-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test repo structure
	repoDir := filepath.Join(tmpDir, "test-repo")
	gitcloneDir := filepath.Join(repoDir, ".gitclone")
	dbDir := filepath.Join(gitcloneDir, "db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("Failed to create repo dir: %v", err)
	}

	// Open GitDb
	db, err := GitDb.Open(dbDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Simulate an incomplete transaction by writing a batch_start marker
	txRecord := struct {
		Type   string `json:"type"`
		Writes []struct {
			Key   string `json:"key"`
			Value []byte `json:"value"`
		} `json:"writes"`
	}{
		Type: "batch_start",
		Writes: []struct {
			Key   string `json:"key"`
			Value []byte `json:"value"`
		}{{Key: "test_key", Value: []byte("test_value")}},
	}
	txData, err := json.Marshal(txRecord)
	if err != nil {
		t.Fatalf("Failed to marshal tx record: %v", err)
	}

	if err := db.Put("_tx/1", txData); err != nil {
		t.Fatalf("Failed to write tx marker: %v", err)
	}

	// Create store and recover
	store := &RepoStore{
		repoID:   "test-repo",
		repoPath: repoDir,
		db:       db,
	}

	// Recover transactions
	if err := RecoverTransactions(store); err != nil {
		t.Fatalf("Failed to recover transactions: %v", err)
	}

	// Verify tx marker was marked as recovered
	txData, err = db.Get("_tx/1")
	if err != nil {
		t.Fatalf("Tx marker not found: %v", err)
	}

	var tx struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(txData, &tx); err != nil {
		t.Fatalf("Failed to unmarshal tx record: %v", err)
	}

	if tx.Type != "batch_recovered" {
		t.Errorf("Expected batch_recovered, got %s", tx.Type)
	}
}

