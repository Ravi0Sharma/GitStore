package commands

import (
	"fmt"
	"os"
	"strconv"

	"gitclone/internal/storage"
)

func Log(args []string) {
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
		if c.Parent2 != nil {
			fmt.Printf("parent2 %d\n", *c.Parent2)
		}
		fmt.Printf("branch %s\n", c.Branch)
		fmt.Printf("message %s\n\n", c.Message)

		if c.Parent == nil {
			break
		}
		id = *c.Parent
	}
}

func Show(args []string) {
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
	if c.Parent2 != nil {
		fmt.Printf("parent2 %d\n", *c.Parent2)
	}
	fmt.Printf("branch %s\n", c.Branch)
	fmt.Printf("message %s\n", c.Message)
}
