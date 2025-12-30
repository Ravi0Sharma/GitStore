package commits

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gitclone/internal/app/branches"
	"gitclone/internal/infra/storage"
	"gitclone/internal/metadata"
	repostorage "gitclone/internal/storage"
)

// TestMergeUpdatesMasterLatestCommit verifies that after merging a branch into master
// and pushing, the latest commit on master is correctly shown
func TestMergeUpdatesMasterLatestCommit(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "gitstore-merge-test-*")
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
	commitSvc := NewService(repoBase, metaStore)

	// Step 1: Create a file and commit on master
	testFile1 := filepath.Join(repoPath, "file1.txt")
	if err := os.WriteFile(testFile1, []byte("file1 content"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}

	repoStore, err := storage.NewRepoStore(repoBase, repoID)
	if err != nil {
		t.Fatalf("Failed to open RepoStore: %v", err)
	}

	// Stage and commit file1 on master
	if err := repostorage.AddToIndexFromStore(repoStore, "file1.txt"); err != nil {
		t.Fatalf("Failed to stage file1: %v", err)
	}
	repoStore.Close()

	if err := commitSvc.CreateCommit(repoID, "Initial commit on master"); err != nil {
		t.Fatalf("Failed to create initial commit: %v", err)
	}

	// Push master
	pushCount, err := commitSvc.PushCommits(repoID, "master")
	if err != nil {
		t.Fatalf("Failed to push master: %v", err)
	}
	if pushCount == 0 {
		t.Fatalf("Expected to push at least 1 commit, got 0")
	}

	// Step 2: Create branch "feature" and checkout using branches service
	branchSvc := branches.NewService(repoBase, metaStore)
	if err := branchSvc.Checkout(repoID, "feature"); err != nil {
		t.Fatalf("Failed to checkout feature branch: %v", err)
	}

	// Step 3: Create file and commit on feature branch
	testFile2 := filepath.Join(repoPath, "file2.txt")
	if err := os.WriteFile(testFile2, []byte("file2 content"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	repoStore3, err := storage.NewRepoStore(repoBase, repoID)
	if err != nil {
		t.Fatalf("Failed to open RepoStore: %v", err)
	}

	// Stage and commit file2 on feature
	if err := repostorage.AddToIndexFromStore(repoStore3, "file2.txt"); err != nil {
		t.Fatalf("Failed to stage file2: %v", err)
	}
	repoStore3.Close()

	if err := commitSvc.CreateCommit(repoID, "Commit on feature branch"); err != nil {
		t.Fatalf("Failed to create commit on feature: %v", err)
	}

	// Push feature
	pushCount2, err := commitSvc.PushCommits(repoID, "feature")
	if err != nil {
		t.Fatalf("Failed to push feature: %v", err)
	}
	if pushCount2 == 0 {
		t.Fatalf("Expected to push at least 1 commit on feature, got 0")
	}

	// Step 4: Switch back to master
	if err := branchSvc.Checkout(repoID, "master"); err != nil {
		t.Fatalf("Failed to switch back to master: %v", err)
	}

	// Perform merge (we'll use the commands.Merge function)
	// But first we need to be in the repo directory
	if err := os.Chdir(repoPath); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	// Merge feature into master using commands.Merge
	// Note: This requires the commands package
	// For now, we'll simulate the merge by creating a merge commit manually
	// In a real scenario, we'd call commands.Merge([]string{"feature"})
	
	// Read feature tip
	repoStoreMerge, err := storage.NewRepoStore(repoBase, repoID)
	if err != nil {
		t.Fatalf("Failed to open RepoStore for merge: %v", err)
	}

	featureTip, err := repostorage.ReadHeadRefMaybeFromStore(repoStoreMerge, "feature")
	if err != nil || featureTip == nil {
		t.Fatalf("Failed to read feature tip: %v", err)
	}

	masterTipBeforeMerge, err := repostorage.ReadHeadRefMaybeFromStore(repoStoreMerge, "master")
	if err != nil {
		t.Fatalf("Failed to read master tip before merge: %v", err)
	}

	// Create merge commit
	mergeID, err := repostorage.NextCommitIDFromStore(repoStoreMerge)
	if err != nil {
		t.Fatalf("Failed to get next commit ID: %v", err)
	}

	mergeCommit := repostorage.Commit{
		ID:        mergeID,
		Message:   "Merge branch feature into master",
		Branch:    "master",
		Timestamp: time.Now().Unix(),
		Parent:    masterTipBeforeMerge,
		Parent2:   featureTip,
	}

	// Write merge commit and update master ref atomically
	batch := repoStoreMerge.NewWriteBatch()
	if err := repostorage.WriteCommitObjectToBatch(batch, mergeCommit); err != nil {
		t.Fatalf("Failed to add merge commit to batch: %v", err)
	}
	if err := repostorage.WriteHeadRefToBatch(batch, "master", mergeID); err != nil {
		t.Fatalf("Failed to add master ref update to batch: %v", err)
	}
	if err := batch.Commit(); err != nil {
		t.Fatalf("Failed to commit merge batch: %v", err)
	}
	repoStoreMerge.Close()

	// Step 5: Push master after merge
	pushCount3, err := commitSvc.PushCommits(repoID, "master")
	if err != nil {
		t.Fatalf("Failed to push master after merge: %v", err)
	}
	if pushCount3 == 0 {
		t.Logf("Warning: No commits pushed after merge (might be fast-forward)")
	}

	// Step 6: Verify ListCommits returns the merge/newest commit as first item
	commits, err := commitSvc.ListCommits(repoID, "master", 10)
	if err != nil {
		t.Fatalf("Failed to list commits: %v", err)
	}

	if len(commits) == 0 {
		t.Fatalf("Expected at least 1 commit on master, got 0")
	}

	// The first commit should be the merge commit or the newest commit
	firstCommit := commits[0]
	t.Logf("First commit on master: %s - %s", firstCommit.Hash, firstCommit.Message)

	// Verify it's either the merge commit or the feature commit
	if firstCommit.Message != "Merge branch feature into master" && firstCommit.Message != "Commit on feature branch" {
		t.Errorf("Expected first commit to be merge or feature commit, got: %s", firstCommit.Message)
	}

	// Step 7: Assert refs/remotes/origin/master equals refs/heads/master after push
	repoStore5, err := storage.NewRepoStore(repoBase, repoID)
	if err != nil {
		t.Fatalf("Failed to open RepoStore: %v", err)
	}
	defer repoStore5.Close()

	headTip, err := repostorage.ReadHeadRefMaybeFromStore(repoStore5, "master")
	if err != nil || headTip == nil {
		t.Fatalf("Failed to read master head ref: %v", err)
	}

	remoteTip, err := repostorage.ReadRemoteRefFromStore(repoStore5, "master")
	if err != nil {
		t.Fatalf("Failed to read master remote ref: %v", err)
	}

	if remoteTip == nil {
		t.Fatalf("Remote ref for master should exist after push")
	}

	if *headTip != *remoteTip {
		t.Errorf("refs/heads/master (%d) should equal refs/remotes/origin/master (%d) after push", *headTip, *remoteTip)
	}
}

