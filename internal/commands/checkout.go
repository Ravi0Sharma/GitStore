package commands

import (
	"fmt"
	"gitclone/internal/storage"
	"os"
)

func Checkout(args []string) {
	if len(args) < 1 {
		fmt.Println("usage: gitclone checkout <branch>")
		return
	}
	branch := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// For now: assume non-bare. (Can auto-detect later.)
	opts := storage.InitOptions{Bare: false}

	// Ensure refs/heads/<branch> exists (empty if new branch)
	if err := storage.EnsureBranchRefExists(cwd, opts, branch); err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Update HEAD to point to that branch
	if err := storage.WriteHEADBranch(cwd, opts, branch); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Switched to branch", branch)
}
