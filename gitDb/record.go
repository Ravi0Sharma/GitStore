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
