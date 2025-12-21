package commands

import (
	"fmt"
	"os"

	"gitclone/internal/storage"
)

// It supports an optional `--bare`

// gitclone init
// gitclone init --bare
func Init(args []string) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	options := storage.InitOptions{Bare: false}

	// if "--bare" is present, set Bare = true.
	for _, a := range args {
		if a == "--bare" {
			options.Bare = true
		}
	}

	if err := storage.InitRepo(cwd, options); err != nil {
		fmt.Println("Error:", err)
		return
	}

	if options.Bare {
		fmt.Println("Initialized bare gitclone repository.")
	} else {
		fmt.Println("Initialized empty gitclone repository in .gitclone/")
	}
}
