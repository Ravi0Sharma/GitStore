package storage

import (
	"os"
	"path/filepath"
	"testing"

	"GitDb"
)

func TestAddToIndex(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "gitstore-index-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize repo
	options := InitOptions{Bare: false}
	if err := InitRepo(tmpDir, options); err != nil {
		t.Fatalf("Failed to init repo: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Stage the file
	if err := AddToIndex(tmpDir, options, "test.txt"); err != nil {
		t.Fatalf("Failed to add to index: %v", err)
	}

	// Verify entry exists (GetIndexEntries opens a new DB connection)
	entries, err := GetIndexEntries(tmpDir, options)
	if err != nil {
		t.Fatalf("Failed to get index entries: %v", err)
	}

	if len(entries) == 0 {
		// Debug: Check what's in the database
		debugDB, _ := openDB(tmpDir, options)
		defer debugDB.Close()

		var allKeys []string
		_ = debugDB.Scan(func(record GitDb.Record) error {
			allKeys = append(allKeys, record.Key)
			return nil
		})
		t.Fatalf("No entries found in index. All keys: %v", allKeys)
	}

	entry, ok := entries["test.txt"]
	if !ok {
		t.Fatal("test.txt not found in index")
	}

	if entry.BlobID == "" {
		t.Error("Entry has empty blobId")
	}

	if entry.Mode != "100644" {
		t.Errorf("Expected mode 100644, got %s", entry.Mode)
	}

	// Verify blob exists
	db, err := openDB(tmpDir, options)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	blobKey := "objects/blob/" + entry.BlobID
	blobContent, err := db.Get(blobKey)
	if err != nil {
		t.Fatalf("Blob not found: %v", err)
	}

	if string(blobContent) != "test content" {
		t.Errorf("Blob content mismatch: got %s, want test content", string(blobContent))
	}
}

func TestHasStagedEntries(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "gitstore-hasstaged-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize repo
	options := InitOptions{Bare: false}
	if err := InitRepo(tmpDir, options); err != nil {
		t.Fatalf("Failed to init repo: %v", err)
	}

	// Initially should have no staged entries
	hasStaged, err := HasStagedEntries(tmpDir, options)
	if err != nil {
		t.Fatalf("Failed to check staged entries: %v", err)
	}
	if hasStaged {
		t.Error("Expected no staged entries initially")
	}

	// Create and stage a file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := AddToIndex(tmpDir, options, "test.txt"); err != nil {
		t.Fatalf("Failed to add to index: %v", err)
	}

	// Now should have staged entries
	hasStaged, err = HasStagedEntries(tmpDir, options)
	if err != nil {
		t.Fatalf("Failed to check staged entries: %v", err)
	}
	if !hasStaged {
		t.Error("Expected staged entries after adding file")
	}
}

func TestClearIndex(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "gitstore-clearindex-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize repo
	options := InitOptions{Bare: false}
	if err := InitRepo(tmpDir, options); err != nil {
		t.Fatalf("Failed to init repo: %v", err)
	}

	// Create and stage a file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := AddToIndex(tmpDir, options, "test.txt"); err != nil {
		t.Fatalf("Failed to add to index: %v", err)
	}

	// Verify entry exists
	hasStaged, err := HasStagedEntries(tmpDir, options)
	if err != nil {
		t.Fatalf("Failed to check staged entries: %v", err)
	}
	if !hasStaged {
		t.Fatal("Expected staged entries before clear")
	}

	// Clear index
	if err := ClearIndex(tmpDir, options); err != nil {
		t.Fatalf("Failed to clear index: %v", err)
	}

	// Verify no staged entries
	hasStaged, err = HasStagedEntries(tmpDir, options)
	if err != nil {
		t.Fatalf("Failed to check staged entries: %v", err)
	}
	if hasStaged {
		t.Error("Expected no staged entries after clear")
	}
}

func TestBuildTreeFromIndex(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "gitstore-tree-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize repo
	options := InitOptions{Bare: false}
	if err := InitRepo(tmpDir, options); err != nil {
		t.Fatalf("Failed to init repo: %v", err)
	}

	// Create and stage files
	testFile1 := filepath.Join(tmpDir, "file1.txt")
	testFile2 := filepath.Join(tmpDir, "file2.txt")
	if err := os.WriteFile(testFile1, []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(testFile2, []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := AddToIndex(tmpDir, options, "file1.txt"); err != nil {
		t.Fatalf("Failed to add to index: %v", err)
	}
	if err := AddToIndex(tmpDir, options, "file2.txt"); err != nil {
		t.Fatalf("Failed to add to index: %v", err)
	}

	// Build tree
	treeID := 1
	if err := BuildTreeFromIndex(tmpDir, options, treeID); err != nil {
		t.Fatalf("Failed to build tree: %v", err)
	}

	// Read tree
	tree, err := ReadTree(tmpDir, options, treeID)
	if err != nil {
		t.Fatalf("Failed to read tree: %v", err)
	}

	if len(tree) != 2 {
		t.Errorf("Expected 2 entries in tree, got %d", len(tree))
	}

	// Verify entries
	entryMap := make(map[string]TreeEntry)
	for _, entry := range tree {
		entryMap[entry.Path] = entry
	}

	if _, ok := entryMap["file1.txt"]; !ok {
		t.Error("file1.txt not found in tree")
	}
	if _, ok := entryMap["file2.txt"]; !ok {
		t.Error("file2.txt not found in tree")
	}
}

func TestBuildTreeFromIndex_EmptyIndex(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "gitstore-tree-empty-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize repo
	options := InitOptions{Bare: false}
	if err := InitRepo(tmpDir, options); err != nil {
		t.Fatalf("Failed to init repo: %v", err)
	}

	// Try to build tree from empty index
	treeID := 1
	err = BuildTreeFromIndex(tmpDir, options, treeID)
	if err == nil {
		t.Fatal("Expected error when building tree from empty index")
	}

	if err.Error() != "nothing to commit. Stage changes first with 'gitclone add'" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

