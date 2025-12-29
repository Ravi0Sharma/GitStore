package http

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"gitclone/internal/app/repos"
	"gitclone/internal/commands"
)

// handleRepoBranches handles GET /api/repos/:id/branches
func (s *Server) handleRepoBranches(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	repoPath, err := repos.ResolveRepoPath(s.repoBase, repoID)
	if err != nil {
		log.Printf("DEBUG: handleRepoBranches - repoID=%s, error=%v", repoID, err)
		RespondJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	log.Printf("DEBUG: handleRepoBranches - repoID=%s, resolvedPath=%s", repoID, repoPath)

	branches, err := s.LoadBranches(repoPath)
	if err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	RespondJSON(w, http.StatusOK, branches)
}

// handleRepoCheckout handles POST /api/repos/:id/checkout
func (s *Server) handleRepoCheckout(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	repoPath, err := repos.ResolveRepoPath(s.repoBase, repoID)
	if err != nil {
		log.Printf("DEBUG: handleRepoCheckout - repoID=%s, error=%v", repoID, err)
		RespondJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	log.Printf("DEBUG: handleRepoCheckout - repoID=%s, resolvedPath=%s", repoID, repoPath)

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

	commands.Checkout([]string{req.Branch})

	meta, err := s.metaStore.GetRepo(repoID)
	if err == nil {
		branches, _ := s.LoadBranches(repoPath)
		meta.CurrentBranch = req.Branch
		meta.BranchCount = len(branches)
		meta.UpdatedAt = time.Now()
		if err := s.metaStore.UpdateRepo(*meta); err != nil {
			log.Printf("Warning: failed to update metadata after checkout: %v", err)
		}
	}

	RespondJSON(w, http.StatusOK, map[string]string{"message": "Branch checked out successfully"})
}
