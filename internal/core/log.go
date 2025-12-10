package core

// Log returns commits in reverse chronological order (from HEAD backwards).
func (repo *Repo) Log() []*Commit {
	var history []*Commit
	commit := repo.HEAD

	for commit != nil {
		history = append(history, commit)
		commit = commit.Parent
	}
	return history
}
