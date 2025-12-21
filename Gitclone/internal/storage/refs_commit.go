package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func NextCommitID(root string, options InitOptions) (int, error) {
	p := filepath.Join(repoRoot(root, options), "NEXT_COMMIT_ID")

	b, err := os.ReadFile(p)
	if err != nil {
		return 0, err
	}

	curStr := strings.TrimSpace(string(b))
	cur, err := strconv.Atoi(curStr)
	if err != nil {
		return 0, fmt.Errorf("invalid NEXT_COMMIT_ID: %q", curStr)
	}

	if err := os.WriteFile(p, []byte(fmt.Sprintf("%d\n", cur+1)), 0o644); err != nil {
		return 0, err
	}

	return cur, nil
}
