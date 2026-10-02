//go:build test

package hooks_test

import (
	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/predicate"
	"github.com/theopenlane/core/v2/internal/ent/generated/tagdefinition"
)

type selectorCallerProbe struct {
	calls      int
	withCaller int
}

// predicate increment the withCaller so if the context was dropped
// the probe sees no caller, mimicking the bug with Count queries
// having no caller before selector_context.tmpl was added
func (p *selectorCallerProbe) predicate() predicate.TagDefinition {
	return func(s *sql.Selector) {
		p.calls++

		if _, ok := auth.CallerFromContext(s.Context()); ok {
			p.withCaller++
		}
	}
}

func (suite *HookTestSuite) TestQuerySelectorCarriesRequestContext() {
	t := suite.T()

	user := suite.seedUser()
	orgID := user.Edges.OrgMemberships[0].OrganizationID

	ctx := generated.NewContext(auth.NewTestContextWithOrgID(user.ID, orgID), suite.client)

	countIDsProbe := &selectorCallerProbe{}
	_, err := suite.client.TagDefinition.Query().
		Where(tagdefinition.OwnerID(orgID), countIDsProbe.predicate()).
		CountIDs(ctx)
	assert.NilError(t, err)
	assert.Check(t, countIDsProbe.calls > 0)
	assert.Check(t, is.Equal(countIDsProbe.withCaller, countIDsProbe.calls))

	existProbe := &selectorCallerProbe{}
	_, err = suite.client.TagDefinition.Query().
		Where(tagdefinition.OwnerID(orgID), existProbe.predicate()).
		Exist(ctx)
	assert.NilError(t, err)
	assert.Check(t, existProbe.calls > 0)
	assert.Check(t, is.Equal(existProbe.withCaller, existProbe.calls))

	groupByProbe := &selectorCallerProbe{}
	_, err = suite.client.TagDefinition.Query().
		Where(tagdefinition.OwnerID(orgID), groupByProbe.predicate()).
		GroupBy(tagdefinition.FieldName).
		Strings(ctx)
	assert.NilError(t, err)
	assert.Check(t, groupByProbe.calls > 0)
	assert.Check(t, is.Equal(groupByProbe.withCaller, groupByProbe.calls))
}
