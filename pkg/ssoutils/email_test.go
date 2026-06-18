package ssoutils

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestEmailDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "standard email",
			email:    "user@theopenlane.io",
			expected: "theopenlane.io",
		},
		{
			name:     "subdomain",
			email:    "user@mail.example.com",
			expected: "mail.example.com",
		},
		{
			name:     "no at sign",
			email:    "notanemail",
			expected: "",
		},
		{
			name:     "empty string",
			email:    "",
			expected: "",
		},
		{
			name:     "multiple at signs uses last",
			email:    "a@b@theopenlane.io",
			expected: "theopenlane.io",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, EmailDomain(tc.email))
		})
	}
}
