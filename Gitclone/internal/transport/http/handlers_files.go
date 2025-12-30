package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"gitclone/internal/app/repos"
)

// handleRepoAdd handles POST /api/repos/:id/add
func (s *Server) handleRepoAdd(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse input
	var req AddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate repo exists
	_, err := repos.ResolveRepoPath(s.repoBase, repoID)
	if err != nil {
		log.Printf("DEBUG: handleRepoAdd - repoID=%s, error=%v", repoID, err)
		RespondJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	// Call service
	path := req.Path
	if path == "" {
		path = "."
	}
	
	// Stage files and get staged entries info
	stagedCount, stagedPaths, err := s.fileSvc.StageFilesWithInfo(repoID, path)
	if err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Return staged info in response
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message":     fmt.Sprintf("Files staged: %s", path),
		"stagedCount": stagedCount,
		"stagedPaths": stagedPaths,
	})
}

// handleRepoFiles handles POST /api/repos/:id/files
func (s *Server) handleRepoFiles(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse input
	var req FileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate input
	if req.Path == "" {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "File path is required"})
		return
	}

	// Validate repo exists
	_, err := repos.ResolveRepoPath(s.repoBase, repoID)
	if err != nil {
		log.Printf("DEBUG: handleRepoFiles - repoID=%s, error=%v", repoID, err)
		RespondJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	// Call service
	if err := s.fileSvc.WriteFile(repoID, req.Path, []byte(req.Content)); err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Write output
	RespondJSON(w, http.StatusOK, map[string]string{
		"message": "File created/updated successfully",
		"path":    req.Path,
	})
}
