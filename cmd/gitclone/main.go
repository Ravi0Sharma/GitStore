package main

import (
	"fmt"
	"os"
	"strconv"

	"gitclone/internal/commands"
	"gitclone/internal/core"
	"gitclone/internal/storage"
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

// diskLog prints commit history from disk following:
// HEAD -> refs/heads/<branch> -> objects/<id>.json -> parent -> ...
func diskLog() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	opts := storage.InitOptions{Bare: false}

	branch, err := storage.ReadHEADBranch(cwd, opts)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Read current branch tip (latest commit id)
	tipPtr, err := storage.ReadHeadRefMaybe(cwd, opts, branch)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if tipPtr == nil {
		fmt.Printf("On branch %s (no commits)\n", branch)
		return
	}

	fmt.Printf("== log (%s) ==\n", branch)

	id := *tipPtr
	for {
		c, err := storage.ReadCommitObject(cwd, opts, id)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Printf("commit %d\n", c.ID)
		if c.Parent != nil {
			fmt.Printf("parent %d\n", *c.Parent)
		}
		fmt.Printf("branch %s\n", c.Branch)
		fmt.Printf("message %s\n", c.Message)
		fmt.Println()

		// Stop at root commit (no parent)
		if c.Parent == nil {
			break
		}
		// Move to parent commit for next iteration
		id = *c.Parent
	}
}

// diskShow prints a single commit object by id.
func diskShow(args []string) {
	if len(args) < 1 {
		fmt.Println("usage: gitclone show <id>")
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Error: id must be a number")
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	opts := storage.InitOptions{Bare: false}

	c, err := storage.ReadCommitObject(cwd, opts, id)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("commit %d\n", c.ID)
	if c.Parent != nil {
		fmt.Printf("parent %d\n", *c.Parent)
	}
	fmt.Printf("branch %s\n", c.Branch)
	fmt.Printf("message %s\n", c.Message)
}

func printHelp() {
	fmt.Println("gitclone - mini git implementation")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gitclone init [--bare]          Initialize a new repository")
	fmt.Println("  gitclone checkout <branch>      Switch branch (updates .gitclone/HEAD)")
	fmt.Println("  gitclone commit -m <msg>        Create a commit (writes objects/<id>.json + updates refs/heads/<branch>)")
	fmt.Println("  gitclone log                    Show commit history from disk (follows parent chain)")
	fmt.Println("  gitclone show <id>              Show a single commit object from disk")
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

	case "log":
		diskLog()

	case "show":
		diskShow(args)

	case "demo":
		runDemo()

	case "demo-disk":
		demoDisk()

	default:
		fmt.Println("Unknown command:", cmd)
		printHelp()
	}
}
