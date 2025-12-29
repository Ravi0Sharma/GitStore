package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gitclone/internal/app/repos"
	"gitclone/internal/commands"
	"gitclone/internal/storage"
)

// handleRepoMerge handles POST /api/repos/:id/merge
func (s *Server) handleRepoMerge(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MergeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	repoPath, err := repos.ResolveRepoPath(s.repoBase, repoID)
	if err != nil {
		log.Printf("DEBUG: handleRepoMerge - repoID=%s, error=%v", repoID, err)
		RespondJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	log.Printf("DEBUG: handleRepoMerge - repoID=%s, resolvedPath=%s", repoID, repoPath)

	opts := storage.InitOptions{Bare: false}

	currentBranch, err := storage.ReadHEADBranch(repoPath, opts)
	if err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	if currentBranch == req.Branch {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Cannot merge a branch into itself"})
		return
	}

	if err := storage.EnsureHeadRefExists(repoPath, opts, currentBranch); err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	if err := storage.EnsureHeadRefExists(repoPath, opts, req.Branch); err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	currentTip, err := storage.ReadHeadRefMaybe(repoPath, opts, currentBranch)
	if err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	otherTip, err := storage.ReadHeadRefMaybe(repoPath, opts, req.Branch)
	if err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	if otherTip == nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("Nothing to merge: branch %s has no commits", req.Branch)})
		return
	}

	if currentTip == nil {
		// Fast-forward merge - proceed
	} else {
		isFastForward := s.IsAncestor(repoPath, opts, *currentTip, *otherTip)
		if !isFastForward {
			RespondJSON(w, http.StatusConflict, ErrorResponse{Error: "Non-fast-forward merge is not allowed"})
			return
		}
	}

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

	commands.Merge([]string{req.Branch})

	meta, err := s.metaStore.GetRepo(repoID)
	if err == nil {
		branches, _ := s.LoadBranches(repoPath)
		currentBranch, _ := storage.ReadHEADBranch(repoPath, storage.InitOptions{Bare: false})
		commits, _ := s.LoadCommits(repoPath, currentBranch, 100)
		meta.BranchCount = len(branches)
		meta.CommitCount = len(commits)
		meta.UpdatedAt = time.Now()
		if err := s.metaStore.UpdateRepo(*meta); err != nil {
			log.Printf("Warning: failed to update metadata after merge: %v", err)
		}
	}

	RespondJSON(w, http.StatusOK, map[string]string{"message": "Fast-forward merge completed successfully", "type": "fast-forward"})
}
