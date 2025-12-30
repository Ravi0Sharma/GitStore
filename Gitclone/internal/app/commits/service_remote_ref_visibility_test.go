package commits

import (
	"os"
	"path/filepath"
	"testing"

	"gitclone/internal/infra/storage"
	"gitclone/internal/metadata"
	repostorage "gitclone/internal/storage"
)

// TestRemoteRefVisibilityAfterPush_Regression verifies that after PushCommits writes
// a remote ref, ListCommits (which opens a new RepoStore) immediately sees it.
// This is a regression test for the GitDb bug where Close() would truncate/rewrite
// the log file, dropping records appended by other handles.
//
// Before the fix: PushCommits would write refs/remotes/origin/master, but when
// ListCommits opened a new RepoStore, it wouldn't see the remote ref because
// the previous RepoStore's Close() truncated the file.
//
// After the fix: Close() only syncs, never rewrites, so all appended records
// are preserved and visible to new handles.
func TestRemoteRefVisibilityAfterPush_Regression(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitstore-remote-ref-visibility-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
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

	// Create commits service
	commitSvc := NewService(repoBase, metaStore)

	// Step 1: Create initial commit on master
	testFile1 := filepath.Join(repoPath, "file1.txt")
	if err := os.WriteFile(testFile1, []byte("file1 content"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}

	repoStore1, err := storage.NewRepoStore(repoBase, repoID)
	if err != nil {
		t.Fatalf("Failed to open RepoStore: %v", err)
	}

	if err := repostorage.AddToIndexFromStore(repoStore1, "file1.txt"); err != nil {
		t.Fatalf("Failed to stage file1: %v", err)
	}
	repoStore1.Close()

	if err := commitSvc.CreateCommit(repoID, "Initial commit"); err != nil {
		t.Fatalf("Failed to create initial commit: %v", err)
	}

	// Step 2: Push master (this writes refs/remotes/origin/master)
	// PushCommits opens its own RepoStore, writes the remote ref, and closes it.
	pushCount, err := commitSvc.PushCommits(repoID, "master")
	if err != nil {
		t.Fatalf("Failed to push master: %v", err)
	}
	if pushCount == 0 {
		t.Fatalf("Expected to push at least 1 commit, got 0")
	}

	// Step 3: Immediately call ListCommits (which opens a NEW RepoStore)
	// This should see the remote ref that was just written by PushCommits.
	// Before the fix, this would fail because the previous RepoStore's Close()
	// would truncate the log file, dropping the remote ref record.
	commits, err := commitSvc.ListCommits(repoID, "master", 10)
	if err != nil {
		t.Fatalf("Failed to list commits: %v", err)
	}

	if len(commits) == 0 {
		t.Fatalf("ListCommits returned 0 commits - remote ref not visible! This indicates the GitDb bug is still present.")
	}

	// Verify we got the expected commit
	if len(commits) < 1 {
		t.Fatalf("Expected at least 1 commit, got %d", len(commits))
	}

	firstCommit := commits[0]
	if firstCommit.Message != "Initial commit" {
		t.Errorf("Expected first commit message 'Initial commit', got: %s", firstCommit.Message)
	}

	// Step 4: Verify remote ref is actually in the DB by reading it directly
	repoStore2, err := storage.NewRepoStore(repoBase, repoID)
	if err != nil {
		t.Fatalf("Failed to open RepoStore for verification: %v", err)
	}
	defer repoStore2.Close()

	remoteTip, err := repostorage.ReadRemoteRefFromStore(repoStore2, "master")
	if err != nil {
		t.Fatalf("Failed to read remote ref: %v", err)
	}
	if remoteTip == nil {
		t.Fatalf("Remote ref should exist after push, but ReadRemoteRefFromStore returned nil")
	}

	// Step 5: Verify the commit ID matches
	headTip, err := repostorage.ReadHeadRefMaybeFromStore(repoStore2, "master")
	if err != nil || headTip == nil {
		t.Fatalf("Failed to read head ref: %v", err)
	}

	if *headTip != *remoteTip {
		t.Errorf("Head tip (%d) should equal remote tip (%d) after push", *headTip, *remoteTip)
	}

	t.Logf("SUCCESS: Remote ref is visible after push. Head tip: %d, Remote tip: %d, Commits listed: %d",
		*headTip, *remoteTip, len(commits))
}

// TestRemoteRefVisibility_MultiplePushes verifies that multiple pushes don't
// cause remote refs to disappear due to Close() truncation.
func TestRemoteRefVisibility_MultiplePushes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitstore-multiple-pushes-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
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

	commitSvc := NewService(repoBase, metaStore)

	// Create and push commit 1
	testFile1 := filepath.Join(repoPath, "file1.txt")
	if err := os.WriteFile(testFile1, []byte("file1"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}

	repoStore, err := storage.NewRepoStore(repoBase, repoID)
	if err != nil {
		t.Fatalf("Failed to open RepoStore: %v", err)
	}
	if err := repostorage.AddToIndexFromStore(repoStore, "file1.txt"); err != nil {
		t.Fatalf("Failed to stage file1: %v", err)
	}
	repoStore.Close()

	if err := commitSvc.CreateCommit(repoID, "Commit 1"); err != nil {
		t.Fatalf("Failed to create commit 1: %v", err)
	}
	if _, err := commitSvc.PushCommits(repoID, "master"); err != nil {
		t.Fatalf("Failed to push commit 1: %v", err)
	}

	// Create and push commit 2
	testFile2 := filepath.Join(repoPath, "file2.txt")
	if err := os.WriteFile(testFile2, []byte("file2"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	repoStore2, err := storage.NewRepoStore(repoBase, repoID)
	if err != nil {
		t.Fatalf("Failed to open RepoStore: %v", err)
	}
	if err := repostorage.AddToIndexFromStore(repoStore2, "file2.txt"); err != nil {
		t.Fatalf("Failed to stage file2: %v", err)
	}
	repoStore2.Close()

	if err := commitSvc.CreateCommit(repoID, "Commit 2"); err != nil {
		t.Fatalf("Failed to create commit 2: %v", err)
	}
	if _, err := commitSvc.PushCommits(repoID, "master"); err != nil {
		t.Fatalf("Failed to push commit 2: %v", err)
	}

	// Verify both commits are visible
	commits, err := commitSvc.ListCommits(repoID, "master", 10)
	if err != nil {
		t.Fatalf("Failed to list commits: %v", err)
	}

	if len(commits) < 2 {
		t.Fatalf("Expected at least 2 commits, got %d. Remote ref may have been dropped by Close() truncation.", len(commits))
	}

	if commits[0].Message != "Commit 2" {
		t.Errorf("Expected first commit to be 'Commit 2', got: %s", commits[0].Message)
	}
	if commits[1].Message != "Commit 1" {
		t.Errorf("Expected second commit to be 'Commit 1', got: %s", commits[1].Message)
	}

	t.Logf("SUCCESS: Both commits visible after multiple pushes. Total commits: %d", len(commits))
}

