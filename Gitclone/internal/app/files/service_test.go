package files

import (
	"os"
	"path/filepath"
	"testing"

	"gitclone/internal/infra/storage"
	repostorage "gitclone/internal/storage"
)

// TestStageThenCommit verifies that files staged via StageFiles are visible to commit
func TestStageThenCommit(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "gitstore-stage-commit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	repoBase := filepath.Join(tmpDir, "repos")
	repoID := "test-repo"
	repoPath := filepath.Join(repoBase, repoID)

	// Initialize repository
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create repo dir: %v", err)
	}

	// Change to repo directory and init
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	if err := os.Chdir(repoPath); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	options := repostorage.InitOptions{Bare: false}
	if err := repostorage.InitRepo(repoPath, options); err != nil {
		t.Fatalf("Failed to init repo: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(repoPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create service
	service := NewService(repoBase)

	// Stage the file
	if err := service.StageFiles(repoID, "test.txt"); err != nil {
		t.Fatalf("Failed to stage file: %v", err)
	}

	// Verify staging succeeded by opening a fresh RepoStore
	repoStore, err := storage.NewRepoStore(repoBase, repoID)
	if err != nil {
		t.Fatalf("Failed to open RepoStore: %v", err)
	}
	defer repoStore.Close()

	// Check staged entries
	entries, err := repostorage.GetIndexEntriesFromStore(repoStore)
	if err != nil {
		t.Fatalf("Failed to get index entries: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("No staged entries found after staging")
	}

	// Verify test.txt is staged
	if _, ok := entries["test.txt"]; !ok {
		t.Errorf("test.txt not found in staged entries. Found: %v", entries)
	}

	// Verify HasStagedEntries returns true
	hasStaged, err := repostorage.HasStagedEntriesFromStore(repoStore)
	if err != nil {
		t.Fatalf("Failed to check staged entries: %v", err)
	}
	if !hasStaged {
		t.Error("HasStagedEntries should return true after staging")
	}
}
