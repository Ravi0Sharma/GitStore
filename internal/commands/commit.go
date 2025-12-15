package commands

import (
	"fmt"
	"os"

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

	id, err := storage.NextCommitID(cwd, options)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if err := storage.WriteHeadRef(cwd, options, branch, id); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("[%s %d] %s\n", branch, id, msg)
}
