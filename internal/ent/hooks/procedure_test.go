package hooks

import (
	"bytes"
	"testing"

	storage "github.com/theopenlane/core/pkg/objects/storage"
)

// TestDetectContentType is a table test verifying MIME detection results for representative payloads
func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected string
	}{
		{
			name:     "empty content returns octet-stream",
			content:  []byte{},
			expected: "application/octet-stream",
		},
		{
			name:     "text content detected",
			content:  []byte("Hello, World!"),
			expected: "text/plain; charset=utf-8",
		},
		{
			name:     "json content detected",
			content:  []byte("{\"a\":1}"),
			expected: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.content)
			result, err := storage.DetectContentType(reader)
			if err != nil {
				t.Fatalf("DetectContentType() error = %v", err)
			}
			if result != tt.expected {
				t.Errorf("DetectContentType() = %v, want %v", result, tt.expected)
			}
		})
	}
}
