package files

import (
	"fmt"
	"log"
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

// StageFiles stages files for commit (legacy method, kept for compatibility)
func (s *Service) StageFiles(repoID, path string) error {
	_, _, err := s.StageFilesWithInfo(repoID, path)
	return err
}

// StageFilesWithInfo stages files and returns staged entries info
func (s *Service) StageFilesWithInfo(repoID, path string) (int, []string, error) {
	// Open per-repo store
	repoStore, err := storage.NewRepoStore(s.repoBase, repoID)
	if err != nil {
		return 0, nil, err
	}
	// Note: We close explicitly at the end after verification, not with defer

	repoPath := repoStore.RepoPath()

	oldDir, err := os.Getwd()
	if err != nil {
		return 0, nil, err
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(repoPath); err != nil {
		return 0, nil, err
	}

	// Debug: log repo info - verify DB path
	dbPath := filepath.Join(repoPath, ".gitclone", "db")
	log.Printf("DEBUG StageFiles: repoID=%s, repoBase=%s, repoPath=%s, dbPath=%s, stagingPath=%s", 
		repoID, s.repoBase, repoPath, dbPath, path)
	
	// Verify RepoStore DB path matches expected
	actualDBPath := filepath.Join(repoStore.RepoPath(), ".gitclone", "db")
	log.Printf("DEBUG StageFiles: RepoStore.RepoPath()=%s, actualDBPath=%s", repoStore.RepoPath(), actualDBPath)

	// Determine path to stage
	if path == "" {
		path = "."
	}

	// Get staged entries count before
	entriesBefore, err := repostorage.GetIndexEntriesFromStore(repoStore)
	if err != nil {
		log.Printf("DEBUG StageFiles: error getting entries before: %v", err)
	} else {
		countBefore := len(entriesBefore)
		log.Printf("DEBUG StageFiles: staged entries before: %d", countBefore)
		// Log existing entry keys for debugging
		if countBefore > 0 {
			listed := 0
			for p := range entriesBefore {
				log.Printf("DEBUG StageFiles: existing staged path: %s", p)
				listed++
				if listed >= 3 {
					break
				}
			}
		}
	}

	// Add to index (handles both single files and directories)
	// This writes directly to the DB instance, so writes are immediately visible
	if err := repostorage.AddToIndexFromStore(repoStore, path); err != nil {
		return 0, nil, fmt.Errorf("failed to stage files: %w", err)
	}

	// Verify writes are visible in current DB instance (before closing)
	entriesAfter, err := repostorage.GetIndexEntriesFromStore(repoStore)
	var countAfter int
	if err != nil {
		log.Printf("DEBUG StageFiles: error getting entries after: %v", err)
	} else {
		countAfter = len(entriesAfter)
		countBefore := len(entriesBefore)
		log.Printf("DEBUG StageFiles: staged entries after (before close): %d (added %d)", countAfter, countAfter-countBefore)
		
		// Log newly staged paths for debugging
		if countAfter > countBefore {
			listed := 0
			for p := range entriesAfter {
				if _, exists := entriesBefore[p]; !exists {
					log.Printf("DEBUG StageFiles: newly staged path: %s", p)
					listed++
					if listed >= 5 {
						break
					}
				}
			}
		}
	}

	// Close will sync DB - this ensures writes are persisted to disk
	// GitDb.Close() writes in-memory log to file, which should include all Put() calls
	// The writes are already in the in-memory log from Put(), so Close() just persists them
	// 
	// IMPORTANT: GitDb.Put() writes directly to the log file (append-only), so writes are
	// immediately persisted. Close() writes the in-memory log to file, which should be identical.
	// When a new RepoStore is opened, GitDb.Open() reads the log file and rebuilds the index,
	// so it should see all writes from the previous session.
	log.Printf("DEBUG StageFiles: closing RepoStore, writes should be persisted")
	
	// Explicitly close to ensure writes are synced
	if err := repoStore.Close(); err != nil {
		log.Printf("DEBUG StageFiles: error closing RepoStore: %v", err)
		return 0, nil, fmt.Errorf("failed to close repo store: %w", err)
	}
	
	// Verify writes are persisted by opening a fresh RepoStore and checking entries
	verifyStore, err := storage.NewRepoStore(s.repoBase, repoID)
	if err != nil {
		log.Printf("DEBUG StageFiles: warning - failed to verify writes with new RepoStore: %v", err)
	} else {
		defer verifyStore.Close()
		verifyEntries, err := repostorage.GetIndexEntriesFromStore(verifyStore)
		if err != nil {
			log.Printf("DEBUG StageFiles: warning - failed to verify entries: %v", err)
		} else {
			verifyCount := len(verifyEntries)
			log.Printf("DEBUG StageFiles: verified staged entries after close: %d (expected: %d)", verifyCount, countAfter)
			if verifyCount != countAfter {
				log.Printf("DEBUG StageFiles: ERROR - entry count mismatch! Writes may not be persisted correctly.")
			}
		}
	}
	
	// Collect staged paths for response (limit to first 10 for response size)
	stagedPaths := make([]string, 0, 10)
	if entriesAfter != nil {
		for p := range entriesAfter {
			if len(stagedPaths) < 10 {
				stagedPaths = append(stagedPaths, p)
			}
		}
	}
	
	return countAfter, stagedPaths, nil
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

