package GitDb

import "fmt"

type DB struct {
	log   []byte
	index *Index
}

// Open initializes a new database instance
func Open(_ string) (*DB, error) {
	return &DB{
		log:   make([]byte, 0, 4096),
		index: newIndex(),
	}, nil
}
