package commands

import (
	"fmt"
	"os"

	"gitclone/internal/storage"
)

// Init handles the `gitclone init` command.
// It supports an optional `--bare` flag, e.g.:
//
//	gitclone init
//	gitclone init --bare
func Init(args []string) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	opts := storage.InitOptions{Bare: false}

	// Very simple flag handling: if "--bare" is present, set Bare = true.
	for _, a := range args {
		if a == "--bare" {
			opts.Bare = true
		}
	}

	if err := storage.InitRepo(cwd, opts); err != nil {
		fmt.Println("Error:", err)
		return
	}

	if opts.Bare {
		fmt.Println("Initialized bare gitclone repository.")
	} else {
		fmt.Println("Initialized empty gitclone repository in .gitclone/")
	}
}
