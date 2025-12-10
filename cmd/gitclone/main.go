package main

import (
	"fmt"

	"gitclone/internal/core"
)

func main() {
	repo := core.NewRepo("test")

	c1 := repo.Commit("Initial commit")
	c2 := repo.Commit("Another commit")

	fmt.Println(c1.ID, c1.Message) // Should print: 0 Initial commit
	fmt.Println(c2.ID, c2.Message) // Should print: 1 Another commit
}
