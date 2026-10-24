package hooks

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestIsQueryValid(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		variables map[string]any
		hasError  bool
	}{
		{
			name: "can store a valid query without variables",
			query: `{
  controls {
    totalCount
  }
}`,
		},
		{
			name: "can store a valid query that has variables",
			query: `query Report($include: Boolean!) {
  controls @include(if: $include) {
    totalCount
  }
}`,
			variables: map[string]any{
				"include": true,
			},
		},
		{
			name:     "cannot store an invalid query",
			query:    `query {`,
			hasError: true,
		},
		{
			name: "unknown field in query should fail validation",
			query: `{
  nonexistentReportField
}`,
			hasError: true,
		},
		{
			name: "report should not allow for mutation",
			query: `mutation {
  __typename
}`,
			hasError: true,
		},
		{
			name: "multiple queries are not allowed",
			query: `query A {
  controls {
    totalCount
  }
}
query B {
  controls {
    totalCount
  }
}`,
			hasError: true,
		},
		{
			name: "query that needs a variable must have it present",
			query: `query Report($include: Boolean!) {
  controls @include(if: $include) {
    totalCount
  }
}`,
			hasError: true,
		},
		{
			name: "query with invalid variable type should fail",
			query: `query Report($include: Boolean!) {
  controls @include(if: $include) {
    totalCount
  }
}`,
			variables: map[string]any{
				"include": "yes",
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := isQueryValid(tt.query, tt.variables)
			assert.Equal(t, tt.hasError, err != nil)
		})
	}
}
