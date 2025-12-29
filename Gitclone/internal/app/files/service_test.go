package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestService_WriteFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "gitstore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test repo directory
	repoDir := filepath.Join(tmpDir, "test-repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("Failed to create repo dir: %v", err)
	}

	svc := NewService(tmpDir)

	// Test writing a file
	err = svc.WriteFile("test-repo", "test.txt", []byte("test content"))
	if err == nil {
		t.Error("Expected error for repo without .gitclone, but got nil")
	}
}

