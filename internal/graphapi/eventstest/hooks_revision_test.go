//go:build test

package eventstest_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

// TestInternalPolicyEdgeOnlyUpdateDoesNotBumpRevision proves HookRevisionUpdate does not bump the revision for an edge-only mutation
func TestInternalPolicyEdgeOnlyUpdateDoesNotBumpRevision(t *testing.T) {
	docUser := suite.UserBuilder(context.Background(), t)
	ctx := th.SetContext(docUser.UserCtx, suite.Client.DB)

	integration, err := suite.Client.DB.Integration.Create().
		SetName("Revision Edge Test").
		SetKind("revedgetest").
		SetDefinitionID("def_revedgetest").
		Save(ctx)
	assert.NilError(t, err)

	details := "the revision must not bump for an edge-only mutation"

	resp, err := suite.Client.API.CreateInternalPolicy(docUser.UserCtx, testclient.CreateInternalPolicyInput{
		Name:    "revision edge policy",
		Details: &details,
	})
	assert.NilError(t, err)

	policyID := resp.CreateInternalPolicy.InternalPolicy.ID
	before := lo.FromPtr(resp.CreateInternalPolicy.InternalPolicy.Revision)

	t.Run("edge-only update leaves the revision unchanged", func(t *testing.T) {
		err := suite.Client.DB.InternalPolicy.UpdateOneID(policyID).AddIntegrationIDs(integration.ID).Exec(ctx)
		assert.NilError(t, err)

		policy, err := suite.Client.DB.InternalPolicy.Get(ctx, policyID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(before, policy.Revision), "an edge-only mutation must not bump the revision")

		integrationIDs, err := suite.Client.DB.InternalPolicy.QueryIntegrations(policy).IDs(ctx)
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual([]string{integration.ID}, integrationIDs))
	})

	t.Run("a details change still bumps the revision", func(t *testing.T) {
		err := suite.Client.DB.InternalPolicy.UpdateOneID(policyID).SetDetails("the revision must bump for a details change").Exec(ctx)
		assert.NilError(t, err)

		policy, err := suite.Client.DB.InternalPolicy.Get(ctx, policyID)
		assert.NilError(t, err)
		assert.Check(t, before != policy.Revision, "a details change must bump the revision")
	})

	th.CleanupOrganizationDataWithContext(docUser.UserCtx, t)
}
