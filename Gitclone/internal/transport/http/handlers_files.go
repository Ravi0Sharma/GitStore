package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"gitclone/internal/app/repos"
	"gitclone/internal/storage"
)

// handleRepoAdd handles POST /api/repos/:id/add
func (s *Server) handleRepoAdd(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	repoPath, err := repos.ResolveRepoPath(s.repoBase, repoID)
	if err != nil {
		log.Printf("DEBUG: handleRepoAdd - repoID=%s, error=%v", repoID, err)
		RespondJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	log.Printf("DEBUG: handleRepoAdd - repoID=%s, resolvedPath=%s", repoID, repoPath)

	oldDir, err := os.Getwd()
	if err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(repoPath); err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	path := req.Path
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
			RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to scan files: %v", err)})
			return
		}
	} else {
		pathsToStage = []string{path}
	}

	opts := storage.InitOptions{Bare: false}
	if err := storage.AddToIndex(repoPath, opts, pathsToStage); err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to stage files: %v", err)})
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("Files staged: %s", path)})
}

// handleRepoFiles handles POST /api/repos/:id/files
func (s *Server) handleRepoFiles(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req FileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	if req.Path == "" {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "File path is required"})
		return
	}

	repoPath, err := repos.ResolveRepoPath(s.repoBase, repoID)
	if err != nil {
		log.Printf("DEBUG: handleRepoFiles - repoID=%s, error=%v", repoID, err)
		RespondJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	log.Printf("DEBUG: handleRepoFiles - repoID=%s, resolvedPath=%s", repoID, repoPath)

	fullPath := filepath.Join(repoPath, req.Path)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to create directory: %v", err)})
		return
	}

	if err := os.WriteFile(fullPath, []byte(req.Content), 0644); err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to write file: %v", err)})
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{
		"message": "File created/updated successfully",
		"path":    req.Path,
	})
}
