package main

import (
	"fmt"
	"os"

	"gitclone/internal/commands"
)

func printHelp() {
	fmt.Println("gitclone - mini git implementation")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gitclone init [--bare]          Initialize a new repository")
	fmt.Println("  gitclone add <path>             Stage files for commit")
	fmt.Println("  gitclone checkout <branch>      Switch branch (updates .gitclone/HEAD)")
	fmt.Println("  gitclone commit -m <msg>        Create a commit")
	fmt.Println("  gitclone merge <branch>         Merge branch into current branch")
	fmt.Println("  gitclone log                    Show commit history")
	fmt.Println("  gitclone show <id>              Show a single commit")
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "init":
		commands.Init(args)

	case "add":
		commands.Add(args)

	case "checkout":
		commands.Checkout(args)

	case "commit":
		commands.Commit(args)

	case "merge":
		commands.Merge(args)

	case "log":
		commands.Log(args)

	case "show":
		commands.Show(args)

	default:
		fmt.Println("Unknown command:", cmd)
		printHelp()
	}
}
