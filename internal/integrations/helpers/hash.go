package helpers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"hash"
	"sort"
	"time"
)

// HashBuilder provides a composable interface for building content hashes
type HashBuilder struct {
	hasher hash.Hash
}

// NewHashBuilder creates a new hash builder using SHA-256
func NewHashBuilder() *HashBuilder {
	return &HashBuilder{
		hasher: sha256.New(),
	}
}

// WriteStrings adds one or more strings to the hash, skipping empty values
func (h *HashBuilder) WriteStrings(values ...string) *HashBuilder {
	for _, value := range values {
		if value == "" {
			continue
		}
		_, _ = h.hasher.Write([]byte(value))
	}

	return h
}

// WriteTime adds a time value to the hash in RFC3339Nano format
func (h *HashBuilder) WriteTime(t time.Time) *HashBuilder {
	if t.IsZero() {
		return h
	}

	_, _ = h.hasher.Write([]byte(t.UTC().Format(time.RFC3339Nano)))

	return h
}

// WriteTimePtr adds a time pointer to the hash, handling nil values
func (h *HashBuilder) WriteTimePtr(t *time.Time) *HashBuilder {
	if t == nil || t.IsZero() {
		return h
	}

	return h.WriteTime(*t)
}

// WriteSortedMap adds a map[string]any to the hash in deterministic order
func (h *HashBuilder) WriteSortedMap(m map[string]any) *HashBuilder {
	if len(m) == 0 {
		return h
	}

	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		h.WriteStrings(key)
		value := m[key]

		switch v := value.(type) {
		case string:
			h.WriteStrings(v)
		default:
			if encoded, err := json.Marshal(v); err == nil {
				_, _ = h.hasher.Write(encoded)
			}
		}
	}

	return h
}

// Hex returns the hex-encoded hash digest
func (h *HashBuilder) Hex() string {
	return hex.EncodeToString(h.hasher.Sum(nil))
}
