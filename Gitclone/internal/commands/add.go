package commands

import (
	"fmt"
	"os"

	"gitclone/internal/storage"
)

// Add stages files to the index
// Usage: gitclone add <path>
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

	// Stage the file(s)
	if err := storage.AddToIndex(cwd, options, path); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Staged: %s\n", path)
}

