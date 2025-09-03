package objects

import (
	"testing"
)

func TestParseDocument(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		mimeType string
		expected string
		hasError bool
	}{
		{
			name:     "plain text",
			content:  []byte("Hello World\nThis is a test."),
			mimeType: "text/plain",
			expected: "Hello World\nThis is a test.",
			hasError: false,
		},
		{
			name:     "markdown",
			content:  []byte("# Hello\n\nThis is **bold** text."),
			mimeType: "text/markdown",
			expected: "# Hello\n\nThis is **bold** text.",
			hasError: false,
		},
		{
			name:     "empty content",
			content:  []byte(""),
			mimeType: "text/plain",
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDocument(tt.content, tt.mimeType)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
