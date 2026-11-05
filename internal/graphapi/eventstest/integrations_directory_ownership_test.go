//go:build test

package eventstest_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	openapi "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
)

// directoryAccountByExternalIDAndIntegration loads one ingested directory account scoped to a
// specific installation, disambiguating rows that share an external id across tenants
func directoryAccountByExternalIDAndIntegration(ctx context.Context, t *testing.T, externalID string, integrationID string) *ent.DirectoryAccount {
	t.Helper()

	da, err := suite.Client.DB.DirectoryAccount.Query().
		Where(directoryaccount.ExternalID(externalID), directoryaccount.IntegrationID(integrationID)).
		Only(ctx)
	th.RequireNoError(t, err)

	return da
}

// TestDirectorySameDefinitionTwoTenantsIsolated verifies two installations sharing one definition
// but distinct source_instance_id tenants never see each other's rows even when every external id
// in their snapshots is identical
func TestDirectorySameDefinitionTwoTenantsIsolated(t *testing.T) {
	ctx := th.SetInternalContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const prefix = "sdtwotenant"
	const tenant1 = "tenant-" + prefix + "-1"
	const tenant2 = "tenant-" + prefix + "-2"

	installationT1, err := suite.Client.DB.Integration.Create().
		SetName("Same Definition Two Tenants T1").
		SetKind("sdtwotenantt1").
		SetDefinitionID("def_dirsynctest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: tenant1}}).
		Save(ctx)
	th.RequireNoError(t, err)

	installationT2, err := suite.Client.DB.Integration.Create().
		SetName("Same Definition Two Tenants T2").
		SetKind("sdtwotenantt2").
		SetDefinitionID("def_dirsynctest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: tenant2}}).
		Save(ctx)
	th.RequireNoError(t, err)

	cleanupDirectoryPrefix(t, ctx, []string{installationT1.ID, installationT2.ID}, prefix)

	counters, teardown := directoryEventCounters(t)
	defer teardown()

	base := newDirectorySnapshot(prefix)

	resultT1 := ingestDirectorySnapshotFixture(ctx, t, installationT1, base, true)
	waitForGala(t, counters.Runtime)
	assert.Check(t, is.Equal(0, resultT1.Failed))
	assert.Check(t, is.Equal(len(base.Accounts)+len(base.Groups)+len(base.Memberships), resultT1.Changed), "T1 must create its own rows")

	resultT2 := ingestDirectorySnapshotFixture(ctx, t, installationT2, base, true)
	waitForGala(t, counters.Runtime)
	assert.Check(t, is.Equal(0, resultT2.Failed))
	assert.Check(t, is.Equal(len(base.Accounts)+len(base.Groups)+len(base.Memberships), resultT2.Changed), "T2 must create its own rows despite sharing the definition and every external id with T1")

	assert.Check(t, is.Equal(int64(len(base.Accounts)*2), counters.AccountCreates.Load()))
	assert.Check(t, is.Equal(int64(len(base.Groups)*2), counters.GroupCreates.Load()))
	assert.Check(t, is.Equal(int64(len(base.Memberships)*2), counters.MembershipCreates.Load()))

	changedExternalID := base.Accounts[0].ExternalID
	materialChange := base.withAccountMaterialChange(changedExternalID)

	t1Before := directoryAccountByExternalIDAndIntegration(ctx, t, changedExternalID, installationT1.ID)

	thirdResult := ingestDirectorySnapshotFixture(ctx, t, installationT2, materialChange, true)
	waitForGala(t, counters.Runtime)

	assert.Check(t, is.Equal(0, thirdResult.Failed))
	assert.Check(t, is.Equal(1, thirdResult.Changed), "a material change under T2 must count as exactly one changed record")
	assert.Check(t, is.Equal(int64(1), counters.AccountUpdates.Load()), "the material change must emit exactly one account update event, scoped to T2")

	t2After := directoryAccountByExternalIDAndIntegration(ctx, t, changedExternalID, installationT2.ID)
	assert.Check(t, is.Equal(materialChange.account(changedExternalID).DisplayName, t2After.DisplayName))

	t1After := directoryAccountByExternalIDAndIntegration(ctx, t, changedExternalID, installationT1.ID)
	assert.Check(t, is.Equal(t1Before.DisplayName, t1After.DisplayName), "T2's material change must not affect T1's isolated row")
	assert.Check(t, t1After.UpdatedAt.Equal(t1Before.UpdatedAt), "T1's row must not be touched by T2's ingest")
}

// TestDirectoryRowsProtectedFromOtherDefinition verifies rows an installation wrote are read-only for a different
// definition sharing the same tenant
func TestDirectoryRowsProtectedFromOtherDefinition(t *testing.T) {
	ctx := th.SetInternalContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const prefix = "otherdefinition"
	const tenant = "tenant-" + prefix

	installationA, err := suite.Client.DB.Integration.Create().
		SetName("Other Definition A").
		SetKind("otherdefinitiona").
		SetDefinitionID("def_dirsynctest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: tenant}}).
		Save(ctx)
	th.RequireNoError(t, err)

	installationC, err := suite.Client.DB.Integration.Create().
		SetName("Other Definition C").
		SetKind("otherdefinitionc").
		SetDefinitionID("def_dirsynctest_other").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: tenant}}).
		Save(ctx)
	th.RequireNoError(t, err)

	cleanupDirectoryPrefix(t, ctx, []string{installationA.ID, installationC.ID}, prefix)

	base := newDirectorySnapshot(prefix)

	seeded := ingestDirectorySnapshotFixture(ctx, t, installationA, base, true)
	waitForGala(t, suite.GalaRuntime)

	assert.Check(t, is.Equal(0, seeded.Failed))
	assert.Check(t, is.Equal(len(base.Accounts)+len(base.Groups)+len(base.Memberships), seeded.Changed))

	counters, teardown := directoryEventCounters(t)
	defer teardown()

	claimedAccounts := lo.Map(base.Accounts, func(a directoryAccountRecord, _ int) *ent.DirectoryAccount {
		row := directoryAccountByExternalID(ctx, t, a.ExternalID)
		assert.Check(t, is.Equal(installationA.DefinitionID, row.SourceDefinitionID))
		assert.Check(t, is.Equal(tenant, row.SourceInstanceID))
		assert.Check(t, is.Equal(installationA.ID, row.ManagedBy))

		return row
	})
	claimedGroups := lo.Map(base.Groups, func(g directoryGroupRecord, _ int) *ent.DirectoryGroup {
		row := directoryGroupByExternalID(ctx, t, g.ExternalID)
		assert.Check(t, is.Equal(installationA.DefinitionID, row.SourceDefinitionID))
		assert.Check(t, is.Equal(tenant, row.SourceInstanceID))
		assert.Check(t, is.Equal(installationA.ID, row.ManagedBy))

		return row
	})
	claimedMemberships := lo.Map(base.Memberships, func(m directoryMembershipRecord, _ int) *ent.DirectoryMembership {
		row := directoryMembershipByExternalIDs(ctx, t, m.DirectoryAccountID, m.DirectoryGroupID)
		assert.Check(t, is.Equal(installationA.DefinitionID, row.SourceDefinitionID))
		assert.Check(t, is.Equal(tenant, row.SourceInstanceID))
		assert.Check(t, is.Equal(installationA.ID, row.ManagedBy))

		return row
	})

	protectedResult := ingestDirectorySnapshotFixture(ctx, t, installationC, base, true)
	waitForGala(t, counters.Runtime)

	assert.Check(t, is.Equal(0, protectedResult.Failed))
	assert.Check(t, is.Equal(0, protectedResult.Changed), "a foreign definition on A's tenant must not claim or modify A's rows")
	assert.Check(t, is.Equal(len(base.Accounts)+len(base.Groups)+len(base.Memberships), protectedResult.Skipped), "every record must be skipped as managed by another definition")
	assert.Check(t, is.Equal(int64(0), counters.AccountCreates.Load()))
	assert.Check(t, is.Equal(int64(0), counters.GroupCreates.Load()))
	assert.Check(t, is.Equal(int64(0), counters.MembershipCreates.Load()))
	assert.Check(t, is.Equal(int64(0), counters.AccountUpdates.Load()), "a foreign definition must not be able to write A's claimed rows")
	assert.Check(t, is.Equal(int64(0), counters.GroupUpdates.Load()))
	assert.Check(t, is.Equal(int64(0), counters.MembershipUpdates.Load()))

	for i, a := range base.Accounts {
		after := directoryAccountByExternalID(ctx, t, a.ExternalID)
		assert.Check(t, is.Equal(claimedAccounts[i].ID, after.ID))
		assert.Check(t, is.Equal(installationA.DefinitionID, after.SourceDefinitionID), "source_definition_id must still belong to A")
		assert.Check(t, is.Equal(installationA.ID, after.IntegrationID))
		assert.Check(t, after.UpdatedAt.Equal(claimedAccounts[i].UpdatedAt))
	}
	for i, g := range base.Groups {
		after := directoryGroupByExternalID(ctx, t, g.ExternalID)
		assert.Check(t, is.Equal(claimedGroups[i].ID, after.ID))
		assert.Check(t, is.Equal(installationA.DefinitionID, after.SourceDefinitionID))
		assert.Check(t, is.Equal(installationA.ID, after.IntegrationID))
		assert.Check(t, after.UpdatedAt.Equal(claimedGroups[i].UpdatedAt))
	}
	for i, m := range base.Memberships {
		after := directoryMembershipByExternalIDs(ctx, t, m.DirectoryAccountID, m.DirectoryGroupID)
		assert.Check(t, is.Equal(claimedMemberships[i].ID, after.ID))
		assert.Check(t, is.Equal(installationA.DefinitionID, after.SourceDefinitionID))
		assert.Check(t, is.Equal(installationA.ID, after.IntegrationID))
		assert.Check(t, after.UpdatedAt.Equal(claimedMemberships[i].UpdatedAt))
	}
}

// TestDirectoryRemoveAndReaddKeepsFlowing verifies a same-tenant, same-definition reinstall keeps
// ingesting normally when the very next sync after reinstall carries a material change, converging
// every row onto the new installation without duplicating rows or removal-inferring memberships, and
// a membership first seen after the reinstall links to the rows the old installation wrote
func TestDirectoryRemoveAndReaddKeepsFlowing(t *testing.T) {
	ctx := th.SetInternalContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const prefix = "removereadd"
	const tenant = "tenant-" + prefix

	installationA, err := suite.Client.DB.Integration.Create().
		SetName("Remove Readd A").
		SetKind("removereadda").
		SetDefinitionID("def_dirsynctest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: tenant}}).
		Save(ctx)
	th.RequireNoError(t, err)

	base := newDirectorySnapshot(prefix)

	seeded := ingestDirectorySnapshotFixture(ctx, t, installationA, base, true)
	waitForGala(t, suite.GalaRuntime)
	assert.Check(t, is.Equal(0, seeded.Failed))
	assert.Check(t, is.Equal(len(base.Accounts)+len(base.Groups)+len(base.Memberships), seeded.Changed))

	th.RequireNoError(t, suite.Client.DB.Integration.DeleteOneID(installationA.ID).Exec(ctx))

	installationB, err := suite.Client.DB.Integration.Create().
		SetName("Remove Readd B").
		SetKind("removereaddb").
		SetDefinitionID("def_dirsynctest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: tenant}}).
		Save(ctx)
	th.RequireNoError(t, err)

	cleanupDirectoryPrefix(t, ctx, []string{installationB.ID}, prefix)

	counters, teardown := directoryEventCounters(t)
	defer teardown()

	changedExternalID := base.Accounts[0].ExternalID
	materialChange := base.withAccountMaterialChange(changedExternalID)

	newMembership := newDirectoryMembershipRecord(base.Accounts[1], base.Groups[1].ExternalID, base.Memberships[0].Role)
	materialChange.Memberships = append(materialChange.Memberships, newMembership)

	result := ingestDirectorySnapshotFixture(ctx, t, installationB, materialChange, true)
	waitForGala(t, counters.Runtime)

	assert.Check(t, is.Equal(0, result.Failed))
	assert.Check(t, is.Equal(2, result.Changed), "a reinstall carrying one material account change and one new membership must count as exactly two changed records")
	assert.Check(t, is.Equal(int64(0), counters.AccountCreates.Load()), "a same-tenant reinstall must not duplicate accounts")
	assert.Check(t, is.Equal(int64(0), counters.GroupCreates.Load()), "a same-tenant reinstall must not duplicate groups")
	assert.Check(t, is.Equal(int64(1), counters.MembershipCreates.Load()), "only the new membership may be created by a same-tenant reinstall")
	assert.Check(t, is.Equal(int64(1), counters.AccountUpdates.Load()), "the reinstall must emit exactly one account update event")
	assert.Check(t, is.Equal(int64(0), counters.GroupUpdates.Load()), "unchanged groups must not emit an update event")
	assert.Check(t, is.Equal(int64(0), counters.MembershipUpdates.Load()), "unchanged memberships must not emit an update event")

	for _, a := range base.Accounts {
		row := directoryAccountByExternalID(ctx, t, a.ExternalID)
		assert.Check(t, is.Equal(installationB.ID, row.IntegrationID), "every account must repoint to the adopting installation")
		assert.Check(t, is.Equal(installationB.ID, row.ManagedBy), "every account's managed_by must repoint to the adopting installation")
	}
	for _, g := range base.Groups {
		row := directoryGroupByExternalID(ctx, t, g.ExternalID)
		assert.Check(t, is.Equal(installationB.ID, row.IntegrationID), "every group must repoint to the adopting installation")
		assert.Check(t, is.Equal(installationB.ID, row.ManagedBy), "every group's managed_by must repoint to the adopting installation")
	}
	for _, m := range base.Memberships {
		row := directoryMembershipByExternalIDs(ctx, t, m.DirectoryAccountID, m.DirectoryGroupID)
		assert.Check(t, is.Equal(installationB.ID, row.IntegrationID), "every membership must repoint to the adopting installation")
		assert.Check(t, is.Equal(installationB.ID, row.ManagedBy), "every membership's managed_by must repoint to the adopting installation")
		assert.Check(t, row.RemovedAt == nil, "no membership must be removed by the reinstall's material-change ingest")
	}

	added := directoryMembershipByExternalIDs(ctx, t, newMembership.DirectoryAccountID, newMembership.DirectoryGroupID)
	assert.Check(t, is.Equal(installationB.ID, added.IntegrationID), "the new membership must link to the rows the old installation wrote")

	assert.Check(t, is.Equal(0, directoryRemovedMembershipCount(ctx, t, installationB.ID)), "no membership must be removal-inferred by the reinstall")

	after := directoryAccountByExternalID(ctx, t, changedExternalID)
	assert.Check(t, is.Equal(materialChange.account(changedExternalID).DisplayName, after.DisplayName))
}
