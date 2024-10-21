package graphapi

import (
	"fmt"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/vektah/gqlparser/v2/ast"
)

func TestStripOperation(t *testing.T) {
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
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRetrieveObjectDetails(t *testing.T) {
	tests := []struct {
		name        string
		fieldName   string
		key         string
		arguments   ast.ArgumentList
		expected    *objects.FileUpload
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
			expected: &objects.FileUpload{
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
			expected:    &objects.FileUpload{},
			expectedErr: fmt.Errorf("unable to determine object type"),
		},
		{
			name:        "No upload argument",
			fieldName:   "createUser",
			key:         "file",
			arguments:   ast.ArgumentList{},
			expected:    &objects.FileUpload{},
			expectedErr: fmt.Errorf("unable to determine object type"),
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
			expected:    &objects.FileUpload{},
			expectedErr: fmt.Errorf("unable to determine object type"),
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

			upload := &objects.FileUpload{
				Filename: "meow.txt",
			}

			result, err := retrieveObjectDetails(rctx, tt.key, upload)
			if tt.expectedErr != nil {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected.CorrelatedObjectType, result.CorrelatedObjectType)
			assert.Equal(t, tt.expected.Key, result.Key)
		})
	}
}
