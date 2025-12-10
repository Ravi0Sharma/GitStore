package core

import "fmt"

// Repo represents a repository
type Repo struct {
	Name         string
	LastCommitID int
	HEAD         *Branch
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
		Parent:  repo.HEAD.Head,
	}
	// Update HEAD to point to the latest commit
	repo.HEAD.Head = commit
	return commit
}

// Checkout changes HEAD to the specified branch (Creates if it does not exist)
func (repo *Repo) Checkout(branchName string) {
	//if branch exist, change
	if branch, ok := repo.Branches[branchName]; ok {
		repo.HEAD = branch
		fmt.Println("Switched to existing branch:", branchName)
		return
	}
	//Create Branch that points on current commit
	newBranch := &Branch{
		Name: branchName,
		Head: repo.HEAD.Head,
	}
	repo.Branches[branchName] = newBranch
	fmt.Println("Switched to new branch:", branchName)
}
