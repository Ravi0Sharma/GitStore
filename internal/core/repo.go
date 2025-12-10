package core

// Repo represents a repository
type Repo struct {
	Name         string
	LastCommitID int
	HEAD         *Commit
}

// NewRepo is a constructor function for creating a repository
func NewRepo(name string) *Repo {
	return &Repo{
		Name:         name,
		LastCommitID: -1, // Start before first commit
		HEAD:         nil,
	}
}

// Commit creates a new commit with an auto-incrementing ID.
func (r *Repo) Commit(message string) *Commit {
	r.LastCommitID++
	return &Commit{
		ID:      r.LastCommitID,
		Message: message,
		Parent:  r.HEAD,
	}
}
