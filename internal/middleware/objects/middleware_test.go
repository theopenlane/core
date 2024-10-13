package objects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{"Bytes", 512, "512 bytes"},
		{"Kilobytes", 2048, "2.00 KB"},
		{"Megabytes", 10485760, "10.00 MB"},
		{"Gigabytes", 10737418240, "10.00 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFileSize(tt.size)
			assert.Equal(t, tt.expected, result)
		})
	}
}
