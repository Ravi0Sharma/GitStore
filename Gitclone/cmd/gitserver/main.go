package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitclone/internal/commands"
	"gitclone/internal/metadata"
	"gitclone/internal/storage"
)

const defaultPort = "8080"
const defaultRepoBase = "./data/repos"
const defaultDBPath = "./data/db"

type Server struct {
	repoBase  string
	metaStore *metadata.Store
}

func main() {
	repoBase := os.Getenv("GITSTORE_REPO_BASE")
	if repoBase == "" {
		repoBase = defaultRepoBase
	}

	dbPath := os.Getenv("GITSTORE_DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Convert paths to absolute for consistency
	repoBaseAbs, err := filepath.Abs(repoBase)
	if err != nil {
		log.Fatalf("Failed to get absolute path for repo base: %v", err)
	}
	repoBase = repoBaseAbs

	dbPathAbs, err := filepath.Abs(dbPath)
	if err != nil {
		log.Fatalf("Failed to get absolute path for db: %v", err)
	}
	dbPath = dbPathAbs

	// Initialize metadata store
	metaStore, err := metadata.NewStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize metadata store: %v", err)
	}
	defer metaStore.Close()

	server := &Server{
		repoBase:  repoBase,
		metaStore: metaStore,
	}

	log.Printf("Repository base directory (absolute): %s", repoBase)
	log.Printf("Metadata database path (absolute): %s", dbPath)

	// CORS middleware
	corsHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/repos", server.handleListRepos)
	mux.HandleFunc("/api/repos/", server.handleRepoRoutes)

	log.Printf("Starting GitStore server on port %s, repo base: %s", port, repoBase)
	log.Fatal(http.ListenAndServe(":"+port, corsHandler(mux)))
}

// API Response types
// RepoListItem matches the client's RepoListItem interface
type RepoListItem struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	CurrentBranch string    `json:"currentBranch"`
	BranchCount   int       `json:"branchCount"`
	CommitCount   int       `json:"commitCount"`
	CreatedAt     time.Time `json:"createdAt,omitempty"`
	UpdatedAt     time.Time `json:"updatedAt,omitempty"`
	LastUpdated   string    `json:"lastUpdated,omitempty"` // ISO string for client compatibility
	Missing       bool      `json:"missing,omitempty"`     // true if repo folder doesn't exist
}

type Branch struct {
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

type Commit struct {
	Hash    string `json:"hash"`
	Message string `json:"message"`
	Author  string `json:"author"`
	Date    string `json:"date"`
}

type Repository struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description,omitempty"`
	CurrentBranch string        `json:"currentBranch"`
	Branches      []Branch      `json:"branches"`
	Commits       []Commit      `json:"commits"`
	Issues        []interface{} `json:"issues"` // Issues not implemented yet
}

type CheckoutRequest struct {
	Branch string `json:"branch"`
}

type CommitRequest struct {
	Message string `json:"message"`
}

type MergeRequest struct {
	Branch string `json:"branch"`
}

type CreateRepoRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (s *Server) handleListRepos(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		log.Printf("GET /api/repos - Loading repos from metadata store")

		// Load repos from metadata store
		metaRepos, err := s.metaStore.ListRepos()
		if err != nil {
			log.Printf("GET /api/repos - Error loading from store: %v", err)
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		// Convert to RepoListItem and check if folders exist
		repos := make([]RepoListItem, 0, len(metaRepos))
		for _, meta := range metaRepos {
			repoPath := filepath.Join(s.repoBase, meta.ID)
			_, err := os.Stat(repoPath)
			missing := err != nil

			// Update missing flag in metadata if needed
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
			repos = append(repos, RepoListItem{
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

		log.Printf("GET /api/repos - Found %d repositories (from metadata store)", len(repos))
		respondJSON(w, http.StatusOK, repos)
	} else if r.Method == http.MethodPost {
		s.handleCreateRepo(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRepoRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/repos/")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "Repository ID required", http.StatusBadRequest)
		return
	}

	repoID := parts[0]

	// Route based on path parts
	if len(parts) == 1 {
		// /api/repos/:id
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
	case "commit":
		s.handleRepoCommit(w, r, repoID)
	case "merge":
		s.handleRepoMerge(w, r, repoID)
	default:
		http.Error(w, "Invalid endpoint", http.StatusNotFound)
	}
}

func (s *Server) handleGetRepo(w http.ResponseWriter, r *http.Request, repoID string) {
	repoPath := filepath.Join(s.repoBase, repoID)
	if !storage.InRepo(repoPath, storage.InitOptions{Bare: false}) {
		respondJSON(w, http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	repo, err := s.loadRepo(repoPath, repoID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, repo)
}

func (s *Server) handleRepoBranches(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	repoPath := filepath.Join(s.repoBase, repoID)
	if !storage.InRepo(repoPath, storage.InitOptions{Bare: false}) {
		respondJSON(w, http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	branches, err := s.loadBranches(repoPath)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, branches)
}

func (s *Server) handleRepoCommits(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	repoPath := filepath.Join(s.repoBase, repoID)
	if !storage.InRepo(repoPath, storage.InitOptions{Bare: false}) {
		respondJSON(w, http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	commits, err := s.loadCommits(repoPath)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, commits)
}

func (s *Server) handleRepoCheckout(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	repoPath := filepath.Join(s.repoBase, repoID)
	if !storage.InRepo(repoPath, storage.InitOptions{Bare: false}) {
		respondJSON(w, http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	// Change to repo directory temporarily
	oldDir, err := os.Getwd()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(repoPath); err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Call checkout command (this creates branch if it doesn't exist)
	commands.Checkout([]string{req.Branch})

	// Update metadata: refresh branch count and current branch
	meta, err := s.metaStore.GetRepo(repoID)
	if err == nil {
		// Reload branch info
		branches, _ := s.loadBranches(repoPath)
		meta.CurrentBranch = req.Branch
		meta.BranchCount = len(branches)
		meta.UpdatedAt = time.Now()
		if err := s.metaStore.UpdateRepo(*meta); err != nil {
			log.Printf("Warning: failed to update metadata after checkout: %v", err)
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Branch checked out successfully"})
}

func (s *Server) handleRepoCommit(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CommitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	repoPath := filepath.Join(s.repoBase, repoID)
	if !storage.InRepo(repoPath, storage.InitOptions{Bare: false}) {
		respondJSON(w, http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	// Change to repo directory temporarily
	oldDir, err := os.Getwd()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(repoPath); err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Call commit command
	commands.Commit([]string{"-m", req.Message})

	// Update metadata: refresh commit count
	meta, err := s.metaStore.GetRepo(repoID)
	if err == nil {
		commits, _ := s.loadCommits(repoPath)
		meta.CommitCount = len(commits)
		meta.UpdatedAt = time.Now()
		if err := s.metaStore.UpdateRepo(*meta); err != nil {
			log.Printf("Warning: failed to update metadata after commit: %v", err)
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Commit created successfully"})
}

func (s *Server) handleRepoMerge(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MergeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	repoPath := filepath.Join(s.repoBase, repoID)
	if !storage.InRepo(repoPath, storage.InitOptions{Bare: false}) {
		respondJSON(w, http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	opts := storage.InitOptions{Bare: false}

	// Read current branch from HEAD
	currentBranch, err := storage.ReadHEADBranch(repoPath, opts)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Cannot merge a branch into itself
	if currentBranch == req.Branch {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Cannot merge a branch into itself"})
		return
	}

	// Ensure both refs exist
	if err := storage.EnsureHeadRefExists(repoPath, opts, currentBranch); err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	if err := storage.EnsureHeadRefExists(repoPath, opts, req.Branch); err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Read latest commit of both branches
	currentTip, err := storage.ReadHeadRefMaybe(repoPath, opts, currentBranch)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	otherTip, err := storage.ReadHeadRefMaybe(repoPath, opts, req.Branch)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// If other branch has no commit, nothing to merge
	if otherTip == nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("Nothing to merge: branch %s has no commits", req.Branch)})
		return
	}

	// If current branch has no commits, it's a fast-forward
	if currentTip == nil {
		// Fast-forward merge - proceed
	} else {
		// Check if merge is fast-forward: currentTip must be an ancestor of otherTip
		// For fast-forward: otherTip must be ahead of currentTip (currentTip is ancestor of otherTip)
		isFastForward := s.isAncestor(repoPath, opts, *currentTip, *otherTip)
		if !isFastForward {
			// Non-fast-forward merge - reject with 409
			respondJSON(w, http.StatusConflict, ErrorResponse{Error: "Non-fast-forward merge is not allowed"})
			return
		}
	}

	// Change to repo directory temporarily
	oldDir, err := os.Getwd()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(repoPath); err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Call merge command (this will perform the fast-forward merge)
	commands.Merge([]string{req.Branch})

	// Update metadata: refresh branch count and commit count
	meta, err := s.metaStore.GetRepo(repoID)
	if err == nil {
		branches, _ := s.loadBranches(repoPath)
		commits, _ := s.loadCommits(repoPath)
		meta.BranchCount = len(branches)
		meta.CommitCount = len(commits)
		meta.UpdatedAt = time.Now()
		if err := s.metaStore.UpdateRepo(*meta); err != nil {
			log.Printf("Warning: failed to update metadata after merge: %v", err)
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Fast-forward merge completed successfully", "type": "fast-forward"})
}

func (s *Server) handleCreateRepo(w http.ResponseWriter, r *http.Request) {
	var req CreateRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("POST /api/repos - Invalid request body: %v", err)
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	log.Printf("POST /api/repos - Creating repo: name=%s, description=%s", req.Name, req.Description)

	if req.Name == "" {
		log.Printf("POST /api/repos - Error: Repository name is required")
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Repository name is required"})
		return
	}

	// Validate repository name (no path separators, no special chars that could cause issues)
	if strings.Contains(req.Name, "/") || strings.Contains(req.Name, "\\") || strings.Contains(req.Name, "..") {
		log.Printf("POST /api/repos - Error: Invalid characters in name: %s", req.Name)
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Repository name contains invalid characters"})
		return
	}

	// Create directory for the new repository
	// Use absolute path to ensure consistency with scanRepos
	repoBaseAbs, err := filepath.Abs(s.repoBase)
	if err != nil {
		log.Printf("POST /api/repos - Error getting absolute path: %v", err)
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	repoPath := filepath.Join(repoBaseAbs, req.Name)
	log.Printf("POST /api/repos - Repo path: %s (base: %s)", repoPath, repoBaseAbs)

	// Check if directory already exists
	if _, err := os.Stat(repoPath); err == nil {
		log.Printf("POST /api/repos - Error: Repository already exists: %s", repoPath)
		respondJSON(w, http.StatusConflict, ErrorResponse{Error: "Repository already exists"})
		return
	}

	// Create the directory
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		log.Printf("POST /api/repos - Error creating directory: %v", err)
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	log.Printf("POST /api/repos - Directory created: %s", repoPath)

	// Change to repo directory temporarily
	oldDir, err := os.Getwd()
	if err != nil {
		log.Printf("POST /api/repos - Error getting working directory: %v", err)
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(repoPath); err != nil {
		log.Printf("POST /api/repos - Error changing directory: %v", err)
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Initialize the repository
	log.Printf("POST /api/repos - Initializing git repository in: %s", repoPath)
	commands.Init([]string{})

	// Verify that .gitclone directory was created
	gitclonePath := filepath.Join(repoPath, storage.RepoDir)
	if _, err := os.Stat(gitclonePath); err != nil {
		log.Printf("POST /api/repos - WARNING: .gitclone directory not found after init: %s", gitclonePath)
	} else {
		log.Printf("POST /api/repos - Repository initialized successfully: %s", gitclonePath)
	}

	// Load repo summary to get branch/commit counts
	repoSummary, err := s.loadRepoSummary(repoPath, req.Name)
	if err != nil {
		log.Printf("POST /api/repos - Error loading repo summary: %v", err)
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Create metadata entry
	meta := metadata.RepoMeta{
		ID:            req.Name,
		Name:          req.Name,
		Description:   req.Description,
		CurrentBranch: repoSummary.CurrentBranch,
		BranchCount:   repoSummary.BranchCount,
		CommitCount:   repoSummary.CommitCount,
		Missing:       false,
	}

	// Save to metadata store
	if err := s.metaStore.CreateRepo(meta); err != nil {
		log.Printf("POST /api/repos - Error saving metadata: %v", err)
		// Continue anyway - repo is created on disk
	}

	// Return RepoListItem matching client format
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
	respondJSON(w, http.StatusCreated, repoItem)
}

// Helper functions

func (s *Server) scanRepos() ([]RepoListItem, error) {
	// Initialize as empty slice (not nil) to ensure JSON returns [] instead of null
	repos := make([]RepoListItem, 0)

	err := filepath.Walk(s.repoBase, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() {
			return nil
		}
		// Check if this directory contains a .gitclone subdirectory
		gitclonePath := filepath.Join(path, storage.RepoDir)
		if _, err := os.Stat(gitclonePath); err == nil {
			// Found a repo
			relPath, _ := filepath.Rel(s.repoBase, path)
			if relPath == "." {
				relPath = filepath.Base(path)
			}
			repo, err := s.loadRepoSummary(path, relPath)
			if err == nil {
				repos = append(repos, repo)
				log.Printf("scanRepos: Found repo %s at %s", repo.ID, path)
			} else {
				log.Printf("scanRepos: Failed to load repo summary for %s: %v", path, err)
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("scanRepos: filepath.Walk error: %v", err)
		return repos, err
	}

	return repos, nil
}

func (s *Server) loadRepoSummary(repoPath, repoID string) (RepoListItem, error) {
	opts := storage.InitOptions{Bare: false}

	currentBranch, _ := storage.ReadHEADBranch(repoPath, opts)
	branches, _ := s.loadBranches(repoPath)
	commits, _ := s.loadCommits(repoPath)

	return RepoListItem{
		ID:            repoID,
		Name:          filepath.Base(repoID),
		CurrentBranch: currentBranch,
		BranchCount:   len(branches),
		CommitCount:   len(commits),
	}, nil
}

func (s *Server) loadRepo(repoPath, repoID string) (Repository, error) {
	opts := storage.InitOptions{Bare: false}

	currentBranch, _ := storage.ReadHEADBranch(repoPath, opts)
	branches, _ := s.loadBranches(repoPath)
	commits, _ := s.loadCommits(repoPath)

	return Repository{
		ID:            repoID,
		Name:          filepath.Base(repoID),
		CurrentBranch: currentBranch,
		Branches:      branches,
		Commits:       commits,
		Issues:        []interface{}{}, // Issues not implemented
	}, nil
}

func (s *Server) loadBranches(repoPath string) ([]Branch, error) {
	opts := storage.InitOptions{Bare: false}

	branchNames, err := storage.ListBranches(repoPath, opts)
	if err != nil {
		return nil, err
	}

	branches := make([]Branch, 0, len(branchNames))
	for _, name := range branchNames {
		branches = append(branches, Branch{
			Name:      name,
			CreatedAt: time.Now().Format(time.RFC3339), // TODO: get actual creation time
		})
	}

	return branches, nil
}

func (s *Server) loadCommits(repoPath string) ([]Commit, error) {
	opts := storage.InitOptions{Bare: false}
	currentBranch, err := storage.ReadHEADBranch(repoPath, opts)
	if err != nil {
		return []Commit{}, nil
	}

	tipPtr, err := storage.ReadHeadRefMaybe(repoPath, opts, currentBranch)
	if err != nil || tipPtr == nil {
		return []Commit{}, nil
	}

	var commits []Commit
	id := *tipPtr

	for {
		c, err := storage.ReadCommitObject(repoPath, opts, id)
		if err != nil {
			break
		}

		commits = append(commits, Commit{
			Hash:    fmt.Sprintf("%d", c.ID),
			Message: c.Message,
			Author:  "system", // TODO: get from commit
			Date:    time.Unix(c.Timestamp, 0).Format(time.RFC3339),
		})

		if c.Parent == nil {
			break
		}
		id = *c.Parent
	}

	return commits, nil
}

// isAncestor checks if commitA is an ancestor of commitB (i.e., commitA is reachable from commitB)
func (s *Server) isAncestor(repoPath string, opts storage.InitOptions, commitA, commitB int) bool {
	// If they're the same, it's trivially an ancestor
	if commitA == commitB {
		return true
	}

	// Walk backwards from commitB following parent pointers
	// If we reach commitA, then commitA is an ancestor of commitB
	visited := make(map[int]bool)
	queue := []int{commitB}
	maxDepth := 1000 // Safety limit to prevent infinite loops
	depth := 0

	for len(queue) > 0 && depth < maxDepth {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true
		depth++

		if current == commitA {
			return true
		}

		// Read commit and add parents to queue
		commit, err := storage.ReadCommitObject(repoPath, opts, current)
		if err != nil {
			// If we can't read the commit, stop searching
			break
		}

		if commit.Parent != nil {
			queue = append(queue, *commit.Parent)
		}
		// Note: We only follow Parent, not Parent2, for fast-forward detection
		// Parent2 would be from a previous merge, which breaks the linear history
	}

	return false
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
