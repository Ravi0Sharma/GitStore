package GitDb

import (
	"fmt"
	"os"
	"path/filepath"
)

type DB struct {
	log     []byte
	index   *Index
	logPath string
}

// Open initializes a new database instance
func Open(path string) (*DB, error) {
	logPath := filepath.Join(path, "log")
	db := &DB{
		log:     make([]byte, 0, 4096),
		index:   newIndex(),
		logPath: logPath,
	}

	// Load existing log file if it exists
	if data, err := os.ReadFile(logPath); err == nil {
		db.log = data
		// Rebuild index from log
		if err := db.rebuildIndex(); err != nil {
			return nil, fmt.Errorf("failed to rebuild index: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read log file: %w", err)
	}

	return db, nil
}

// rebuildIndex reconstructs the index by reading all records from the log
func (db *DB) rebuildIndex() error {
	offset := int64(0)
	for offset < int64(len(db.log)) {
		record, size, err := DecodeRecord(db.log, offset)
		if err != nil {
			return err
		}
		// Update index with latest offset for this key
		db.index.Set(record.Key, offset)
		offset += size
	}
	return nil
}

// Close shuts down the database
// Since Put() already appends to the log file, Close() ensures the in-memory log
// matches the file by writing it (which should be identical if no errors occurred).
// This also ensures any buffered writes are flushed.
func (db *DB) Close() error {
	// IMPORTANT: Close must never rewrite/truncate the on-disk log.
	// Rewriting from an in-memory snapshot is unsafe if multiple DB handles exist:
	// a stale handle could drop records appended by a newer handle.
	// Since Put() already appends to the log file and syncs, Close() only needs to flush.
	return db.Flush()
}

// Flush ensures any previously written log data is persisted to disk.
// It MUST NOT rewrite/truncate the log file.
func (db *DB) Flush() error {
	if err := os.MkdirAll(filepath.Dir(db.logPath), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}
	// Sync the file if it exists (or create it empty if this DB was never written to).
	file, err := os.OpenFile(db.logPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file for flush: %w", err)
	}
	defer file.Close()
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync log file: %w", err)
	}
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

	// Append to log file for persistence
	if err := os.MkdirAll(filepath.Dir(db.logPath), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}
	file, err := os.OpenFile(db.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	if _, err := file.Write(encoded); err != nil {
		file.Close()
		return fmt.Errorf("failed to write to log file: %w", err)
	}
	// Sync to ensure write is persisted to disk immediately
	// This is critical for ensuring writes are visible when a new DB instance is opened
	if err := file.Sync(); err != nil {
		file.Close()
		return fmt.Errorf("failed to sync log file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close log file: %w", err)
	}
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

// Scan iterates through all records in the log, calling fn for each record.
func (db *DB) Scan(fn func(Record) error) error {
	offset := int64(0)
	for offset < int64(len(db.log)) {
		record, bytesConsumed, err := DecodeRecord(db.log, offset)
		if err != nil {
			return err
		}
		if err := fn(record); err != nil {
			return err
		}
		offset += bytesConsumed
	}
	return nil
}
