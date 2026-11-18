package schemaparse

import (
	"testing"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc/load"
	"github.com/theopenlane/entx"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestCheckEntGQLAnnotation(t *testing.T) {
	testCases := []struct {
		name            string
		annotation      *entgql.Annotation
		expectedExclude bool
	}{
		{
			name:            "no entgql annotation",
			expectedExclude: false,
		},
		{
			name:            "skip all excludes",
			annotation:      &entgql.Annotation{Skip: entgql.SkipAll},
			expectedExclude: true,
		},
		{
			name:            "skip type excludes",
			annotation:      &entgql.Annotation{Skip: entgql.SkipType},
			expectedExclude: true,
		},
		{
			name:            "skip mutation inputs does not exclude",
			annotation:      &entgql.Annotation{Skip: entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput},
			expectedExclude: false,
		},
		{
			name:            "skip where input does not exclude",
			annotation:      &entgql.Annotation{Skip: entgql.SkipWhereInput},
			expectedExclude: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema := &load.Schema{Name: "TestSchema", Annotations: map[string]any{}}
			if tc.annotation != nil {
				schema.Annotations[tc.annotation.Name()] = tc.annotation
			}

			si := parsedSchema{}
			si.checkEntGQLAnnotation(schema)

			assert.Check(t, is.Equal(tc.expectedExclude, si.ExcludeFromGeneration))
		})
	}
}

func TestCheckSkipAnnotation(t *testing.T) {
	testCases := []struct {
		name             string
		skip             entx.FGASkipMode
		expectedExclude  bool
		expectedViewOnly bool
	}{
		{
			name:            "skip all excludes",
			skip:            entx.SkipAll,
			expectedExclude: true,
		},
		{
			name:             "skip create update and delete is view only",
			skip:             entx.SkipCreate | entx.SkipUpdate | entx.SkipDelete,
			expectedViewOnly: true,
		},
		{
			name: "skip create and delete is not view only",
			skip: entx.SkipCreate | entx.SkipDelete,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ant := entx.FGACrudSkip(tc.skip)
			schema := &load.Schema{Name: "TestSchema", Annotations: map[string]any{ant.Name(): ant}}

			si := parsedSchema{}
			si.checkSkipAnnotation(schema)

			assert.Check(t, is.Equal(tc.expectedExclude, si.ExcludeFromGeneration))
			assert.Check(t, is.Equal(tc.expectedViewOnly, si.ViewOnly))
		})
	}
}
