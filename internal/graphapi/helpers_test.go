package graphapi

import (
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/vektah/gqlparser/v2/ast"
)

func TestGetArgumentName(t *testing.T) {
	tests := []struct {
		name     string
		op       *graphql.OperationContext
		key      string
		expected string
	}{
		{
			name: "Argument found",
			op: &graphql.OperationContext{
				Operation: &ast.OperationDefinition{
					SelectionSet: ast.SelectionSet{
						&ast.Field{
							Arguments: ast.ArgumentList{
								&ast.Argument{
									Name:  "arg1",
									Value: &ast.Value{Raw: "key1"},
								},
							},
						},
					},
				},
			},
			key:      "key1",
			expected: "arg1",
		},
		{
			name: "Argument not found",
			op: &graphql.OperationContext{
				Operation: &ast.OperationDefinition{
					SelectionSet: ast.SelectionSet{
						&ast.Field{
							Arguments: ast.ArgumentList{
								&ast.Argument{
									Name:  "arg1",
									Value: &ast.Value{Raw: "key1"},
								},
							},
						},
					},
				},
			},
			key:      "key2",
			expected: "",
		},
		{
			name: "empty arguments",
			op: &graphql.OperationContext{
				Operation: &ast.OperationDefinition{
					SelectionSet: ast.SelectionSet{
						&ast.Field{
							Arguments: ast.ArgumentList{},
						},
					},
				},
			},
			key:      "key1",
			expected: "",
		},
		{
			name: "Argument found",
			op: &graphql.OperationContext{
				Operation: &ast.OperationDefinition{
					SelectionSet: ast.SelectionSet{
						&ast.Field{
							Arguments: ast.ArgumentList{
								&ast.Argument{
									Name:  "arg1",
									Value: &ast.Value{Raw: "key1"},
								},
							},
						},
					},
				},
			},
			key:      "key1",
			expected: "arg1",
		},
		{
			name: "empty selection set",
			op: &graphql.OperationContext{
				Operation: &ast.OperationDefinition{
					SelectionSet: nil,
				},
			},
			key:      "key1",
			expected: "",
		},
		{
			name: "empty operation",
			op: &graphql.OperationContext{
				Operation: nil,
			},
			key:      "key1",
			expected: "",
		},
		{
			name:     "empty operation context",
			op:       nil,
			key:      "key1",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getArgumentName(tt.op, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}
