package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gitclone/internal/commands"
)

// isGitCloneRepo checks if current directory is a GitClone repo
// Returns true if .gitclone/ exists and .git/ does not exist
func isGitCloneRepo() bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	gitclonePath := filepath.Join(cwd, ".gitclone")
	gitPath := filepath.Join(cwd, ".git")

	_, errGitClone := os.Stat(gitclonePath)
	_, errGit := os.Stat(gitPath)

	// GitClone repo: .gitclone exists and .git does not
	return errGitClone == nil && os.IsNotExist(errGit)
}

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

	// Get program name to detect if called as "git" or "gitclone"
	programName := filepath.Base(os.Args[0])
	cmd := os.Args[1]
	args := os.Args[2:]

	// If called as "git" and we're in a GitClone repo, handle git commands
	if programName == "git" || cmd == "git" {
		if cmd == "git" && len(args) > 0 {
			// Handle: git add, git commit, etc.
			cmd = args[0]
			args = args[1:]
		}

		// Only alias git commands if we're in a GitClone repo
		if isGitCloneRepo() {
			// Alias git commands to gitclone commands
			switch cmd {
			case "add":
				commands.Add(args)
				return
			case "commit":
				commands.Commit(args)
				return
			case "checkout":
				commands.Checkout(args)
				return
			case "merge":
				commands.Merge(args)
				return
			case "log":
				commands.Log(args)
				return
			case "show":
				commands.Show(args)
				return
			case "init":
				commands.Init(args)
				return
			default:
				fmt.Printf("Error: 'git %s' is not supported in GitClone repositories.\n", cmd)
				fmt.Println("Use 'gitclone' commands instead, or switch to a standard Git repository.")
				os.Exit(1)
				return
			}
		} else {
			// Not a GitClone repo, tell user to use system git
			fmt.Println("Error: Not a GitClone repository.")
			fmt.Println("This is a GitClone CLI. For standard Git repositories, use the system 'git' command.")
			os.Exit(1)
			return
		}
	}

	// Normal gitclone command handling
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
