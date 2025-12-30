package http

import (
	"encoding/json"
	"log"
	"net/http"

	"gitclone/internal/app/repos"
)

// handleRepoBranches handles GET /api/repos/:id/branches
func (s *Server) handleRepoBranches(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate repo exists
	_, err := repos.ResolveRepoPath(s.repoBase, repoID)
	if err != nil {
		log.Printf("DEBUG: handleRepoBranches - repoID=%s, error=%v", repoID, err)
		RespondJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	// Call service
	branches, err := s.branchSvc.ListBranches(repoID)
	if err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Convert to HTTP types
	httpBranches := make([]Branch, len(branches))
	for i, b := range branches {
		httpBranches[i] = Branch{
			Name:      b.Name,
			CreatedAt: b.CreatedAt,
		}
	}

	// Write output
	RespondJSON(w, http.StatusOK, httpBranches)
}

// handleRepoCheckout handles POST /api/repos/:id/checkout
func (s *Server) handleRepoCheckout(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse input
	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate repo exists
	_, err := repos.ResolveRepoPath(s.repoBase, repoID)
	if err != nil {
		log.Printf("DEBUG: handleRepoCheckout - repoID=%s, error=%v", repoID, err)
		RespondJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	// Call service
	if err := s.branchSvc.Checkout(repoID, req.Branch); err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Write output
	RespondJSON(w, http.StatusOK, map[string]string{"message": "Branch checked out successfully"})
}
