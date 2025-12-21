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
	targetBranch := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// For now: assume non-bare. (Can auto-detect later.)
	opts := storage.InitOptions{Bare: false}

	// Read current branch from HEAD
	currentBranch, err := storage.ReadHEADBranch(cwd, opts)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// If checking out the same branch, print message and return
	if targetBranch == currentBranch {
		fmt.Printf("Already on branch %s\n", targetBranch)
		return
	}

	// Ensure target branch ref file exists
	if err := storage.EnsureHeadRefExists(cwd, opts, targetBranch); err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Check if target branch is new (empty/missing)
	targetTip, err := storage.ReadHeadRefMaybe(cwd, opts, targetBranch)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// If target branch is new (nil tip), copy current branch's tip commit ID
	if targetTip == nil {
		currentTip, err := storage.ReadHeadRefMaybe(cwd, opts, currentBranch)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		// If current branch has commits, copy the tip to the new branch
		if currentTip != nil {
			if err := storage.WriteHeadRef(cwd, opts, targetBranch, *currentTip); err != nil {
				fmt.Println("Error:", err)
				return
			}
		}
		// If current branch has no commits, new branch ref remains empty (already created by EnsureHeadRefExists)
	}

	// Update HEAD to point to target branch
	if err := storage.WriteHEADBranch(cwd, opts, targetBranch); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Switched to branch %s\n", targetBranch)
}
