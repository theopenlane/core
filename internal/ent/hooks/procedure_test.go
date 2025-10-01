package hooks

import (
	"testing"
)

func TestDetectMimeTypeFromContent(t *testing.T) {
	tests := []struct {
		name             string
		content          []byte
		fallbackMimeType string
		expected         string
	}{
		{
			name:             "empty content uses fallback",
			content:          []byte{},
			fallbackMimeType: "text/plain",
			expected:         "text/plain; charset=utf-8",
		},
		{
			name:             "text content detected",
			content:          []byte("Hello, World!"),
			fallbackMimeType: "application/octet-stream",
			expected:         "text/plain; charset=utf-8",
		},
		{
			name:             "fallback normalized",
			content:          []byte{},
			fallbackMimeType: "  TEXT/PLAIN; charset=utf-8  ",
			expected:         "text/plain; charset=utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectMimeTypeFromContent(tt.content, tt.fallbackMimeType)
			if result != tt.expected {
				t.Errorf("detectMimeTypeFromContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}
