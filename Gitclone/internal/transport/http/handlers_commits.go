package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"gitclone/internal/app/repos"
)

// handleRepoCommits handles GET /api/repos/:id/commits
func (s *Server) handleRepoCommits(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate repo exists
	_, err := repos.ResolveRepoPath(s.repoBase, repoID)
	if err != nil {
		log.Printf("handleRepoCommits: repoID=%s resolve repo path: %v", repoID, err)
		RespondJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	// Parse query parameters
	branch := r.URL.Query().Get("branch")
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Call service
	commits, err := s.commitSvc.ListCommits(repoID, branch, limit)
	if err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Convert to HTTP types
	httpCommits := make([]Commit, len(commits))
	for i, c := range commits {
		httpCommits[i] = Commit{
			Hash:    c.Hash,
			Message: c.Message,
			Author:  c.Author,
			Date:    c.Date,
		}
	}

	// Write output
	RespondJSON(w, http.StatusOK, httpCommits)
}

// handleRepoCommit handles POST /api/repos/:id/commit
func (s *Server) handleRepoCommit(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse input
	var req CommitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate repo exists
	_, err := repos.ResolveRepoPath(s.repoBase, repoID)
	if err != nil {
		log.Printf("handleRepoCommit: repoID=%s resolve repo path: %v", repoID, err)
		RespondJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	// Call service
	if err := s.commitSvc.CreateCommit(repoID, req.Message); err != nil {
		// Check if it's a business logic error (no staged files)
		// Return 400 (Bad Request) instead of 500 for user errors
		errMsg := err.Error()
		if strings.Contains(errMsg, "Nothing to commit") || strings.Contains(errMsg, "Stage changes first") {
			RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: errMsg})
			return
		}
		// Other errors are server errors
		log.Printf("ERROR handleRepoCommit: repoID=%s, error=%v", repoID, err)
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: errMsg})
		return
	}

	// Write output
	RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Commit created successfully (local only)",
	})
}

// handleRepoPush handles POST /api/repos/:id/push
func (s *Server) handleRepoPush(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse input
	var req PushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate repo exists
	_, err := repos.ResolveRepoPath(s.repoBase, repoID)
	if err != nil {
		log.Printf("handleRepoPush: repoID=%s resolve repo path: %v", repoID, err)
		RespondJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	// Call service
	count, err := s.commitSvc.PushCommits(repoID, req.Branch)
	if err != nil {
		// Check if it's "no commits to push" or "already up to date"
		if err.Error() == "no commits to push" {
			RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Write output
	if count == 0 {
		RespondJSON(w, http.StatusOK, map[string]string{"message": "Already up to date"})
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Pushed %d commit(s) to remote successfully", count),
	})
}
