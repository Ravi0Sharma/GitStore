package commands

import (
	"fmt"
	"os"

	"gitclone/internal/storage"
)

// Add stages files to the index
// Usage: gitclone add <path> or git add <path> (in GitClone repos)
func Add(args []string) {
	path := "."
	if len(args) >= 1 {
		path = args[0]
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	options := storage.InitOptions{Bare: false}

	// Get count of staged entries before adding
	entriesBefore, err := storage.GetIndexEntries(cwd, options)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	countBefore := len(entriesBefore)

	// Stage the file(s)
	if err := storage.AddToIndex(cwd, options, path); err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Verify staging succeeded by checking entries after
	entriesAfter, err := storage.GetIndexEntries(cwd, options)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	countAfter := len(entriesAfter)

	// Check if anything was actually staged
	if countAfter == countBefore {
		// Nothing was staged - this is an error
		if path == "." {
			fmt.Println("Error: No changes to stage. No files found or all files are already staged.")
		} else {
			fmt.Printf("Error: Failed to stage '%s'. File may not exist or is already staged.\n", path)
		}
		os.Exit(1)
		return
	}

	// Success - show accurate staging information
	stagedCount := countAfter - countBefore
	if stagedCount == 1 {
		// Find the newly staged entry
		for p := range entriesAfter {
			if _, exists := entriesBefore[p]; !exists {
				fmt.Printf("Staged: %s\n", p)
				return
			}
		}
		fmt.Printf("Staged: %s\n", path)
	} else {
		// Multiple files staged
		fmt.Printf("Staged %d file(s)\n", stagedCount)
		// Optionally list a few staged paths
		listed := 0
		for p := range entriesAfter {
			if _, exists := entriesBefore[p]; !exists {
				if listed < 3 {
					fmt.Printf("  %s\n", p)
					listed++
				}
			}
		}
		if stagedCount > 3 {
			fmt.Printf("  ... and %d more\n", stagedCount-3)
		}
	}
}

