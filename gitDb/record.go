package GitDb

import (
	"encoding/binary"
	"fmt"
)

type Record struct {
	Key   string
	Value []byte
}

// Encode converts a Record into a byte slice.
func (record Record) Encode() ([]byte, error) {
	if record.Key == "" {
		return nil, fmt.Errorf("empty key")
	}
	keyBytes := []byte(r.Key)
	keyLen := uint32(len(keyBytes))
	valLen := uint32(len(record.Value))

	// 8 bytes header + payload
	buf := make([]byte, 8+len(keyBytes)+len(record.Value))

	// key length header
	binary.LittleEndian.PutUint32(buf[0:4], keyLen)
	// value length header
	binary.LittleEndian.PutUint32(buf[4:8], valLen)

	// Copy key bytes after the 8-byte header
	copy(buf[8:8+len(keyBytes)], keyBytes)

	// Copy value bytes immediately after the key
	copy(buf[8+len(keyBytes):], record.Value)

	return buf, nil
}

// DecodeRecord decodes a record and its size from bytes.
func DecodeRecord(log []byte, offset int64) (rec Record, size int64, err error) {
	if offset < 0 || offset >= int64(len(log)) {
		return Record{}, 0, fmt.Errorf("offset out of range")
	}

	if int64(len(log))-offset < 8 {
		return Record{}, 0, fmt.Errorf("not enough bytes for header")
	}

	// Reads key & value length from header
	keyLen := int64(binary.LittleEndian.Uint32(log[offset : offset+4]))
	valLen := int64(binary.LittleEndian.Uint32(log[offset+4 : offset+8]))

	total := 8 + keyLen + valLen
	if total < 8 {
		return Record{}, 0, fmt.Errorf("invalid lengths")
	}

	if int64(len(log))-offset < total {
		return Record{}, 0, fmt.Errorf("not enough bytes for record")
	}

	keyStart := offset + 8
	keyEnd := keyStart + keyLen
	valStart := keyEnd
	valEnd := valStart + valLen

	key := string(log[keyStart:keyEnd])
	val := make([]byte, valLen)
	copy(val, log[valStart:valEnd])

	return Record{Key: key, Value: val}, total, nil

}
