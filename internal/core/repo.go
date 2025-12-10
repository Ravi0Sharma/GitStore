package core

// Repo represents a repository
type Repo struct {
	Name         string
	LastCommitID int
}

// NewRepo creates a new in-memory repository
func NewRepo(name string) *Repo {
	return &Repo{
		Name:         name,
		LastCommitID: -1, // Start before first commit
	}
}
