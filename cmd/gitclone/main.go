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

// demoDisk runs a small scenario that exercises disk persistence:
// init -> commit -> checkout -> commit, then prints HEAD + branch pointers.
func demoDisk() {
	fmt.Println("== demo-disk ==")

	// Initialize (ignore error if already initialized, commands.Init prints message)
	commands.Init([]string{})

	// Commit on master
	commands.Commit([]string{"-m", "first on master"})

	// Switch to testing
	commands.Checkout([]string{"testing"})

	// Commit on testing
	commands.Commit([]string{"-m", "first on testing"})

	fmt.Println("Now inspect:")
	fmt.Println("  .gitclone/HEAD")
	fmt.Println("  .gitclone/refs/heads/master")
	fmt.Println("  .gitclone/refs/heads/testing")
}

func printHelp() {
	fmt.Println("gitclone - mini git implementation")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gitclone init [--bare]          Initialize a new repository")
	fmt.Println("  gitclone checkout <branch>      Switch branch (updates .gitclone/HEAD)")
	fmt.Println("  gitclone commit -m <msg>        Create a commit (updates refs/heads/<branch>)")
	fmt.Println("  gitclone demo                   Run in-memory demo of commits/branches")
	fmt.Println("  gitclone demo-disk              Run disk demo (init/checkout/commit)")
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

	case "commit":
		commands.Commit(args)

	case "demo":
		runDemo()

	case "demo-disk":
		demoDisk()

	default:
		fmt.Println("Unknown command:", cmd)
		printHelp()
	}
}
