package files

import (
	"fmt"
	"os"
	"path/filepath"

	"gitclone/internal/infra/storage"
	repostorage "gitclone/internal/storage"
)

// Service handles file operations
type Service struct {
	repoBase string
}

// NewService creates a new files service
func NewService(repoBase string) *Service {
	return &Service{
		repoBase: repoBase,
	}
}

// StageFiles stages files for commit
func (s *Service) StageFiles(repoID, path string) error {
	// Open per-repo store
	repoStore, err := storage.NewRepoStore(s.repoBase, repoID)
	if err != nil {
		return err
	}
	defer repoStore.Close()

	repoPath := repoStore.RepoPath()

	oldDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(repoPath); err != nil {
		return err
	}

	// Determine path to stage
	if path == "" {
		path = "."
	}

	// Add to index (handles both single files and directories)
	if err := repostorage.AddToIndexFromStore(repoStore, path); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	return nil
}

// WriteFile writes content to a file in the repository
func (s *Service) WriteFile(repoID, filePath string, content []byte) error {
	// Open per-repo store (to validate repo exists)
	repoStore, err := storage.NewRepoStore(s.repoBase, repoID)
	if err != nil {
		return err
	}
	defer repoStore.Close()

	repoPath := repoStore.RepoPath()
	fullPath := filepath.Join(repoPath, filePath)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

