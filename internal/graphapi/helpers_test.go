package graphapi

import (
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"github.com/vektah/gqlparser/v2/ast"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/objects/storage"
)

func TestStripOperation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Create operation",
			input:    "createUser",
			expected: "User",
		},
		{
			name:     "Update operation",
			input:    "updateUser",
			expected: "User",
		},
		{
			name:     "Delete operation",
			input:    "deleteUser",
			expected: "User",
		},
		{
			name:     "Get operation",
			input:    "getUser",
			expected: "User",
		},
		{
			name:     "No operation",
			input:    "User",
			expected: "User",
		},
		{
			name:     "Non-matching prefix",
			input:    "fetchUser",
			expected: "fetchUser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripOperation(tt.input)
			assert.Check(t, is.Equal(tt.expected, result))
		})
	}
}

func TestRetrieveObjectDetails(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fieldName   string
		key         string
		arguments   ast.ArgumentList
		expected    *storage.File
		expectedErr error
	}{
		{
			name:      "Matching upload argument",
			fieldName: "createUser",
			key:       "file",
			arguments: ast.ArgumentList{
				&ast.Argument{
					Name: "file",
					Value: &ast.Value{
						ExpectedType: &ast.Type{
							NamedType: "Upload",
						},
					},
				},
			},
			expected: &storage.File{
				CorrelatedObjectType: "User",
				Key:                  "file",
			},
			expectedErr: nil,
		},
		{
			name:      "Non-matching upload argument",
			fieldName: "createUser",
			key:       "image",
			arguments: ast.ArgumentList{
				&ast.Argument{
					Name: "file",
					Value: &ast.Value{
						ExpectedType: &ast.Type{
							NamedType: "Upload",
						},
					},
				},
			},
			expected:    &storage.File{},
			expectedErr: ErrUnableToDetermineObjectType,
		},
		{
			name:        "No upload argument",
			fieldName:   "createUser",
			key:         "file",
			arguments:   ast.ArgumentList{},
			expected:    &storage.File{},
			expectedErr: ErrUnableToDetermineObjectType,
		},
		{
			name:      "Non-upload argument",
			fieldName: "createUser",
			key:       "file",
			arguments: ast.ArgumentList{
				&ast.Argument{
					Name: "file",
					Value: &ast.Value{
						ExpectedType: &ast.Type{
							NamedType: "String",
						},
					},
				},
			},
			expected:    &storage.File{},
			expectedErr: ErrUnableToDetermineObjectType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rctx := &graphql.FieldContext{
				Field: graphql.CollectedField{
					Field: &ast.Field{
						Name:      tt.fieldName,
						Arguments: tt.arguments,
					},
				},
			}

			upload := &storage.File{
				Filename: "meow.txt",
			}

			result, err := retrieveObjectDetails(rctx, tt.key, upload)
			if tt.expectedErr != nil {

				return
			}

			assert.NilError(t, err)
			assert.Check(t, is.Equal(tt.expected.CorrelatedObjectType, result.CorrelatedObjectType))
			assert.Check(t, is.Equal(tt.expected.Key, result.Key))
		})
	}
}
func TestGetOrgOwnerFromInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       any
		expected    *string
		expectedErr error
	}{
		{
			name:        "Nil input",
			input:       nil,
			expected:    nil,
			expectedErr: nil,
		},
		{
			name: "Valid input with owner ID",
			input: generated.CreateProcedureInput{
				Name:    "Test Procedure",
				OwnerID: lo.ToPtr("owner123"),
			},
			expected:    lo.ToPtr("owner123"),
			expectedErr: nil,
		},
		{
			name:  "Valid input without owner ID",
			input: generated.CreateRiskInput{
				// No OwnerID field set
			},
			expected:    nil,
			expectedErr: nil,
		},
		{
			name: "Invalid input type will return nil",
			input: struct {
				Name string `json:"name"`
			}{
				Name: "test",
			},
			expected:    nil,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getOrgOwnerFromInput(&tt.input)
			if tt.expectedErr != nil {

				assert.Check(t, is.Nil(result))
				return
			}

			assert.NilError(t, err)
			assert.Check(t, is.DeepEqual(tt.expected, result))
		})
	}
}
func TestGetBulkUploadOwnerInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       []*generated.CreateProcedureInput // used as an example, should work with any type
		expected    *string
		expectedErr error
	}{
		{
			name:        "Nil input, nothing to do",
			input:       nil,
			expected:    nil,
			expectedErr: nil,
		},
		{
			name: "Valid input with consistent owner IDs",
			input: []*generated.CreateProcedureInput{
				{
					Name:    "Test Procedure 1",
					OwnerID: lo.ToPtr("owner123"),
				},
				{
					Name:    "Test Procedure 2",
					OwnerID: lo.ToPtr("owner123"),
				},
			},
			expected:    lo.ToPtr("owner123"),
			expectedErr: nil,
		},
		{
			name: "Valid input with inconsistent owner IDs",
			input: []*generated.CreateProcedureInput{
				{
					Name:    "Test Procedure 1",
					OwnerID: lo.ToPtr("owner123"),
				},
				{
					Name:    "Test Procedure 2",
					OwnerID: lo.ToPtr("owner456"),
				},
			},
			expected:    nil,
			expectedErr: ErrNoOrganizationID,
		},
		{
			name: "Valid input with missing owner ID",
			input: []*generated.CreateProcedureInput{
				{
					Name: "Test Procedure 1",
				},
			},
			expected:    nil,
			expectedErr: ErrNoOrganizationID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getBulkUploadOwnerInput(tt.input)
			if tt.expectedErr != nil {

				assert.Check(t, is.Nil(result))
				return
			}

			assert.NilError(t, err)
			assert.Check(t, is.DeepEqual(tt.expected, result))
		})
	}
}
