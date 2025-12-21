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

// Close shuts down the database (nothing to close in memory).
func (db *DB) Close() error {
	return nil
}

// Append record to the log and update the index
func (db *DB) Put(key string, value []byte) error {
	record := Record{Key: key, Value: value}
	encoded, err := record.Encode()
	if err != nil {
		return err
	}

	offset := int64(len(db.log))
	db.log = append(db.log, encoded...)
	db.index.Set(key, offset)
	return nil
}

// Get retrieves a value by key from the database
func (db *DB) Get(key string) ([]byte, error) {
	offset, ok := db.index.Get(key)
	if !ok {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	record, _, err := DecodeRecord(db.log, offset)
	if err != nil {
		return nil, err
	}
	return record.Value, nil
}
