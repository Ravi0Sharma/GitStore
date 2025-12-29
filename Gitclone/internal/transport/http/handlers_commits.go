package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"gitclone/internal/app/repos"
	"gitclone/internal/commands"
	"gitclone/internal/storage"
)

// handleRepoCommits handles GET /api/repos/:id/commits
func (s *Server) handleRepoCommits(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	repoPath, err := repos.ResolveRepoPath(s.repoBase, repoID)
	if err != nil {
		log.Printf("DEBUG: handleRepoCommits - repoID=%s, error=%v", repoID, err)
		RespondJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	log.Printf("DEBUG: handleRepoCommits - repoID=%s, resolvedPath=%s", repoID, repoPath)

	branch := r.URL.Query().Get("branch")
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	commits, err := s.LoadCommits(repoPath, branch, limit)
	if err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	RespondJSON(w, http.StatusOK, commits)
}

// handleRepoCommit handles POST /api/repos/:id/commit
func (s *Server) handleRepoCommit(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CommitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	repoPath, err := repos.ResolveRepoPath(s.repoBase, repoID)
	if err != nil {
		log.Printf("DEBUG: handleRepoCommit - repoID=%s, error=%v", repoID, err)
		RespondJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	log.Printf("DEBUG: handleRepoCommit - repoID=%s, resolvedPath=%s", repoID, repoPath)

	opts := storage.InitOptions{Bare: false}
	stagedFiles, err := storage.GetStagedFiles(repoPath, opts)
	if err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to check staged files: %v", err)})
		return
	}

	if len(stagedFiles) == 0 {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "Nothing to commit. Stage changes first with 'git add <path>'",
		})
		return
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

	commands.Commit([]string{"-m", req.Message})

	if err := storage.ClearIndex(repoPath, opts); err != nil {
		log.Printf("Warning: failed to clear staging area after commit: %v", err)
	}

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

	var req PushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	repoPath, err := repos.ResolveRepoPath(s.repoBase, repoID)
	if err != nil {
		log.Printf("DEBUG: handleRepoPush - repoID=%s, error=%v", repoID, err)
		RespondJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	log.Printf("DEBUG: handleRepoPush - repoID=%s, resolvedPath=%s", repoID, repoPath)

	opts := storage.InitOptions{Bare: false}
	branch := req.Branch
	if branch == "" {
		currentBranch, err := storage.ReadHEADBranch(repoPath, opts)
		if err != nil {
			branch = "main"
		} else {
			branch = currentBranch
		}
	}

	tipPtr, err := storage.ReadHeadRefMaybe(repoPath, opts, branch)
	if err != nil || tipPtr == nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "No commits to push"})
		return
	}

	pushedCommits, err := storage.GetPushedCommits(repoPath, opts, branch)
	if err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to get pushed commits: %v", err)})
		return
	}

	var commitsToPush []int
	currentID := *tipPtr

	for {
		isPushed := false
		for _, id := range pushedCommits {
			if id == currentID {
				isPushed = true
				break
			}
		}

		if isPushed {
			break
		}

		commitsToPush = append(commitsToPush, currentID)

		c, err := storage.ReadCommitObject(repoPath, opts, currentID)
		if err != nil {
			break
		}

		if c.Parent == nil {
			break
		}
		currentID = *c.Parent
	}

	if len(commitsToPush) == 0 {
		RespondJSON(w, http.StatusOK, map[string]string{"message": "Already up to date"})
		return
	}

	for _, commitID := range commitsToPush {
		if err := storage.PushCommit(repoPath, opts, branch, commitID); err != nil {
			RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to push commit %d: %v", commitID, err)})
			return
		}
	}

	meta, err := s.metaStore.GetRepo(repoID)
	if err == nil {
		commits, _ := s.LoadCommits(repoPath, branch, 100)
		meta.CommitCount = len(commits)
		meta.UpdatedAt = time.Now()
		if err := s.metaStore.UpdateRepo(*meta); err != nil {
			log.Printf("Warning: failed to update metadata after push: %v", err)
		}
	}

	RespondJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Pushed %d commit(s) to remote successfully", len(commitsToPush)),
	})
}
