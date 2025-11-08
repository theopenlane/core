package graphapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectGraphQLOperationType(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{
			name:     "mutation operation",
			query:    "mutation CreateUser { createUser(input: {name: \"John\"}) { id } }",
			expected: false,
		},
		{
			name:     "mutation with leading whitespace",
			query:    "   mutation CreateUser { createUser(input: {name: \"John\"}) { id } }",
			expected: false,
		},
		{
			name:     "mutation uppercase",
			query:    "MUTATION CreateUser { createUser(input: {name: \"John\"}) { id } }",
			expected: false,
		},
		{
			name:     "subscription operation",
			query:    "subscription MessageAdded { messageAdded { id content } }",
			expected: true,
		},
		{
			name:     "subscription with leading whitespace",
			query:    "\t\nsubscription MessageAdded { messageAdded { id content } }",
			expected: true,
		},
		{
			name:     "subscription uppercase",
			query:    "SUBSCRIPTION MessageAdded { messageAdded { id content } }",
			expected: true,
		},
		{
			name:     "query operation",
			query:    "query GetUser { user(id: \"123\") { name email } }",
			expected: true,
		},
		{
			name:     "query with leading whitespace",
			query:    "  query GetUser { user(id: \"123\") { name email } }",
			expected: true,
		},
		{
			name:     "query uppercase",
			query:    "QUERY GetUser { user(id: \"123\") { name email } }",
			expected: true,
		},
		{
			name:     "anonymous query",
			query:    "{ user(id: \"123\") { name email } }",
			expected: true,
		},
		{
			name:     "anonymous query with whitespace",
			query:    "  { user(id: \"123\") { name email } }  ",
			expected: true,
		},
		{
			name:     "empty string",
			query:    "",
			expected: false,
		},
		{
			name:     "whitespace only",
			query:    "   \t\n  ",
			expected: false,
		},
		{
			name:     "invalid operation type",
			query:    "invalid { user { id } }",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectGraphQLOperationType(tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}
