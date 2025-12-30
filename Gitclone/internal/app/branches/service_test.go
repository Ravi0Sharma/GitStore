package branches

import (
	"os"
	"path/filepath"
	"testing"

	"GitDb"
	"gitclone/internal/infra/storage"
	"gitclone/internal/metadata"
	repostorage "gitclone/internal/storage"
)

// TestBranchPersistsInEmptyRepo verifies that creating a branch in an empty repo
// persists correctly and is visible in ListBranches
func TestBranchPersistsInEmptyRepo(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "gitstore-branch-test-*")
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

	// Create metadata store and register repo
	metaStore, err := metadata.NewStore(repoBase)
	if err != nil {
		t.Fatalf("Failed to create metadata store: %v", err)
	}
	defer metaStore.Close()

	// Register repo in metadata
	repoMeta := metadata.RepoMeta{
		ID:          repoID,
		Name:        "Test Repo",
		Description: "",
	}
	if err := metaStore.CreateRepo(repoMeta); err != nil {
		t.Fatalf("Failed to register repo: %v", err)
	}

	// Create service
	service := NewService(repoBase, metaStore)

	// Verify initial state: should have master branch
	initialBranches, err := service.ListBranches(repoID)
	if err != nil {
		t.Fatalf("Failed to list initial branches: %v", err)
	}

	masterFound := false
	for _, b := range initialBranches {
		if b.Name == "master" {
			masterFound = true
			break
		}
	}
	if !masterFound {
		t.Errorf("Expected master branch in initial branches, got: %v", initialBranches)
	}

	// Create new branch "testarNU" via Checkout
	branchName := "testarNU"
	if err := service.Checkout(repoID, branchName); err != nil {
		t.Fatalf("Failed to checkout/create branch %s: %v", branchName, err)
	}

	// Immediately list branches - should include both master and testarNU
	branches, err := service.ListBranches(repoID)
	if err != nil {
		t.Fatalf("Failed to list branches after checkout: %v", err)
	}

	if len(branches) < 2 {
		t.Errorf("Expected at least 2 branches (master + %s), got %d: %v", branchName, len(branches), branches)
		// Debug: dump refs/heads/* keys
		dumpRefsHeads(t, repoPath)
		dumpHEAD(t, repoPath)
	}

	masterFound = false
	testarNUFound := false
	for _, b := range branches {
		if b.Name == "master" {
			masterFound = true
		}
		if b.Name == branchName {
			testarNUFound = true
		}
	}

	if !masterFound {
		t.Errorf("Expected master branch, got: %v", branches)
	}
	if !testarNUFound {
		t.Errorf("Expected %s branch, got: %v", branchName, branches)
		dumpRefsHeads(t, repoPath)
		dumpHEAD(t, repoPath)
	}

	// Verify refs/heads/testarNU exists in the repo KV store
	repoStore, err := storage.NewRepoStore(repoBase, repoID)
	if err != nil {
		t.Fatalf("Failed to open RepoStore: %v", err)
	}
	defer repoStore.Close()

	// Check if refs/heads/testarNU exists
	key := "refs/heads/" + branchName
	data, err := repoStore.DB().Get(key)
	if err != nil {
		t.Errorf("refs/heads/%s does not exist in KV store: %v", branchName, err)
		dumpRefsHeads(t, repoPath)
	} else {
		t.Logf("refs/heads/%s exists with value: %q", branchName, string(data))
	}

	// Verify HEAD points to testarNU
	headData, err := repoStore.DB().Get("meta/HEAD")
	if err != nil {
		t.Fatalf("Failed to read HEAD: %v", err)
	}
	headContent := string(headData)
	expectedHEAD := "ref: refs/heads/" + branchName + "\n"
	if headContent != expectedHEAD {
		t.Errorf("HEAD should point to %s, got: %q (expected: %q)", branchName, headContent, expectedHEAD)
	}

	// Re-open a new RepoStore instance (simulate fresh process) and call ListBranches again
	repoStore.Close()
	service2 := NewService(repoBase, metaStore)
	branchesAfterReopen, err := service2.ListBranches(repoID)
	if err != nil {
		t.Fatalf("Failed to list branches after reopen: %v", err)
	}

	if len(branchesAfterReopen) < 2 {
		t.Errorf("After reopen: Expected at least 2 branches, got %d: %v", len(branchesAfterReopen), branchesAfterReopen)
		dumpRefsHeads(t, repoPath)
	}

	testarNUFoundAfterReopen := false
	for _, b := range branchesAfterReopen {
		if b.Name == branchName {
			testarNUFoundAfterReopen = true
			break
		}
	}

	if !testarNUFoundAfterReopen {
		t.Errorf("After reopen: Expected %s branch, got: %v", branchName, branchesAfterReopen)
		dumpRefsHeads(t, repoPath)
		dumpHEAD(t, repoPath)
	}
}

// TestBranchPersistsAfterFileOperations verifies that branch persists after file operations
func TestBranchPersistsAfterFileOperations(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "gitstore-branch-fileops-test-*")
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

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	if err := os.Chdir(repoPath); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	options := repostorage.InitOptions{Bare: false}
	if err := repostorage.InitRepo(repoPath, options); err != nil {
		t.Fatalf("Failed to init repo: %v", err)
	}

	// Create metadata store and register repo
	metaStore, err := metadata.NewStore(repoBase)
	if err != nil {
		t.Fatalf("Failed to create metadata store: %v", err)
	}
	defer metaStore.Close()

	repoMeta := metadata.RepoMeta{
		ID:          repoID,
		Name:        "Test Repo",
		Description: "",
	}
	if err := metaStore.CreateRepo(repoMeta); err != nil {
		t.Fatalf("Failed to register repo: %v", err)
	}

	// Create services
	branchSvc := NewService(repoBase, metaStore)

	// Create new branch "testarNU"
	branchName := "testarNU"
	if err := branchSvc.Checkout(repoID, branchName); err != nil {
		t.Fatalf("Failed to checkout/create branch %s: %v", branchName, err)
	}

	// Verify branch exists before file operations
	branchesBefore, err := branchSvc.ListBranches(repoID)
	if err != nil {
		t.Fatalf("Failed to list branches before file ops: %v", err)
	}

	testarNUFound := false
	for _, b := range branchesBefore {
		if b.Name == branchName {
			testarNUFound = true
			break
		}
	}
	if !testarNUFound {
		t.Errorf("Branch %s should exist before file ops, got: %v", branchName, branchesBefore)
		dumpRefsHeads(t, repoPath)
	}

	// Create a file, stage it, commit it
	testFile := filepath.Join(repoPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Stage file
	repoStore, err := storage.NewRepoStore(repoBase, repoID)
	if err != nil {
		t.Fatalf("Failed to open RepoStore: %v", err)
	}

	if err := repostorage.AddToIndexFromStore(repoStore, "test.txt"); err != nil {
		t.Fatalf("Failed to stage file: %v", err)
	}
	repoStore.Close()

	// Commit (we need commits service for this)
	// For now, just verify branch still exists after staging
	branchesAfterStage, err := branchSvc.ListBranches(repoID)
	if err != nil {
		t.Fatalf("Failed to list branches after stage: %v", err)
	}

	testarNUFoundAfterStage := false
	for _, b := range branchesAfterStage {
		if b.Name == branchName {
			testarNUFoundAfterStage = true
			break
		}
	}
	if !testarNUFoundAfterStage {
		t.Errorf("Branch %s should exist after stage, got: %v", branchName, branchesAfterStage)
		dumpRefsHeads(t, repoPath)
	}

	// Verify refs/heads/testarNU still exists
	repoStore2, err := storage.NewRepoStore(repoBase, repoID)
	if err != nil {
		t.Fatalf("Failed to open RepoStore: %v", err)
	}
	defer repoStore2.Close()

	key := "refs/heads/" + branchName
	_, err = repoStore2.DB().Get(key)
	if err != nil {
		t.Errorf("refs/heads/%s should exist after file operations, got error: %v", branchName, err)
		dumpRefsHeads(t, repoPath)
	}
}

// TestListBranchesFromRefsHeads verifies that ListBranches reads from refs/heads/* keys
func TestListBranchesFromRefsHeads(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitstore-branch-list-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	repoBase := filepath.Join(tmpDir, "repos")
	repoID := "test-repo"
	repoPath := filepath.Join(repoBase, repoID)

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create repo dir: %v", err)
	}

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	if err := os.Chdir(repoPath); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	options := repostorage.InitOptions{Bare: false}
	if err := repostorage.InitRepo(repoPath, options); err != nil {
		t.Fatalf("Failed to init repo: %v", err)
	}

	metaStore, err := metadata.NewStore(repoBase)
	if err != nil {
		t.Fatalf("Failed to create metadata store: %v", err)
	}
	defer metaStore.Close()

	repoMeta := metadata.RepoMeta{
		ID:          repoID,
		Name:        "Test Repo",
		Description: "",
	}
	if err := metaStore.CreateRepo(repoMeta); err != nil {
		t.Fatalf("Failed to register repo: %v", err)
	}

	service := NewService(repoBase, metaStore)

	// Manually create refs/heads/test-branch in the KV store
	repoStore, err := storage.NewRepoStore(repoBase, repoID)
	if err != nil {
		t.Fatalf("Failed to open RepoStore: %v", err)
	}

	// Create ref directly
	key := "refs/heads/test-branch"
	if err := repoStore.DB().Put(key, []byte("")); err != nil {
		t.Fatalf("Failed to create ref directly: %v", err)
	}
	repoStore.Close()

	// ListBranches should include test-branch
	branches, err := service.ListBranches(repoID)
	if err != nil {
		t.Fatalf("Failed to list branches: %v", err)
	}

	testBranchFound := false
	for _, b := range branches {
		if b.Name == "test-branch" {
			testBranchFound = true
			break
		}
	}

	if !testBranchFound {
		t.Errorf("ListBranches should include test-branch (created directly in KV), got: %v", branches)
		dumpRefsHeads(t, repoPath)
	}
}

// Helper functions

func dumpRefsHeads(t *testing.T, repoPath string) {
	t.Helper()
	// Use RepoStore to access DB
	repoBase := filepath.Dir(repoPath)
	repoID := filepath.Base(repoPath)
	repoStore, err := storage.NewRepoStore(repoBase, repoID)
	if err != nil {
		t.Logf("Failed to open RepoStore for dump: %v", err)
		return
	}
	defer repoStore.Close()

	t.Logf("=== Dumping refs/heads/* keys ===")
	db := repoStore.DB()
	err = db.Scan(func(record GitDb.Record) error {
		if len(record.Key) >= 11 && record.Key[:11] == "refs/heads/" {
			t.Logf("  Key: %s, Value: %q", record.Key, string(record.Value))
		}
		return nil
	})
	if err != nil {
		t.Logf("Error scanning DB: %v", err)
	}
}

func dumpHEAD(t *testing.T, repoPath string) {
	t.Helper()
	// Use RepoStore to access DB
	repoBase := filepath.Dir(repoPath)
	repoID := filepath.Base(repoPath)
	repoStore, err := storage.NewRepoStore(repoBase, repoID)
	if err != nil {
		t.Logf("Failed to open RepoStore for HEAD dump: %v", err)
		return
	}
	defer repoStore.Close()

	headData, err := repoStore.DB().Get("meta/HEAD")
	if err != nil {
		t.Logf("HEAD not found or error: %v", err)
	} else {
		t.Logf("=== HEAD value: %q ===", string(headData))
	}
}
