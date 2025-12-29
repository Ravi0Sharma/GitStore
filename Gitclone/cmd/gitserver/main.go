package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
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

type AddRequest struct {
	Path string `json:"path"`
}

type CommitRequest struct {
	Message string `json:"message"`
}

type PushRequest struct {
	Remote string `json:"remote"`
	Branch string `json:"branch"`
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

type Issue struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Body         string    `json:"body"`
	Status       string    `json:"status"`   // "open" or "closed"
	Priority     string    `json:"priority"` // "low", "medium", "high"
	Labels       []Label   `json:"labels"`
	Author       string    `json:"author"`
	AuthorAvatar string    `json:"authorAvatar"`
	CreatedAt    time.Time `json:"createdAt"`
	CommentCount int       `json:"commentCount"`
}

type Label struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type CreateIssueRequest struct {
	Title    string  `json:"title"`
	Body     string  `json:"body"`
	Priority string  `json:"priority"`
	Labels   []Label `json:"labels"`
	Author   string  `json:"author,omitempty"` // Optional: email from frontend
}

type FileRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
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
		// Check if it's a specific issue operation (e.g., /api/repos/:id/issues/:issueId)
		if len(parts) >= 3 && parts[2] != "" {
			issueID := parts[2]
			s.handleIssue(w, r, repoID, issueID)
		} else {
			s.handleRepoIssues(w, r, repoID)
		}
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

	// Get branch from query parameter (defaults to current branch if not specified)
	branch := r.URL.Query().Get("branch")
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	commits, err := s.loadCommits(repoPath, branch, limit)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Limit is already applied in loadCommits

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

	// Only GitClone repos are supported
	if !s.isGitCloneRepo(repoPath) {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "GitClone repository (.gitclone) not found.",
		})
		return
	}

	// Check if there are staged files
	opts := storage.InitOptions{Bare: false}
	stagedFiles, err := storage.GetStagedFiles(repoPath, opts)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to check staged files: %v", err)})
		return
	}

	if len(stagedFiles) == 0 {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "Nothing to commit. Stage changes first with 'git add <path>'",
		})
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

	// Use GitClone's commit command
	commands.Commit([]string{"-m", req.Message})

	// Clear staging area after successful commit
	if err := storage.ClearIndex(repoPath, opts); err != nil {
		log.Printf("Warning: failed to clear staging area after commit: %v", err)
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Commit created successfully (local only)",
	})
}

func (s *Server) handleRepoAdd(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	repoPath := filepath.Join(s.repoBase, repoID)
	if !storage.InRepo(repoPath, storage.InitOptions{Bare: false}) {
		respondJSON(w, http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	// Only GitClone repos are supported
	if !s.isGitCloneRepo(repoPath) {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "GitClone repository (.gitclone) not found.",
		})
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

	// Add files to staging area
	path := req.Path
	if path == "" {
		path = "."
	}

	// If path is ".", add all files in repo (excluding .gitclone)
	var pathsToStage []string
	if path == "." {
		// Find all files in repo (excluding .gitclone)
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
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to scan files: %v", err)})
			return
		}
	} else {
		pathsToStage = []string{path}
	}

	// Add to staging area
	opts := storage.InitOptions{Bare: false}
	if err := storage.AddToIndex(repoPath, opts, pathsToStage); err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to stage files: %v", err)})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("Files staged: %s", path)})
}

func (s *Server) handleRepoPush(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	repoPath := filepath.Join(s.repoBase, repoID)
	if !storage.InRepo(repoPath, storage.InitOptions{Bare: false}) {
		respondJSON(w, http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	// Only GitClone repos are supported
	if !s.isGitCloneRepo(repoPath) {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "GitClone repository (.gitclone) not found.",
		})
		return
	}

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

	// Get current branch tip
	tipPtr, err := storage.ReadHeadRefMaybe(repoPath, opts, branch)
	if err != nil || tipPtr == nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "No commits to push"})
		return
	}

	// Get already pushed commits
	pushedCommits, err := storage.GetPushedCommits(repoPath, opts, branch)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to get pushed commits: %v", err)})
		return
	}

	// Find commits that need to be pushed (walk from tip to last pushed commit)
	var commitsToPush []int
	currentID := *tipPtr

	// Walk commits from tip backwards until we hit a pushed commit
	for {
		// Check if this commit is already pushed
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

		// Add to push list
		commitsToPush = append(commitsToPush, currentID)

		// Read commit to get parent
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
		respondJSON(w, http.StatusOK, map[string]string{"message": "Already up to date"})
		return
	}

	// Push commits (mark as pushed)
	for _, commitID := range commitsToPush {
		if err := storage.PushCommit(repoPath, opts, branch, commitID); err != nil {
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to push commit %d: %v", commitID, err)})
			return
		}
	}

	// Update metadata commit count
	meta, err := s.metaStore.GetRepo(repoID)
	if err == nil {
		commits, _ := s.loadCommits(repoPath, branch, 100)
		meta.CommitCount = len(commits)
		meta.UpdatedAt = time.Now()
		if err := s.metaStore.UpdateRepo(*meta); err != nil {
			log.Printf("Warning: failed to update metadata after push: %v", err)
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Pushed %d commit(s) to remote successfully", len(commitsToPush)),
	})
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
		// Get current branch for commit count
		currentBranch, _ := storage.ReadHEADBranch(repoPath, storage.InitOptions{Bare: false})
		commits, _ := s.loadCommits(repoPath, currentBranch, 100)
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
	log.Printf("POST /api/repos - Initializing GitClone repository in: %s", repoPath)
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

// isGitCloneRepo checks if a repository is a GitClone repo (.gitclone/)
func (s *Server) isGitCloneRepo(repoPath string) bool {
	gitclonePath := filepath.Join(repoPath, storage.RepoDir)
	hasGitClone, err := os.Stat(gitclonePath)
	return err == nil && hasGitClone.IsDir()
}

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
	commits, _ := s.loadCommits(repoPath, currentBranch, 100)

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
	commits, _ := s.loadCommits(repoPath, currentBranch, 100)
	issues, _ := s.loadIssues(repoID)

	// Convert issues to []interface{} for Repository struct
	issuesInterface := make([]interface{}, len(issues))
	for i, issue := range issues {
		issuesInterface[i] = issue
	}

	return Repository{
		ID:            repoID,
		Name:          filepath.Base(repoID),
		CurrentBranch: currentBranch,
		Branches:      branches,
		Commits:       commits,
		Issues:        issuesInterface,
	}, nil
}

func (s *Server) loadBranches(repoPath string) ([]Branch, error) {
	opts := storage.InitOptions{Bare: false}

	branchNames, err := storage.ListBranches(repoPath, opts)
	if err != nil {
		return nil, err
	}

	// Deduplicate branches by name (use map to track seen branches)
	seen := make(map[string]bool)
	uniqueNames := make([]string, 0, len(branchNames))
	for _, name := range branchNames {
		if !seen[name] {
			seen[name] = true
			uniqueNames = append(uniqueNames, name)
		}
	}

	branches := make([]Branch, 0, len(uniqueNames))
	for _, name := range uniqueNames {
		branches = append(branches, Branch{
			Name:      name,
			CreatedAt: time.Now().Format(time.RFC3339), // TODO: get actual creation time
		})
	}

	return branches, nil
}

func (s *Server) loadCommits(repoPath string, branchName string, limit int) ([]Commit, error) {
	opts := storage.InitOptions{Bare: false}

	// Only GitClone repos are supported
	if !s.isGitCloneRepo(repoPath) {
		return []Commit{}, nil
	}

	// Use provided branch name, or default to current branch
	var targetBranch string
	if branchName != "" {
		targetBranch = branchName
	} else {
		var err error
		targetBranch, err = storage.ReadHEADBranch(repoPath, opts)
		if err != nil {
			return []Commit{}, nil
		}
	}

	// Get pushed commits for this branch
	pushedCommits, err := storage.GetPushedCommits(repoPath, opts, targetBranch)
	if err != nil {
		return []Commit{}, err
	}

	if len(pushedCommits) == 0 {
		return []Commit{}, nil
	}

	// Create a map for quick lookup
	pushedMap := make(map[int]bool)
	for _, id := range pushedCommits {
		pushedMap[id] = true
	}

	// Read commits using GitClone storage, but only include pushed ones
	tipPtr, err := storage.ReadHeadRefMaybe(repoPath, opts, targetBranch)
	if err != nil || tipPtr == nil {
		return []Commit{}, nil
	}

	var commits []Commit
	id := *tipPtr
	count := 0

	for count < limit {
		c, err := storage.ReadCommitObject(repoPath, opts, id)
		if err != nil {
			break
		}

		// Only include pushed commits
		if pushedMap[c.ID] {
			commits = append(commits, Commit{
				Hash:    fmt.Sprintf("%d", c.ID),
				Message: c.Message,
				Author:  "system", // TODO: get from commit
				Date:    time.Unix(c.Timestamp, 0).Format(time.RFC3339),
			})
			count++
		}

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

func (s *Server) handleRepoIssues(w http.ResponseWriter, r *http.Request, repoID string) {
	repoPath := filepath.Join(s.repoBase, repoID)
	if !storage.InRepo(repoPath, storage.InitOptions{Bare: false}) {
		respondJSON(w, http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	if r.Method == http.MethodGet {
		// Get all issues for the repo
		issues, err := s.loadIssues(repoID)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}
		respondJSON(w, http.StatusOK, issues)
	} else if r.Method == http.MethodPost {
		// Create a new issue
		var req CreateIssueRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
			return
		}

		if req.Title == "" {
			respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Issue title is required"})
			return
		}

		// Get author email from request or default to "system"
		authorEmail := req.Author
		if authorEmail == "" {
			authorEmail = "system"
		}
		// Use initials avatar (unisex) instead of avataaars
		avatarURL := fmt.Sprintf("https://api.dicebear.com/7.x/initials/svg?seed=%s", url.QueryEscape(authorEmail))

		// Create issue
		issue := Issue{
			ID:           fmt.Sprintf("%s-%d", repoID, time.Now().UnixNano()),
			Title:        req.Title,
			Body:         req.Body,
			Status:       "open",
			Priority:     req.Priority,
			Labels:       req.Labels,
			Author:       authorEmail,
			AuthorAvatar: avatarURL,
			CreatedAt:    time.Now(),
			CommentCount: 0,
		}

		// Save issue
		if err := s.saveIssue(repoID, issue); err != nil {
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		respondJSON(w, http.StatusCreated, issue)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleIssue(w http.ResponseWriter, r *http.Request, repoID, issueID string) {
	repoPath := filepath.Join(s.repoBase, repoID)
	if !storage.InRepo(repoPath, storage.InitOptions{Bare: false}) {
		respondJSON(w, http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	if r.Method == http.MethodGet {
		// Get specific issue
		issues, err := s.loadIssues(repoID)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		for _, issue := range issues {
			if issue.ID == issueID {
				respondJSON(w, http.StatusOK, issue)
				return
			}
		}

		respondJSON(w, http.StatusNotFound, ErrorResponse{Error: "Issue not found"})
	} else if r.Method == http.MethodPatch || r.Method == http.MethodPut {
		// Update issue (toggle status or update body)
		issues, err := s.loadIssues(repoID)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		var found bool
		for i := range issues {
			if issues[i].ID == issueID {
				found = true
				// Toggle status if it's a status update
				var updateReq struct {
					Status string `json:"status,omitempty"`
					Body   string `json:"body,omitempty"`
				}
				// Try to decode request body, but default to toggle if empty
				_ = json.NewDecoder(r.Body).Decode(&updateReq)

				// If status is explicitly provided, use it; otherwise toggle
				if updateReq.Status != "" {
					issues[i].Status = updateReq.Status
				} else {
					// Toggle status
					if issues[i].Status == "open" {
						issues[i].Status = "closed"
					} else {
						issues[i].Status = "open"
					}
				}

				// Update body if provided
				if updateReq.Body != "" {
					issues[i].Body = updateReq.Body
				}
				break
			}
		}

		if !found {
			respondJSON(w, http.StatusNotFound, ErrorResponse{Error: "Issue not found"})
			return
		}

		// Save updated issues
		db := s.metaStore.GetDB()
		if db == nil {
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Database not available"})
			return
		}

		key := fmt.Sprintf("repo:%s:issues", repoID)
		data, err := json.Marshal(issues)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		if err := db.Put(key, data); err != nil {
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		// Return updated issue
		for _, issue := range issues {
			if issue.ID == issueID {
				respondJSON(w, http.StatusOK, issue)
				return
			}
		}
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) loadIssues(repoID string) ([]Issue, error) {
	// Use metadata store's db directly
	db := s.metaStore.GetDB()
	if db == nil {
		return []Issue{}, nil
	}

	key := fmt.Sprintf("repo:%s:issues", repoID)
	data, err := db.Get(key)
	if err != nil {
		// No issues yet, return empty array
		return []Issue{}, nil
	}

	var issues []Issue
	if err := json.Unmarshal(data, &issues); err != nil {
		return nil, fmt.Errorf("failed to unmarshal issues: %w", err)
	}

	return issues, nil
}

func (s *Server) saveIssue(repoID string, issue Issue) error {
	// Load existing issues
	issues, err := s.loadIssues(repoID)
	if err != nil {
		return err
	}

	// Add new issue
	issues = append(issues, issue)

	// Save back using metadata store's db
	db := s.metaStore.GetDB()
	if db == nil {
		return fmt.Errorf("database not available")
	}

	key := fmt.Sprintf("repo:%s:issues", repoID)
	data, err := json.Marshal(issues)
	if err != nil {
		return fmt.Errorf("failed to marshal issues: %w", err)
	}

	if err := db.Put(key, data); err != nil {
		return fmt.Errorf("failed to save issues: %w", err)
	}

	return nil
}

func (s *Server) handleRepoFiles(w http.ResponseWriter, r *http.Request, repoID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req FileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	if req.Path == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "File path is required"})
		return
	}

	repoPath := filepath.Join(s.repoBase, repoID)
	if !storage.InRepo(repoPath, storage.InitOptions{Bare: false}) {
		respondJSON(w, http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	// Create full file path
	fullPath := filepath.Join(repoPath, req.Path)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to create directory: %v", err)})
		return
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(req.Content), 0644); err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to write file: %v", err)})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "File created/updated successfully",
		"path":    req.Path,
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
