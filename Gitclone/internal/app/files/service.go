package files

import (
	"fmt"
	"os"
	"path/filepath"

	"gitclone/internal/app/repos"
	"gitclone/internal/storage"
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
	repoPath, err := repos.ResolveRepoPath(s.repoBase, repoID)
	if err != nil {
		return err
	}

	oldDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(repoPath); err != nil {
		return err
	}

	// Determine paths to stage
	if path == "" {
		path = "."
	}

	var pathsToStage []string
	if path == "." {
		err := filepath.Walk(repoPath, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				if info.Name() == storage.RepoDir {
					return filepath.SkipDir
				}
				return nil
			}
			relPath, err := filepath.Rel(repoPath, filePath)
			if err == nil {
				pathsToStage = append(pathsToStage, relPath)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to scan files: %w", err)
		}
	} else {
		pathsToStage = []string{path}
	}

	opts := storage.InitOptions{Bare: false}
	if err := storage.AddToIndex(repoPath, opts, pathsToStage); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	return nil
}

// WriteFile writes content to a file in the repository
func (s *Service) WriteFile(repoID, filePath string, content []byte) error {
	repoPath, err := repos.ResolveRepoPath(s.repoBase, repoID)
	if err != nil {
		return err
	}

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

