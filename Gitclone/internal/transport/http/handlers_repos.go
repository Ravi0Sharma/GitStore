package http

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitclone/internal/app/repos"
	"gitclone/internal/commands"
	"gitclone/internal/metadata"
	"gitclone/internal/storage"
)

// handleListRepos handles GET /api/repos
func (s *Server) handleListRepos(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET /api/repos - Loading repos from metadata store")

	metaRepos, err := s.metaStore.ListRepos()
	if err != nil {
		log.Printf("GET /api/repos - Error loading from store: %v", err)
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	repoList := make([]RepoListItem, 0, len(metaRepos))
	for _, meta := range metaRepos {
		_, err := repos.ResolveRepoPath(s.repoBase, meta.ID)
		missing := err != nil

		if missing != meta.Missing {
			meta.Missing = missing
			if err := s.metaStore.UpdateRepo(meta); err != nil {
				log.Printf("GET /api/repos - Warning: failed to update missing flag for %s: %v", meta.ID, err)
			}
		}

		lastUpdated := ""
		if !meta.UpdatedAt.IsZero() {
			lastUpdated = meta.UpdatedAt.Format(time.RFC3339)
		}
		repoList = append(repoList, RepoListItem{
			ID:            meta.ID,
			Name:          meta.Name,
			Description:   meta.Description,
			CurrentBranch: meta.CurrentBranch,
			BranchCount:   meta.BranchCount,
			CommitCount:   meta.CommitCount,
			CreatedAt:     meta.CreatedAt,
			UpdatedAt:     meta.UpdatedAt,
			LastUpdated:   lastUpdated,
			Missing:       missing,
		})
	}

	log.Printf("GET /api/repos - Found %d repositories (from metadata store)", len(repoList))
	RespondJSON(w, http.StatusOK, repoList)
}

// handleGetRepo handles GET /api/repos/:id
func (s *Server) handleGetRepo(w http.ResponseWriter, r *http.Request, repoID string) {
	repoPath, err := repos.ResolveRepoPath(s.repoBase, repoID)
	if err != nil {
		log.Printf("handleGetRepo: repoID=%s resolve repo path: %v", repoID, err)
		RespondJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	repo, err := s.LoadRepo(repoPath, repoID)
	if err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	RespondJSON(w, http.StatusOK, repo)
}

// handleCreateRepo handles POST /api/repos
func (s *Server) handleCreateRepo(w http.ResponseWriter, r *http.Request) {
	var req CreateRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("POST /api/repos - Invalid request body: %v", err)
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	log.Printf("POST /api/repos - Creating repo: name=%s, description=%s", req.Name, req.Description)

	if req.Name == "" {
		log.Printf("POST /api/repos - Error: Repository name is required")
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Repository name is required"})
		return
	}

	if strings.Contains(req.Name, "/") || strings.Contains(req.Name, "\\") || strings.Contains(req.Name, "..") {
		log.Printf("POST /api/repos - Error: Invalid characters in name: %s", req.Name)
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Repository name contains invalid characters"})
		return
	}

	repoBaseAbs, err := filepath.Abs(s.repoBase)
	if err != nil {
		log.Printf("POST /api/repos - Error getting absolute path: %v", err)
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	repoPath := filepath.Join(repoBaseAbs, req.Name)
	log.Printf("POST /api/repos - Repo path: %s (base: %s)", repoPath, repoBaseAbs)

	if _, err := os.Stat(repoPath); err == nil {
		log.Printf("POST /api/repos - Error: Repository already exists: %s", repoPath)
		RespondJSON(w, http.StatusConflict, ErrorResponse{Error: "Repository already exists"})
		return
	}

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		log.Printf("POST /api/repos - Error creating directory: %v", err)
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	log.Printf("POST /api/repos - Directory created: %s", repoPath)

	oldDir, err := os.Getwd()
	if err != nil {
		log.Printf("POST /api/repos - Error getting working directory: %v", err)
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(repoPath); err != nil {
		log.Printf("POST /api/repos - Error changing directory: %v", err)
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	log.Printf("POST /api/repos - Initializing GitClone repository in: %s", repoPath)
	commands.Init([]string{})

	gitclonePath := filepath.Join(repoPath, storage.RepoDir)
	if _, err := os.Stat(gitclonePath); err != nil {
		log.Printf("POST /api/repos - WARNING: .gitclone directory not found after init: %s", gitclonePath)
	} else {
		log.Printf("POST /api/repos - Repository initialized successfully: %s", gitclonePath)
	}

	repoSummary, err := s.LoadRepoSummary(repoPath, req.Name)
	if err != nil {
		log.Printf("POST /api/repos - Error loading repo summary: %v", err)
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	meta := metadata.RepoMeta{
		ID:            req.Name,
		Name:          req.Name,
		Description:   req.Description,
		CurrentBranch: repoSummary.CurrentBranch,
		BranchCount:   repoSummary.BranchCount,
		CommitCount:   repoSummary.CommitCount,
		Missing:       false,
	}

	if err := s.metaStore.CreateRepo(meta); err != nil {
		log.Printf("POST /api/repos - Error saving metadata: %v", err)
	}

	lastUpdated := ""
	if !meta.UpdatedAt.IsZero() {
		lastUpdated = meta.UpdatedAt.Format(time.RFC3339)
	}
	repoItem := RepoListItem{
		ID:            meta.ID,
		Name:          meta.Name,
		Description:   meta.Description,
		CurrentBranch: meta.CurrentBranch,
		BranchCount:   meta.BranchCount,
		CommitCount:   meta.CommitCount,
		CreatedAt:     meta.CreatedAt,
		UpdatedAt:     meta.UpdatedAt,
		LastUpdated:   lastUpdated,
		Missing:       false,
	}

	log.Printf("POST /api/repos - Repository created successfully: id=%s, name=%s", repoItem.ID, repoItem.Name)
	RespondJSON(w, http.StatusCreated, repoItem)
}

// handleRepoRoutes routes requests to specific repo endpoints
func (s *Server) handleRepoRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/repos/")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "Repository ID required", http.StatusBadRequest)
		return
	}

	repoID := parts[0]

	if len(parts) == 1 {
		if r.Method == http.MethodGet {
			s.handleGetRepo(w, r, repoID)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	action := parts[1]
	switch action {
	case "branches":
		s.handleRepoBranches(w, r, repoID)
	case "commits":
		s.handleRepoCommits(w, r, repoID)
	case "checkout":
		s.handleRepoCheckout(w, r, repoID)
	case "add":
		s.handleRepoAdd(w, r, repoID)
	case "commit":
		s.handleRepoCommit(w, r, repoID)
	case "push":
		s.handleRepoPush(w, r, repoID)
	case "merge":
		s.handleRepoMerge(w, r, repoID)
	case "files":
		s.handleRepoFiles(w, r, repoID)
	case "issues":
		if len(parts) >= 3 && parts[2] != "" {
			s.handleIssue(w, r, repoID, parts[2])
		} else {
			s.handleRepoIssues(w, r, repoID)
		}
	default:
		http.Error(w, "Invalid endpoint", http.StatusNotFound)
	}
}
