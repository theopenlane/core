//go:build test

package eventstest_test

import (
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	openapi "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/finding"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
)

// TestFindingSameDefinitionTwoTenantsIsolated verifies two installations sharing one definition but
// distinct source_instance_id tenants never overwrite each other's Finding rows, even though the
// catalog upsert's lookup key (owner + external id) alone would otherwise match the same row
func TestFindingSameDefinitionTwoTenantsIsolated(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const tenant1 = "tenant-findtwotenant-1"
	const tenant2 = "tenant-findtwotenant-2"

	installationT1, err := suite.Client.DB.Integration.Create().
		SetName("Finding Two Tenants T1").
		SetKind("findtwotenantt1").
		SetDefinitionID("def_findtwotenant").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: tenant1}}).
		Save(ctx)
	th.RequireNoError(t, err)

	installationT2, err := suite.Client.DB.Integration.Create().
		SetName("Finding Two Tenants T2").
		SetKind("findtwotenantt2").
		SetDefinitionID("def_findtwotenant").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: tenant2}}).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		findings, err := suite.Client.DB.Finding.Query().Where(finding.ExternalIDHasPrefix("find-twotenant-")).All(ctx)
		th.RequireNoError(t, err)

		if len(findings) > 0 {
			(&th.Cleanup[*ent.FindingDeleteOne]{Client: suite.Client.DB.Finding, IDs: lo.Map(findings, func(f *ent.Finding, _ int) string { return f.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installationT1.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installationT2.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	resultT1 := ingestFindingPayloads(ctx, t, installationT1, `{"external_id":"find-twotenant-1","display_name":"Owned By T1","description":"from T1"}`)
	assert.Check(t, is.Equal(1, resultT1.Changed), "T1's create must count as changed")

	before := findingByExternalID(ctx, t, "find-twotenant-1")

	linkedToT1, err := suite.Client.DB.Finding.Query().Where(finding.ID(before.ID), finding.HasIntegrationsWith(integration.ID(installationT1.ID))).Exist(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, linkedToT1, "T1's create must add its integration edge")

	resultT2 := ingestFindingPayloads(ctx, t, installationT2, `{"external_id":"find-twotenant-1","display_name":"Owned By T2","description":"from T2"}`)
	assert.Check(t, is.Equal(1, resultT2.Skipped), "a different tenant sharing the definition must be read-only for T1's row")
	assert.Check(t, is.Equal(0, resultT2.Changed))

	after := findingByExternalID(ctx, t, "find-twotenant-1")
	assert.Check(t, is.Equal(before.Description, after.Description), "T2's payload must not change T1's fields")
	assert.Check(t, is.Equal(before.DisplayName, after.DisplayName))
	assert.Check(t, after.UpdatedAt.Equal(before.UpdatedAt))

	linkedToT2, err := suite.Client.DB.Finding.Query().Where(finding.ID(after.ID), finding.HasIntegrationsWith(integration.ID(installationT2.ID))).Exist(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, !linkedToT2, "a foreign-tenant ingest must not add its integration edge")
}
