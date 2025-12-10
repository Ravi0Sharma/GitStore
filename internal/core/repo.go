package core

// Repo represents a repository
type Repo struct {
	Name         string
	LastCommitID int
	HEAD         *Commit
	Branches     map[string]*Branch
}

// NewRepo is a constructor function for creating a repository
func NewRepo(name string) *Repo {
	master := &Branch{
		Name: name,
		Head: nil,
	}

	return &Repo{
		Name:         name,
		LastCommitID: -1, // Start before first commit
		HEAD:         nil,
		Branches: map[string]*Branch{
			"master": master,
		},
	}
}

// Commit creates a new commit with an auto-incrementing ID.
func (repo *Repo) Commit(message string) *Commit {
	repo.LastCommitID++

	commit := &Commit{
		ID:      repo.LastCommitID,
		Message: message,
		Parent:  repo.HEAD,
	}
	// Update HEAD to point to the latest commit
	repo.HEAD = commit
	return commit
}
