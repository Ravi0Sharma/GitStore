package branches

import (
	"testing"
)

func TestService_ListBranches_EmptyRepo(t *testing.T) {
	// This is a simple test to verify the service structure
	// In a real scenario, we'd set up a test repository
	svc := NewService("/tmp/test-repo", nil)
	
	// This will fail because repo doesn't exist, but tests the structure
	_, err := svc.ListBranches("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent repo")
	}
}

