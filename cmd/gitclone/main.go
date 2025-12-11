package main

import (
	"fmt"

	"gitclone/internal/core"
)

func printLog(label string, commits []*core.Commit) {
	fmt.Println("==", label, "==")
	for _, c := range commits {
		fmt.Println(c.ID, c.Message)
	}
}

func main() {
	repo := core.NewRepo("test")

	repo.Commit("Initial commit")
	repo.Commit("Change 1")

	printLog("master log", repo.Log()) // 1, 0

	repo.Checkout("testing")
	repo.Commit("Change 2 on testing")

	printLog("testing log", repo.Log()) // 2,1,0

	repo.Checkout("master")
	printLog("master log again", repo.Log()) // fortfarande 1,0
}
