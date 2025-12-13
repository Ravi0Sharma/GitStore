package main

import (
	"fmt"
	"os"

	"gitclone/internal/commands"
	"gitclone/internal/core"
)

func printLog(label string, commits []*core.Commit) {
	fmt.Println("==", label, "==")
	for _, c := range commits {
		fmt.Println(c.ID, c.Message)
	}
}

func runDemo() {
	repo := core.NewRepo("test")

	repo.Commit("Initial commit")
	repo.Commit("Change 1")

	printLog("master log", repo.Log()) // 1, 0

	repo.Checkout("testing")
	repo.Commit("Change 2 on testing")

	printLog("testing log", repo.Log()) // 2,1,0

	repo.Checkout("master")
	printLog("master log again", repo.Log()) // 1,0
}

func printHelp() {
	fmt.Println("gitclone - mini git implementation")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gitclone init [--bare]        Initialize a new repository")
	fmt.Println("  gitclone checkout <branch>    Switch branch (updates .gitclone/HEAD)")
	fmt.Println("  gitclone demo                 Run in-memory demo of commits/branches")
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

	case "checkout":
		commands.Checkout(args)

	case "demo":
		runDemo()

	default:
		fmt.Println("Unknown command:", cmd)
		printHelp()
	}
}
