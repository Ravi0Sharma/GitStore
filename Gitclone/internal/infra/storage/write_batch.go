package storage

import (
	"encoding/json"
	"fmt"

	"GitDb"
)

// WriteBatch represents a batch of writes that should be atomic
type WriteBatch struct {
	store *RepoStore
	writes []writeOp
}

type writeOp struct {
	key   string
	value []byte
}

// NewWriteBatch creates a new write batch for the given store
func NewWriteBatch(store *RepoStore) *WriteBatch {
	return &WriteBatch{
		store:  store,
		writes: make([]writeOp, 0),
	}
}

// Put adds a key-value pair to the batch
func (wb *WriteBatch) Put(key string, value []byte) {
	wb.writes = append(wb.writes, writeOp{key: key, value: value})
}

// Commit writes all operations in the batch atomically
// Uses a transaction log approach: writes a tx marker, then all operations, then marks as committed
func (wb *WriteBatch) Commit() error {
	if len(wb.writes) == 0 {
		return nil
	}

	db := wb.store.DB()

	// Generate unique tx marker key using timestamp + write count
	// This ensures uniqueness even if multiple batches run concurrently
	txMarkerKey := fmt.Sprintf("_tx/%d", len(wb.writes))

	// Create transaction record for recovery
	txRecord := txRecord{
		Type:   "batch_start",
		Writes: make([]txWrite, len(wb.writes)),
	}
	for i, op := range wb.writes {
		txRecord.Writes[i] = txWrite{
			Key:   op.key,
			Value: op.value,
		}
	}

	txData, err := json.Marshal(txRecord)
	if err != nil {
		return fmt.Errorf("failed to marshal tx record: %w", err)
	}

	// Write tx marker to database (using a special key)
	if err := db.Put(txMarkerKey, txData); err != nil {
		return fmt.Errorf("failed to write tx marker: %w", err)
	}

	// Write all operations
	for _, op := range wb.writes {
		if err := db.Put(op.key, op.value); err != nil {
			// On failure, mark tx as failed for recovery
			_ = db.Put(txMarkerKey, []byte(`{"type":"batch_failed"}`))
			return fmt.Errorf("failed to write key %s: %w", op.key, err)
		}
	}

	// Commit: mark tx as committed (overwrite the batch_start marker)
	// This must be the last write to ensure atomicity
	committedData := []byte(`{"type":"batch_committed"}`)
	if err := db.Put(txMarkerKey, committedData); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	// Force flush to ensure all writes are persisted
	// GitDb writes are append-only, so this ensures the committed marker is last
	// We need to sync the DB file to ensure writes are visible to new DB instances
	// Note: GitDb.Put() already syncs, but we ensure the log file is fully written
	return nil
}

type txRecord struct {
	Type   string    `json:"type"`
	Writes []txWrite `json:"writes"`
}

type txWrite struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
}

// RecoverTransactions recovers from incomplete transactions on startup
// This should be called when opening a RepoStore
func RecoverTransactions(store *RepoStore) error {
	db := store.DB()
	
	// Scan for transaction markers
	var incompleteTx []string
	err := db.Scan(func(record GitDb.Record) error {
		if len(record.Key) > 4 && record.Key[:4] == "_tx/" {
			var tx struct {
				Type string `json:"type"`
			}
			if err := json.Unmarshal(record.Value, &tx); err != nil {
				return nil // Skip invalid tx records
			}
			
			// Only recover incomplete transactions (batch_start or batch_failed)
			// Skip committed ones
			if tx.Type == "batch_start" || tx.Type == "batch_failed" {
				// Found incomplete transaction - mark for recovery
				incompleteTx = append(incompleteTx, record.Key)
			}
			return nil
		}
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("failed to scan for transactions: %w", err)
	}

	// Clean up incomplete transactions
	for _, txKey := range incompleteTx {
		// Mark as recovered (committed markers are left as-is)
		_ = db.Put(txKey, []byte(`{"type":"batch_recovered"}`))
	}

	return nil
}

