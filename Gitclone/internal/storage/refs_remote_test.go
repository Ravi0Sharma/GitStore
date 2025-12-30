package storage

import (
	"os"
	"testing"
)

func TestReadWriteRemoteRef(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitstore-remote-ref-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	options := InitOptions{Bare: false}
	if err := InitRepo(tmpDir, options); err != nil {
		t.Fatalf("Failed to init repo: %v", err)
	}

	branch := "main"
	commitID := 42

	// Write remote ref
	if err := WriteRemoteRef(tmpDir, options, branch, commitID); err != nil {
		t.Fatalf("Failed to write remote ref: %v", err)
	}

	// Read remote ref
	readID, err := ReadRemoteRef(tmpDir, options, branch)
	if err != nil {
		t.Fatalf("Failed to read remote ref: %v", err)
	}

	if readID == nil {
		t.Fatal("Remote ref should exist but returned nil")
	}

	if *readID != commitID {
		t.Errorf("Expected commit ID %d, got %d", commitID, *readID)
	}

	// Test reading non-existent branch
	nonExistent, err := ReadRemoteRef(tmpDir, options, "nonexistent")
	if err != nil {
		t.Fatalf("Failed to read non-existent remote ref: %v", err)
	}
	if nonExistent != nil {
		t.Errorf("Expected nil for non-existent branch, got %d", *nonExistent)
	}
}

func TestRemoteRefPath(t *testing.T) {
	// Verify remote refs are stored at correct path
	tmpDir, err := os.MkdirTemp("", "gitstore-remote-ref-path-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	options := InitOptions{Bare: false}
	if err := InitRepo(tmpDir, options); err != nil {
		t.Fatalf("Failed to init repo: %v", err)
	}

	branch := "main"
	commitID := 100

	// Write remote ref
	if err := WriteRemoteRef(tmpDir, options, branch, commitID); err != nil {
		t.Fatalf("Failed to write remote ref: %v", err)
	}

	// Verify it's stored at the correct key
	db, err := openDB(tmpDir, options)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	expectedKey := "refs/remotes/origin/" + branch
	data, err := db.Get(expectedKey)
	if err != nil {
		t.Fatalf("Remote ref key %s not found in DB", expectedKey)
	}

	// Verify content
	content := string(data)
	if content != "100\n" {
		t.Errorf("Expected remote ref content '100\\n', got %q", content)
	}
}

