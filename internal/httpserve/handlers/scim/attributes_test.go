package scim

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractMemberIDsFromValue(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  []string
	}{
		{
			name: "extract unique member IDs",
			value: []any{
				map[string]any{"value": "user1"},
				map[string]any{"value": "user2"},
				map[string]any{"value": "user3"},
			},
			want: []string{"user1", "user2", "user3"},
		},
		{
			name: "deduplicate member IDs",
			value: []any{
				map[string]any{"value": "user1"},
				map[string]any{"value": "user2"},
				map[string]any{"value": "user1"},
				map[string]any{"value": "user3"},
				map[string]any{"value": "user2"},
			},
			want: []string{"user1", "user2", "user3"},
		},
		{
			name: "skip non-map items",
			value: []any{
				"not-a-map",
				map[string]any{"value": "user1"},
				123,
				map[string]any{"value": "user2"},
			},
			want: []string{"user1", "user2"},
		},
		{
			name: "skip empty values",
			value: []any{
				map[string]any{"value": ""},
				map[string]any{"value": "user1"},
				map[string]any{"other": "field"},
			},
			want: []string{"user1"},
		},
		{
			name: "skip non-string values",
			value: []any{
				map[string]any{"value": 123},
				map[string]any{"value": "user1"},
			},
			want: []string{"user1"},
		},
		{
			name:  "non-array value returns nil",
			value: "not-an-array",
			want:  nil,
		},
		{
			name:  "nil value returns nil",
			value: nil,
			want:  nil,
		},
		{
			name:  "empty array returns empty slice",
			value: []any{},
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractMemberIDsFromValue(tt.value)
			assert.Equal(t, tt.want, got)
		})
	}
}
