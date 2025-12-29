package storage

// Legacy functions kept for backward compatibility
// New code should use functions from index_entries.go

// GetStagedFiles returns the list of staged file paths (for backward compatibility)
func GetStagedFiles(root string, options InitOptions) ([]string, error) {
	entries, err := GetIndexEntries(root, options)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(entries))
	for path, entry := range entries {
		if entry.BlobID != "" {
			paths = append(paths, path)
		}
	}

	return paths, nil
}
