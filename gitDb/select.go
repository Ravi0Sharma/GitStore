package GitDb

import "fmt"

// SelectAll prints all records in the database using Scan.
func SelectAll(db *DB) error {
	return db.Scan(func(record Record) error {
		fmt.Printf("key: %s, value: %s\n", record.Key, string(record.Value))
		return nil
	})
}

