package hooks

import (
	"context"
	"testing"

	"entgo.io/ent"
	"github.com/99designs/gqlgen/graphql"
	"github.com/stretchr/testify/assert"
)

func TestGetObjectIDFromEntValue(t *testing.T) {
	tests := []struct {
		name    string
		input   ent.Value
		want    string
		wantErr bool
	}{
		{
			name: "valid object ID",
			input: map[string]interface{}{
				"id": "12345",
			},
			want:    "12345",
			wantErr: false,
		},
		{
			name: "missing object ID",
			input: map[string]interface{}{
				"name": "test",
			},
			want:    "",
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   make(chan int),
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getObjectIDFromEntValue(tt.input)
			if tt.wantErr {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseGraphqlInputForEdgeIDs(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		parentField string
		want        []string
		wantErr     bool
	}{
		{
			name: "valid input",
			ctx: func() context.Context {
				fCtx := &graphql.FieldContext{
					Args: map[string]interface{}{
						"input": map[string]interface{}{
							"parentField": []string{"id1", "id2"},
						},
					},
				}
				return graphql.WithFieldContext(context.Background(), fCtx)
			}(),
			parentField: "parentField",
			want:        []string{"id1", "id2"},
			wantErr:     false,
		},
		{
			name: "missing input",
			ctx: func() context.Context {
				fCtx := &graphql.FieldContext{
					Args: map[string]interface{}{},
				}
				return graphql.WithFieldContext(context.Background(), fCtx)
			}(),
			parentField: "parentField",
			want:        nil,
			wantErr:     false,
		},
		{
			name: "invalid JSON",
			ctx: func() context.Context {
				fCtx := &graphql.FieldContext{
					Args: map[string]interface{}{
						"input": make(chan int),
					},
				}
				return graphql.WithFieldContext(context.Background(), fCtx)
			}(),
			parentField: "parentField",
			want:        nil,
			wantErr:     true,
		},
		{
			name: "missing parent field",
			ctx: func() context.Context {
				fCtx := &graphql.FieldContext{
					Args: map[string]interface{}{
						"input": map[string]interface{}{
							"otherField": []string{"id1", "id2"},
						},
					},
				}
				return graphql.WithFieldContext(context.Background(), fCtx)
			}(),
			parentField: "parentField",
			want:        nil,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGraphqlInputForEdgeIDs(tt.ctx, tt.parentField)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
