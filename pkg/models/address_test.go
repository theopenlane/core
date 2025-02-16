package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddressString(t *testing.T) {
	tests := []struct {
		name     string
		address  Address
		expected string
	}{
		{
			name:     "Empty address",
			address:  Address{},
			expected: "",
		},
		{
			name: "Address with all fields",
			address: Address{
				Line1:      "123 Main St",
				Line2:      "Apt 4B",
				City:       "Springfield",
				State:      "IL",
				PostalCode: "62701",
				Country:    "USA",
			},
			expected: "123 Main St Apt 4B Springfield, IL 62701 USA",
		},
		{
			name: "Address without the first line",
			address: Address{
				State:      "IL",
				PostalCode: "62701",
				Country:    "USA",
			},
			expected: "IL 62701 USA",
		},
		{
			name: "Address with only State and Postal",
			address: Address{
				State:      "IL",
				PostalCode: "62701",
			},
			expected: "IL 62701",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.address.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}
