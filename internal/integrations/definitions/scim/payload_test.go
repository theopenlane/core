package scim

import (
	"testing"

	"github.com/elimity-com/scim"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			got := ExtractMemberIDsFromValue(tt.value)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDirectoryAccountExternalID(t *testing.T) {
	tests := []struct {
		name       string
		attributes scim.ResourceAttributes
		want       string
	}{
		{
			name: "uses externalId when present",
			attributes: scim.ResourceAttributes{
				"externalId": "ext-123",
				"userName":   "user@example.com",
			},
			want: "ext-123",
		},
		{
			name: "falls back to userName when externalId absent",
			attributes: scim.ResourceAttributes{
				"userName": "user@example.com",
			},
			want: "user@example.com",
		},
		{
			name: "falls back to email when externalId and userName absent",
			attributes: scim.ResourceAttributes{
				"emails": []any{
					map[string]any{"value": "first@example.com"},
				},
			},
			want: "first@example.com",
		},
		{
			name: "trims whitespace from externalId",
			attributes: scim.ResourceAttributes{
				"externalId": "  ext-456  ",
			},
			want: "ext-456",
		},
		{
			name: "skips blank externalId",
			attributes: scim.ResourceAttributes{
				"externalId": "   ",
				"userName":   "fallback@example.com",
			},
			want: "fallback@example.com",
		},
		{
			name:       "returns empty when no fields present",
			attributes: scim.ResourceAttributes{},
			want:       "",
		},
		{
			name: "skips non-map email items",
			attributes: scim.ResourceAttributes{
				"emails": []any{
					"not-a-map",
					map[string]any{"value": "valid@example.com"},
				},
			},
			want: "valid@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DirectoryAccountExternalID(tt.attributes)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDirectoryGroupExternalID(t *testing.T) {
	tests := []struct {
		name       string
		attributes scim.ResourceAttributes
		want       string
	}{
		{
			name: "uses externalId when present",
			attributes: scim.ResourceAttributes{
				"externalId":  "grp-123",
				"displayName": "Engineering",
			},
			want: "grp-123",
		},
		{
			name: "falls back to displayName when externalId absent",
			attributes: scim.ResourceAttributes{
				"displayName": "Engineering",
			},
			want: "Engineering",
		},
		{
			name: "trims whitespace",
			attributes: scim.ResourceAttributes{
				"externalId": "  grp-456  ",
			},
			want: "grp-456",
		},
		{
			name: "skips blank externalId",
			attributes: scim.ResourceAttributes{
				"externalId":  "   ",
				"displayName": "Fallback Group",
			},
			want: "Fallback Group",
		},
		{
			name:       "returns empty when no fields present",
			attributes: scim.ResourceAttributes{},
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DirectoryGroupExternalID(tt.attributes)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildDirectoryAccountPayloadSet(t *testing.T) {
	tests := []struct {
		name       string
		attributes scim.ResourceAttributes
		action     string
		wantErr    error
		wantID     string
	}{
		{
			name: "valid payload with externalId",
			attributes: scim.ResourceAttributes{
				"externalId": "ext-001",
				"userName":   "user@example.com",
			},
			action:  "create",
			wantErr: nil,
			wantID:  "ext-001",
		},
		{
			name: "valid payload falling back to userName",
			attributes: scim.ResourceAttributes{
				"userName": "user@example.com",
			},
			action:  "update",
			wantErr: nil,
			wantID:  "user@example.com",
		},
		{
			name:       "empty attributes returns error",
			attributes: scim.ResourceAttributes{},
			action:     "create",
			wantErr:    ErrInvalidAttributes,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildDirectoryAccountPayloadSet(tt.attributes, tt.action)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, got.Envelopes)
			assert.Equal(t, tt.wantID, got.Envelopes[0].Resource)
			assert.Equal(t, tt.action, got.Envelopes[0].Action)
		})
	}
}
