//go:build test

package hooks_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/generated"
)

func (suite *HookTestSuite) TestHookReportCreate() {

	t := suite.T()

	user := suite.seedSystemAdmin()
	orgID := user.Edges.OrgMemberships[0].OrganizationID
	ctx := auth.NewTestContextForSystemAdmin(user.ID, orgID)
	ctx = generated.NewContext(ctx, suite.client)

	tests := []struct {
		name      string
		query     string
		variables map[string]any
		hasError  bool
	}{
		{
			name:  "storing a valid query",
			query: `{ controls { totalCount } }`,
		},
		{
			name:      "storing a valid query with variables",
			query:     `query Report($include: Boolean!) { controls @include(if: $include) { totalCount } }`,
			variables: map[string]any{"include": true},
		},
		{
			name:     "cannot store an invalid query",
			query:    `query {`,
			hasError: true,
		},
		{
			name:     "unknown field",
			query:    `{ nonexistentReportField }`,
			hasError: true,
		},
		{
			name:     "mutation rejected",
			query:    `mutation { __typename }`,
			hasError: true,
		},
		{
			name:     "multiple operations rejected",
			query:    `query A { controls { totalCount } } query B { controls { totalCount } }`,
			hasError: true,
		},
		{
			name:     "missing required variable",
			query:    `query Report($include: Boolean!) { controls @include(if: $include) { totalCount } }`,
			hasError: true,
		},
		{
			name:      "wrong variable type",
			query:     `query Report($include: Boolean!) { controls @include(if: $include) { totalCount } }`,
			variables: map[string]any{"include": "yes"},
			hasError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := suite.client.Report.Create().
				SetName(tt.name).
				SetOwnerID(orgID).
				SetUserID(user.ID).
				SetQuery(tt.query).
				SetVariables(tt.variables).
				Save(ctx)
			if tt.hasError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.query, report.Query)
			assert.Equal(t, tt.variables, report.Variables)
		})
	}
}
