package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Commit struct {
	ID        int    `json:"id"`
	Message   string `json:"message"`
	Branch    string `json:"branch"`
	Timestamp int64  `json:"timestamp"`
	Parent    *int   `json:"parent,omitempty"`
}
