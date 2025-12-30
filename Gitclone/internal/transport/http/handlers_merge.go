package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gitclone/internal/commands"
	"gitclone/internal/infra/storage"
	repostorage "gitclone/internal/storage"
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

	// Open per-repo store
	repoStore, err := storage.NewRepoStore(s.repoBase, repoID)
	if err != nil {
		log.Printf("DEBUG: handleRepoMerge - repoID=%s, error=%v", repoID, err)
		RespondJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	defer repoStore.Close()

	repoPath := repoStore.RepoPath()
	log.Printf("DEBUG handleRepoMerge: repoID=%s, repoPath=%s", repoID, repoPath)

	currentBranch, err := repostorage.ReadHEADBranchFromStore(repoStore)
	if err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Debug: log HEAD branch and refs before merge
	log.Printf("DEBUG handleRepoMerge: HEAD branch before merge: %s", currentBranch)
	currentTipBefore, _ := repostorage.ReadHeadRefMaybeFromStore(repoStore, currentBranch)
	if currentTipBefore != nil {
		log.Printf("DEBUG handleRepoMerge: refs/heads/%s before merge: %d", currentBranch, *currentTipBefore)
	} else {
		log.Printf("DEBUG handleRepoMerge: refs/heads/%s before merge: (empty)", currentBranch)
	}
	sourceTipBefore, _ := repostorage.ReadHeadRefMaybeFromStore(repoStore, req.Branch)
	if sourceTipBefore != nil {
		log.Printf("DEBUG handleRepoMerge: refs/heads/%s (source) before merge: %d", req.Branch, *sourceTipBefore)
	} else {
		log.Printf("DEBUG handleRepoMerge: refs/heads/%s (source) before merge: (empty)", req.Branch)
	}
	remoteTipBefore, _ := repostorage.ReadRemoteRefFromStore(repoStore, currentBranch)
	if remoteTipBefore != nil {
		log.Printf("DEBUG handleRepoMerge: refs/remotes/origin/%s before merge: %d", currentBranch, *remoteTipBefore)
	} else {
		log.Printf("DEBUG handleRepoMerge: refs/remotes/origin/%s before merge: (empty)", currentBranch)
	}

	if currentBranch == req.Branch {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Cannot merge a branch into itself"})
		return
	}

	if err := repostorage.EnsureHeadRefExistsFromStore(repoStore, currentBranch); err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	if err := repostorage.EnsureHeadRefExistsFromStore(repoStore, req.Branch); err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	currentTip, err := repostorage.ReadHeadRefMaybeFromStore(repoStore, currentBranch)
	if err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	otherTip, err := repostorage.ReadHeadRefMaybeFromStore(repoStore, req.Branch)
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
		isFastForward := s.IsAncestorFromStore(repoStore, *currentTip, *otherTip)
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

	// Determine merge type before merge
	isFastForwardMerge := currentTip == nil || (currentTip != nil && otherTip != nil && s.IsAncestorFromStore(repoStore, *currentTip, *otherTip))
	mergeType := "fast-forward"
	if !isFastForwardMerge && currentTip != nil && otherTip != nil {
		// Will create merge commit
		mergeType = "merge-commit"
	}
	log.Printf("DEBUG handleRepoMerge: merge type will be: %s", mergeType)

	commands.Merge([]string{req.Branch})

	// Debug: log refs after merge
	currentTipAfter, _ := repostorage.ReadHeadRefMaybeFromStore(repoStore, currentBranch)
	if currentTipAfter != nil {
		log.Printf("DEBUG handleRepoMerge: refs/heads/%s after merge: %d (changed: %v)", 
			currentBranch, *currentTipAfter, currentTipBefore == nil || (currentTipBefore != nil && *currentTipBefore != *currentTipAfter))
	} else {
		log.Printf("DEBUG handleRepoMerge: refs/heads/%s after merge: (empty)", currentBranch)
	}
	remoteTipAfter, _ := repostorage.ReadRemoteRefFromStore(repoStore, currentBranch)
	if remoteTipAfter != nil {
		log.Printf("DEBUG handleRepoMerge: refs/remotes/origin/%s after merge: %d (changed: %v)", 
			currentBranch, *remoteTipAfter, remoteTipBefore == nil || (remoteTipBefore != nil && *remoteTipBefore != *remoteTipAfter))
	} else {
		log.Printf("DEBUG handleRepoMerge: refs/remotes/origin/%s after merge: (empty) - NOTE: merge updates local ref only, push required for remote ref", currentBranch)
	}

	// Update metadata (using global store for repo registry)
	meta, err := s.metaStore.GetRepo(repoID)
	if err == nil {
		branches, _ := s.branchSvc.ListBranches(repoID)
		currentBranch, _ := repostorage.ReadHEADBranchFromStore(repoStore)
		commits, _ := s.commitSvc.ListCommits(repoID, currentBranch, 100)
		meta.BranchCount = len(branches)
		meta.CommitCount = len(commits)
		meta.UpdatedAt = time.Now()
		if err := s.metaStore.UpdateRepo(*meta); err != nil {
			log.Printf("Warning: failed to update metadata after merge: %v", err)
		}
	}

	RespondJSON(w, http.StatusOK, map[string]string{"message": "Fast-forward merge completed successfully", "type": "fast-forward"})
}
