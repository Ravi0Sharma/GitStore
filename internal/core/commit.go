package core

// Commit represents a single commit.
type Commit struct {
	ID      int
	Message string
	Parent  *Commit
}
