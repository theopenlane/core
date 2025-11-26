package resolvers


import (
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"github.com/vektah/gqlparser/v2/ast"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
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
		{
			name:     "Create Upload operation",
			input:    "createUploadProcedure",
			expected: "Procedure",
		},
		{
			name:     "Create Upload Document",
			input:    "createUploadDocument",
			expected: "Document",
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
		expected    *pkgobjects.File
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
			expected: &pkgobjects.File{
				CorrelatedObjectType: "User",
				FileMetadata: pkgobjects.FileMetadata{
					Key: "file",
				},
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
			expected:    &pkgobjects.File{},
			expectedErr: ErrUnableToDetermineObjectType,
		},
		{
			name:        "No upload argument",
			fieldName:   "createUser",
			key:         "file",
			arguments:   ast.ArgumentList{},
			expected:    &pkgobjects.File{},
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
			expected:    &pkgobjects.File{},
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

			upload := &pkgobjects.File{
				OriginalName: "meow.txt",
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

func TestIsEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name:     "Nil value",
			input:    nil,
			expected: true,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "Non-empty string",
			input:    "hello",
			expected: false,
		},
		{
			name:     "Zero integer",
			input:    0,
			expected: true,
		},
		{
			name:     "Non-zero integer",
			input:    42,
			expected: false,
		},
		{
			name:     "Zero float",
			input:    0.0,
			expected: true,
		},
		{
			name:     "Non-zero float",
			input:    3.14,
			expected: false,
		},
		{
			name:     "Empty slice",
			input:    []int{},
			expected: true,
		},
		{
			name:     "Non-empty slice",
			input:    []int{1, 2, 3},
			expected: false,
		},
		{
			name:     "Empty map",
			input:    map[string]any{},
			expected: true,
		},
		{
			name:     "Non-empty map",
			input:    map[string]int{"a": 1},
			expected: false,
		},
		{
			name:     "Boolean false",
			input:    false,
			expected: true,
		},
		{
			name:     "Boolean true",
			input:    true,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEmpty(tt.input)
			assert.Check(t, is.Equal(tt.expected, result))
		})
	}
}
