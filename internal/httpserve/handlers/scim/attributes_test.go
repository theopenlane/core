package scim

import (
	"testing"

	"github.com/elimity-com/scim"
	"github.com/stretchr/testify/assert"
)

func TestExtractEmail(t *testing.T) {
	tests := []struct {
		name       string
		attributes scim.ResourceAttributes
		want       string
		wantErr    bool
	}{
		{
			name: "primary email from emails array",
			attributes: scim.ResourceAttributes{
				"emails": []any{
					map[string]any{
						"value":   "secondary@example.com",
						"primary": false,
					},
					map[string]any{
						"value":   "primary@example.com",
						"primary": true,
					},
				},
			},
			want:    "primary@example.com",
			wantErr: false,
		},
		{
			name: "first email when no primary",
			attributes: scim.ResourceAttributes{
				"emails": []any{
					map[string]any{
						"value": "first@example.com",
					},
					map[string]any{
						"value": "second@example.com",
					},
				},
			},
			want:    "first@example.com",
			wantErr: false,
		},
		{
			name: "userName fallback when no emails",
			attributes: scim.ResourceAttributes{
				"userName": "user@example.com",
			},
			want:    "user@example.com",
			wantErr: false,
		},
		{
			name: "skip invalid emails",
			attributes: scim.ResourceAttributes{
				"emails": []any{
					map[string]any{
						"value": "not-an-email",
					},
					map[string]any{
						"value": "valid@example.com",
					},
				},
			},
			want:    "valid@example.com",
			wantErr: false,
		},
		{
			name: "skip empty email values",
			attributes: scim.ResourceAttributes{
				"emails": []any{
					map[string]any{
						"value": "",
					},
					map[string]any{
						"value": "valid@example.com",
					},
				},
			},
			want:    "valid@example.com",
			wantErr: false,
		},
		{
			name: "error when no valid email found",
			attributes: scim.ResourceAttributes{
				"userName": "not-an-email",
			},
			want:    "",
			wantErr: true,
		},
		{
			name:       "error when no email fields present",
			attributes: scim.ResourceAttributes{},
			want:       "",
			wantErr:    true,
		},
		{
			name: "skip non-map items in emails array",
			attributes: scim.ResourceAttributes{
				"emails": []any{
					"not-a-map",
					map[string]any{
						"value": "valid@example.com",
					},
				},
			},
			want:    "valid@example.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractEmail(tt.attributes)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "no valid email found")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestExtractUserAttributes(t *testing.T) {
	tests := []struct {
		name       string
		attributes scim.ResourceAttributes
		want       *UserAttributes
		wantErr    bool
	}{
		{
			name: "complete user attributes",
			attributes: scim.ResourceAttributes{
				"userName":   "user@example.com",
				"externalId": "ext123",
				"name": map[string]any{
					"givenName":  "John",
					"familyName": "Doe",
				},
				"displayName":       "John Doe",
				"preferredLanguage": "en",
				"locale":            "en_US",
				"profileUrl":        "https://example.com/profile",
				"active":            true,
			},
			want: &UserAttributes{
				UserName:          "user@example.com",
				Email:             "user@example.com",
				ExternalID:        "ext123",
				FirstName:         "John",
				LastName:          "Doe",
				DisplayName:       "John Doe",
				PreferredLanguage: "en",
				Locale:            "en_US",
				ProfileURL:        "https://example.com/profile",
				Active:            true,
			},
			wantErr: false,
		},
		{
			name: "displayName inferred from names",
			attributes: scim.ResourceAttributes{
				"userName": "user@example.com",
				"name": map[string]any{
					"givenName":  "Jane",
					"familyName": "Smith",
				},
			},
			want: &UserAttributes{
				UserName:    "user@example.com",
				Email:       "user@example.com",
				FirstName:   "Jane",
				LastName:    "Smith",
				DisplayName: "Jane Smith",
				Active:      true,
			},
			wantErr: false,
		},
		{
			name: "displayName inferred from names with trimming",
			attributes: scim.ResourceAttributes{
				"userName": "user@example.com",
				"name": map[string]any{
					"givenName":  "  Alice  ",
					"familyName": "  Jones  ",
				},
			},
			want: &UserAttributes{
				UserName:    "user@example.com",
				Email:       "user@example.com",
				FirstName:   "  Alice  ",
				LastName:    "  Jones  ",
				DisplayName: "Alice     Jones",
				Active:      true,
			},
			wantErr: false,
		},
		{
			name: "displayName falls back to email when names empty",
			attributes: scim.ResourceAttributes{
				"userName": "user@EXAMPLE.com",
			},
			want: &UserAttributes{
				UserName:    "user@EXAMPLE.com",
				Email:       "user@EXAMPLE.com",
				DisplayName: "user@example.com",
				Active:      true,
			},
			wantErr: false,
		},
		{
			name: "userName defaults to email",
			attributes: scim.ResourceAttributes{
				"emails": []any{
					map[string]any{
						"value":   "primary@example.com",
						"primary": true,
					},
				},
			},
			want: &UserAttributes{
				UserName:    "primary@example.com",
				Email:       "primary@example.com",
				DisplayName: "primary@example.com",
				Active:      true,
			},
			wantErr: false,
		},
		{
			name: "active defaults to true",
			attributes: scim.ResourceAttributes{
				"userName": "user@example.com",
			},
			want: &UserAttributes{
				UserName:    "user@example.com",
				Email:       "user@example.com",
				DisplayName: "user@example.com",
				Active:      true,
			},
			wantErr: false,
		},
		{
			name: "active set to false",
			attributes: scim.ResourceAttributes{
				"userName": "user@example.com",
				"active":   false,
			},
			want: &UserAttributes{
				UserName:    "user@example.com",
				Email:       "user@example.com",
				DisplayName: "user@example.com",
				Active:      false,
			},
			wantErr: false,
		},
		{
			name: "error when no email found",
			attributes: scim.ResourceAttributes{
				"userName": "not-an-email",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractUserAttributes(tt.attributes)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestExtractPatchUserAttribute(t *testing.T) {
	tests := []struct {
		name    string
		op      scim.PatchOperation
		want    *PatchUserAttributes
		wantErr bool
	}{
		{
			name: "patch with userName",
			op: scim.PatchOperation{
				Op: "replace",
				Value: map[string]any{
					"userName": "user@example.com",
				},
			},
			want: &PatchUserAttributes{
				UserName: strPtr("user@example.com"),
				Email:    strPtr("user@example.com"),
			},
			wantErr: false,
		},
		{
			name: "patch with primary email",
			op: scim.PatchOperation{
				Op: "replace",
				Value: map[string]any{
					"emails": []any{
						map[string]any{
							"value":   "primary@example.com",
							"primary": true,
						},
					},
				},
			},
			want: &PatchUserAttributes{
				Email: strPtr("primary@example.com"),
			},
			wantErr: false,
		},
		{
			name: "patch with name fields",
			op: scim.PatchOperation{
				Op: "replace",
				Value: map[string]any{
					"name": map[string]any{
						"givenName":  "John",
						"familyName": "Doe",
					},
				},
			},
			want: &PatchUserAttributes{
				FirstName: strPtr("John"),
				LastName:  strPtr("Doe"),
			},
			wantErr: false,
		},
		{
			name: "patch with displayName",
			op: scim.PatchOperation{
				Op: "replace",
				Value: map[string]any{
					"displayName": "Custom Display",
				},
			},
			want: &PatchUserAttributes{
				DisplayName: strPtr("Custom Display"),
			},
			wantErr: false,
		},
		{
			name: "patch with active status",
			op: scim.PatchOperation{
				Op: "replace",
				Value: map[string]any{
					"active": false,
				},
			},
			want: &PatchUserAttributes{
				Active: boolPtr(false),
			},
			wantErr: false,
		},
		{
			name: "patch with optional fields",
			op: scim.PatchOperation{
				Op: "replace",
				Value: map[string]any{
					"externalId":        "ext123",
					"preferredLanguage": "en",
					"locale":            "en_US",
					"profileUrl":        "https://example.com/profile",
				},
			},
			want: &PatchUserAttributes{
				ExternalID:        strPtr("ext123"),
				PreferredLanguage: strPtr("en"),
				Locale:            strPtr("en_US"),
				ProfileURL:        strPtr("https://example.com/profile"),
			},
			wantErr: false,
		},
		{
			name: "error when userName is invalid email",
			op: scim.PatchOperation{
				Op: "replace",
				Value: map[string]any{
					"userName": "not-an-email",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "non-map value returns empty patch",
			op: scim.PatchOperation{
				Op:    "replace",
				Value: "string-value",
			},
			want:    &PatchUserAttributes{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractPatchUserAttribute(tt.op)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestExtractGroupAttributes(t *testing.T) {
	tests := []struct {
		name       string
		attributes scim.ResourceAttributes
		want       *GroupAttributes
		wantErr    bool
	}{
		{
			name: "complete group attributes",
			attributes: scim.ResourceAttributes{
				"displayName": "Engineering Team",
				"externalId":  "ext123",
				"active":      true,
				"members": []any{
					map[string]any{"value": "user1"},
					map[string]any{"value": "user2"},
				},
			},
			want: &GroupAttributes{
				DisplayName: "Engineering Team",
				ExternalID:  "ext123",
				Active:      true,
				MemberIDs:   []string{"user1", "user2"},
			},
			wantErr: false,
		},
		{
			name: "group with deduplicated members",
			attributes: scim.ResourceAttributes{
				"displayName": "Engineering Team",
				"members": []any{
					map[string]any{"value": "user1"},
					map[string]any{"value": "user2"},
					map[string]any{"value": "user1"},
				},
			},
			want: &GroupAttributes{
				DisplayName: "Engineering Team",
				Active:      true,
				MemberIDs:   []string{"user1", "user2"},
			},
			wantErr: false,
		},
		{
			name: "group without members",
			attributes: scim.ResourceAttributes{
				"displayName": "Engineering Team",
			},
			want: &GroupAttributes{
				DisplayName: "Engineering Team",
				Active:      true,
				MemberIDs:   nil,
			},
			wantErr: false,
		},
		{
			name: "active defaults to true",
			attributes: scim.ResourceAttributes{
				"displayName": "Engineering Team",
			},
			want: &GroupAttributes{
				DisplayName: "Engineering Team",
				Active:      true,
				MemberIDs:   nil,
			},
			wantErr: false,
		},
		{
			name: "active set to false",
			attributes: scim.ResourceAttributes{
				"displayName": "Engineering Team",
				"active":      false,
			},
			want: &GroupAttributes{
				DisplayName: "Engineering Team",
				Active:      false,
				MemberIDs:   nil,
			},
			wantErr: false,
		},
		{
			name:       "error when displayName missing",
			attributes: scim.ResourceAttributes{},
			want:       nil,
			wantErr:    true,
		},
		{
			name: "error when displayName empty",
			attributes: scim.ResourceAttributes{
				"displayName": "",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractGroupAttributes(tt.attributes)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

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

func TestExtractPatchGroupAttribute(t *testing.T) {
	tests := []struct {
		name string
		op   scim.PatchOperation
		want *PatchGroupAttributes
	}{
		{
			name: "patch with displayName",
			op: scim.PatchOperation{
				Op: "replace",
				Value: map[string]any{
					"displayName": "Updated Team",
				},
			},
			want: &PatchGroupAttributes{
				DisplayName: strPtr("Updated Team"),
			},
		},
		{
			name: "patch with externalId",
			op: scim.PatchOperation{
				Op: "replace",
				Value: map[string]any{
					"externalId": "ext456",
				},
			},
			want: &PatchGroupAttributes{
				ExternalID: strPtr("ext456"),
			},
		},
		{
			name: "patch with active status",
			op: scim.PatchOperation{
				Op: "replace",
				Value: map[string]any{
					"active": false,
				},
			},
			want: &PatchGroupAttributes{
				Active: boolPtr(false),
			},
		},
		{
			name: "patch with all fields",
			op: scim.PatchOperation{
				Op: "replace",
				Value: map[string]any{
					"displayName": "Updated Team",
					"externalId":  "ext789",
					"active":      true,
				},
			},
			want: &PatchGroupAttributes{
				DisplayName: strPtr("Updated Team"),
				ExternalID:  strPtr("ext789"),
				Active:      boolPtr(true),
			},
		},
		{
			name: "non-map value returns empty patch",
			op: scim.PatchOperation{
				Op:    "replace",
				Value: "string-value",
			},
			want: &PatchGroupAttributes{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractPatchGroupAttribute(tt.op)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Helper functions for creating pointers
func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
