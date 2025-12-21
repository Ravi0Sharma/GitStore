package GitDb

// Index maps keys to their latest log offsets
type Index struct {
	latest map[string]int64
}

// newIndex creates an empty index
func newIndex() *Index {
	return &Index{latest: make(map[string]int64)}
}

// Set updates the offset for a key
func (index *Index) Set(key string, offset int64) {
	index.latest[key] = offset
}

// Get returns the offset for a key
func (index *Index) Get(key string) (int64, bool) {
	off, ok := index.latest[key]
	return off, ok
}
