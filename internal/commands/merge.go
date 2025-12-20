package commands

import (
	"fmt"
	"gitclone/internal/storage"
	"os"
	"time"
)

func Merge(args []string) {
	if len(args) < 1 {
		fmt.Println("usage: gitclone merge <branch>")
		return
	}
	otherBranch := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	writeOptions := storage.InitOptions{Bare: false}

	// Read current branch from HEAD
	currentBranch, err := storage.ReadHEADBranch(cwd, writeOptions)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Cannot merge a branch into itself
	if currentBranch == otherBranch {
		fmt.Printf("Cannot merge branch %s into itself\n", otherBranch)
		return
	}

	// Ensure both refs exist
	if err := storage.EnsureHeadRefExists(cwd, writeOptions, currentBranch); err != nil {
		fmt.Println("Error:", err)
		return
	}
	if err := storage.EnsureHeadRefExists(cwd, writeOptions, otherBranch); err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Read latest commit of both branches
	currentTip, err := storage.ReadHeadRefMaybe(cwd, writeOptions, currentBranch)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	otherTip, err := storage.ReadHeadRefMaybe(cwd, writeOptions, otherBranch)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// If other branch has no commits, nothing to merge
	if otherTip == nil {
		fmt.Printf("Nothing to merge: branch %s has no commits\n", otherBranch)
		return
	}

	// If current branch has no commits, fast-forward
	if currentTip == nil {
		if err := storage.WriteHeadRef(cwd, writeOptions, currentBranch, *otherTip); err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("Fast-forward: branch %s updated to commit %d\n", currentBranch, *otherTip)
		return
	}

	// Create merge commit with two parents
	mergeID, err := storage.NextCommitID(cwd, writeOptions)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	mergeMessage := fmt.Sprintf("Merge branch %s into %s", otherBranch, currentBranch)
	commit := storage.Commit{
		ID:        mergeID,
		Message:   mergeMessage,
		Branch:    currentBranch,
		Timestamp: time.Now().Unix(),
		Parent:    currentTip,
		Parent2:   otherTip,
	}

	// Write merge commit to disk
	if err := storage.WriteCommitObject(cwd, writeOptions, commit); err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Update current branch ref to point to merge commit
	if err := storage.WriteHeadRef(cwd, writeOptions, currentBranch, mergeID); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("[%s %d] %s\n", currentBranch, mergeID, mergeMessage)
}
