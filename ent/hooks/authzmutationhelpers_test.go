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
			got, err := GetObjectIDFromEntValue(tt.input)
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
			name: "valid input, singular",
			ctx: func() context.Context {
				fCtx := &graphql.FieldContext{
					Args: map[string]interface{}{
						"input": map[string]interface{}{
							"parentField": "id1",
						},
					},
				}
				return graphql.WithFieldContext(context.Background(), fCtx)
			}(),
			parentField: "parentField",
			want:        []string{"id1"},
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

func TestCheckForEdge(t *testing.T) {
	tests := []struct {
		name        string
		parentField string
		edgeField   string
		want        string
	}{
		{
			name:        "edge matches parent field",
			parentField: "parent_id",
			edgeField:   "parent",
			want:        "parent",
		},
		{
			name:        "edge matches pluralized parent field",
			parentField: "parent_id",
			edgeField:   "parents",
			want:        "parents",
		},
		{
			name:        "edge does not match parent field",
			parentField: "parent_id",
			edgeField:   "child",
			want:        "",
		},
		{
			name:        "edge does not match pluralized parent field",
			parentField: "parent_id",
			edgeField:   "children",
			want:        "",
		},
		{
			name:        "edge matches parent field without _id",
			parentField: "parent",
			edgeField:   "parent",
			want:        "parent",
		},
		{
			name:        "edge matches pluralized parent field without _id",
			parentField: "parent",
			edgeField:   "parents",
			want:        "parents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkForEdge(tt.parentField, tt.edgeField)
			assert.Equal(t, tt.want, got)
		})
	}
}
