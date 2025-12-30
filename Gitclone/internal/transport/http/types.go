package http

import "time"

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
	Issues        []interface{} `json:"issues"`
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
