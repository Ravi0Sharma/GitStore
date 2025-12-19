package commands

import (
	"fmt"
	"os"
	"time"

	"gitclone/internal/storage"
)

func Commit(args []string) {
	msg := ""
	for i := 0; i < len(args); i++ {
		if args[i] == "-m" && i+1 < len(args) {
			msg = args[i+1]
		}
	}
	if msg == "" {
		fmt.Println("usage: gitclone commit -m \"message\"")
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	options := storage.InitOptions{Bare: false} // keep simple for now

	branch, err := storage.ReadHEADBranch(cwd, options)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Read current branch tip to determine parent commit
	parentPtr, err := storage.ReadHeadRefMaybe(cwd, options, branch)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Allocate new commit ID
	id, err := storage.NextCommitID(cwd, options)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Create commit object
	commit := storage.Commit{
		ID:        id,
		Message:   msg,
		Branch:    branch,
		Timestamp: time.Now().Unix(),
		Parent:    parentPtr,
	}

	// Write commit object to disk
	if err := storage.WriteCommitObject(cwd, options, commit); err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Update branch ref to point to new commit
	if err := storage.WriteHeadRef(cwd, options, branch, id); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("[%s %d] %s\n", branch, id, msg)
}
