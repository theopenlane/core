//go:build test

package eventstest_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	openapi "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorymembership"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
)

// TestDirectoryIntegrationDeleteThenReinstallRelinks verifies a same-tenant reinstall converges onto surviving rows instead of duplicating them
func TestDirectoryIntegrationDeleteThenReinstallRelinks(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const prefix = "delreinstall"
	const tenant = "tenant-" + prefix

	installationA, err := suite.Client.DB.Integration.Create().
		SetName("Delete Reinstall Test A").
		SetKind("delreinstalla").
		SetDefinitionID("def_dirsynctest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: tenant}}).
		Save(ctx)
	th.RequireNoError(t, err)

	base := newDirectorySnapshot(prefix)

	seeded := ingestDirectorySnapshotFixture(ctx, t, installationA, base, true)
	waitForGala(t, suite.GalaRuntime)
	assert.Check(t, is.Equal(0, seeded.Failed))
	assert.Check(t, is.Equal(len(base.Accounts)+len(base.Groups)+len(base.Memberships), seeded.Changed), "the first sync must create every account, group, and membership")

	accountsBefore := lo.Map(base.Accounts, func(a directoryAccountRecord, _ int) *ent.DirectoryAccount {
		return directoryAccountByExternalID(ctx, t, a.ExternalID)
	})
	groupsBefore := lo.Map(base.Groups, func(g directoryGroupRecord, _ int) *ent.DirectoryGroup {
		return directoryGroupByExternalID(ctx, t, g.ExternalID)
	})
	membershipsBefore := lo.Map(base.Memberships, func(m directoryMembershipRecord, _ int) *ent.DirectoryMembership {
		return directoryMembershipByExternalIDs(ctx, t, m.DirectoryAccountID, m.DirectoryGroupID)
	})

	th.RequireNoError(t, suite.Client.DB.Integration.DeleteOneID(installationA.ID).Exec(ctx))

	for i, a := range base.Accounts {
		row := directoryAccountByExternalID(ctx, t, a.ExternalID)
		assert.Check(t, is.Equal(accountsBefore[i].ID, row.ID))
		assert.Check(t, is.Equal(installationA.ID, row.IntegrationID), "a soft-deleted installation must not lose attribution of its directory accounts")
	}
	for i, g := range base.Groups {
		row := directoryGroupByExternalID(ctx, t, g.ExternalID)
		assert.Check(t, is.Equal(groupsBefore[i].ID, row.ID))
		assert.Check(t, is.Equal(installationA.ID, row.IntegrationID), "a soft-deleted installation must not lose attribution of its directory groups")
	}
	for i, m := range base.Memberships {
		row := directoryMembershipByExternalIDs(ctx, t, m.DirectoryAccountID, m.DirectoryGroupID)
		assert.Check(t, is.Equal(membershipsBefore[i].ID, row.ID))
		assert.Check(t, is.Equal(installationA.ID, row.IntegrationID), "a soft-deleted installation must not lose attribution of its directory memberships")
		assert.Check(t, row.RemovedAt == nil, "a soft-deleted installation must not cause its memberships to be marked removed")
	}

	installationB, err := suite.Client.DB.Integration.Create().
		SetName("Delete Reinstall Test B").
		SetKind("delreinstallb").
		SetDefinitionID("def_dirsynctest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: tenant}}).
		Save(ctx)
	th.RequireNoError(t, err)

	cleanupDirectoryPrefix(t, ctx, []string{installationB.ID}, prefix)

	counters, teardown := directoryEventCounters(t)
	defer teardown()

	relinked := ingestDirectorySnapshotFixture(ctx, t, installationB, base.identical(), true)
	waitForGala(t, counters.Runtime)

	assert.Check(t, is.Equal(0, relinked.Failed))
	assert.Check(t, is.Equal(0, relinked.Changed), "a same-tenant reinstall confirming identical rows must not count as changed")
	assert.Check(t, is.Equal(int64(0), counters.AccountCreates.Load()), "a same-tenant reinstall must not duplicate accounts")
	assert.Check(t, is.Equal(int64(0), counters.GroupCreates.Load()), "a same-tenant reinstall must not duplicate groups")
	assert.Check(t, is.Equal(int64(0), counters.MembershipCreates.Load()), "a same-tenant reinstall must not duplicate memberships")
	assert.Check(t, is.Equal(int64(0), counters.AccountUpdates.Load()), "a same-tenant reinstall converging on unchanged rows must not emit an update event")
	assert.Check(t, is.Equal(int64(0), counters.GroupUpdates.Load()), "a same-tenant reinstall converging on unchanged rows must not emit an update event")
	assert.Check(t, is.Equal(int64(0), counters.MembershipUpdates.Load()), "a same-tenant reinstall converging on unchanged rows must not emit an update event")

	for i, a := range base.Accounts {
		row := directoryAccountByExternalID(ctx, t, a.ExternalID)
		assert.Check(t, is.Equal(accountsBefore[i].ID, row.ID), "relink must repoint the surviving row, not create a new one")
		assert.Check(t, is.Equal(installationB.ID, row.IntegrationID), "relink must repoint the account to the adopting installation")
		assert.Check(t, is.Equal(installationB.ID, row.ManagedBy), "relink must repoint the account's managed_by to the adopting installation")
	}
	for i, g := range base.Groups {
		row := directoryGroupByExternalID(ctx, t, g.ExternalID)
		assert.Check(t, is.Equal(groupsBefore[i].ID, row.ID), "relink must repoint the surviving row, not create a new one")
		assert.Check(t, is.Equal(installationB.ID, row.IntegrationID), "relink must repoint the group to the adopting installation")
		assert.Check(t, is.Equal(installationB.ID, row.ManagedBy), "relink must repoint the group's managed_by to the adopting installation")
	}
	for i, m := range base.Memberships {
		row := directoryMembershipByExternalIDs(ctx, t, m.DirectoryAccountID, m.DirectoryGroupID)
		assert.Check(t, is.Equal(membershipsBefore[i].ID, row.ID), "relink must repoint the surviving row, not create a new one")
		assert.Check(t, is.Equal(installationB.ID, row.IntegrationID), "relink must repoint the membership to the adopting installation")
		assert.Check(t, is.Equal(installationB.ID, row.ManagedBy), "relink must repoint the membership's managed_by to the adopting installation")
		assert.Check(t, row.RemovedAt == nil)
		assert.Check(t, is.Equal(tenant, row.SourceInstanceID))
	}

	changedExternalID := base.Accounts[0].ExternalID
	materialChange := base.withAccountMaterialChange(changedExternalID)

	thirdRun := ingestDirectorySnapshotFixture(ctx, t, installationB, materialChange, true)
	waitForGala(t, counters.Runtime)

	assert.Check(t, is.Equal(0, thirdRun.Failed))
	assert.Check(t, is.Equal(1, thirdRun.Changed), "a single material account change must count as exactly one changed record")
	assert.Check(t, is.Equal(int64(1), counters.AccountUpdates.Load()), "a single material account change must emit exactly one account update event")
}

// TestDirectoryDifferentDefinitionSameTenantIsReadOnly pins the intent that a sync run must be read-only for another definition's installation on the same tenant
func TestDirectoryDifferentDefinitionSameTenantIsReadOnly(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const prefix = "diffdefsametenant"
	const tenant = "tenant-" + prefix

	installationA, err := suite.Client.DB.Integration.Create().
		SetName("Different Definition Same Tenant A").
		SetKind("diffdefa").
		SetDefinitionID("def_dirsynctest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: tenant}}).
		Save(ctx)
	th.RequireNoError(t, err)

	installationC, err := suite.Client.DB.Integration.Create().
		SetName("Different Definition Same Tenant C").
		SetKind("diffdefc").
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

	changedExternalID := base.Accounts[0].ExternalID
	groupExternalID := base.Groups[0].ExternalID
	membership := base.Memberships[0]

	accountBefore := directoryAccountByExternalID(ctx, t, changedExternalID)

	counters, teardown := directoryEventCounters(t)
	defer teardown()

	materialChange := base.withAccountMaterialChange(changedExternalID)
	result := ingestDirectorySnapshotFixture(ctx, t, installationC, materialChange, true)
	waitForGala(t, counters.Runtime)

	assert.Check(t, is.Equal(0, result.Failed))
	assert.Check(t, is.Equal(int64(0), counters.AccountCreates.Load()), "a different definition on the same tenant must not create rows")
	assert.Check(t, is.Equal(int64(0), counters.GroupCreates.Load()), "a different definition on the same tenant must not create rows")
	assert.Check(t, is.Equal(int64(0), counters.MembershipCreates.Load()), "a different definition on the same tenant must not create rows")
	assert.Check(t, is.Equal(int64(0), counters.AccountUpdates.Load()), "a different definition must not be able to write another definition's rows")
	assert.Check(t, is.Equal(int64(0), counters.GroupUpdates.Load()), "a different definition must not be able to write another definition's rows")
	assert.Check(t, is.Equal(int64(0), counters.MembershipUpdates.Load()), "a different definition must not be able to write another definition's rows")

	accountAfter := directoryAccountByExternalID(ctx, t, changedExternalID)
	assert.Check(t, is.Equal(accountBefore.DisplayName, accountAfter.DisplayName), "a different definition's material change must not be applied to another definition's row")
	assert.Check(t, is.Equal(installationA.ID, accountAfter.IntegrationID), "a different definition's confirming sync must not repoint another definition's account")

	groupAfter := directoryGroupByExternalID(ctx, t, groupExternalID)
	assert.Check(t, is.Equal(installationA.ID, groupAfter.IntegrationID), "a different definition's confirming sync must not repoint another definition's group")

	membershipAfter := directoryMembershipByExternalIDs(ctx, t, membership.DirectoryAccountID, membership.DirectoryGroupID)
	assert.Check(t, is.Equal(installationA.ID, membershipAfter.IntegrationID), "a different definition's confirming sync must not repoint another definition's membership")
}

// TestDirectoryDifferentDefinitionDifferentTenantIsolated verifies two installations on distinct source_instance_id tenants never see each other's directory rows
func TestDirectoryDifferentDefinitionDifferentTenantIsolated(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const prefixA = "isotena"
	const prefixC = "isotenc"

	installationA, err := suite.Client.DB.Integration.Create().
		SetName("Isolated Tenant A").
		SetKind("isotenanta").
		SetDefinitionID("def_dirsynctest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-" + prefixA}}).
		Save(ctx)
	th.RequireNoError(t, err)
	cleanupDirectoryPrefix(t, ctx, []string{installationA.ID}, prefixA)

	installationC, err := suite.Client.DB.Integration.Create().
		SetName("Isolated Tenant C").
		SetKind("isotenantc").
		SetDefinitionID("def_dirsynctest_other").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-" + prefixC}}).
		Save(ctx)
	th.RequireNoError(t, err)
	cleanupDirectoryPrefix(t, ctx, []string{installationC.ID}, prefixC)

	counters, teardown := directoryEventCounters(t)
	defer teardown()

	baseA := newDirectorySnapshot(prefixA)
	baseC := newDirectorySnapshot(prefixC)

	resultA := ingestDirectorySnapshotFixture(ctx, t, installationA, baseA, true)
	waitForGala(t, counters.Runtime)
	assert.Check(t, is.Equal(0, resultA.Failed))
	assert.Check(t, is.Equal(len(baseA.Accounts)+len(baseA.Groups)+len(baseA.Memberships), resultA.Changed), "A's installation must create its own rows")

	resultC := ingestDirectorySnapshotFixture(ctx, t, installationC, baseC, true)
	waitForGala(t, counters.Runtime)
	assert.Check(t, is.Equal(0, resultC.Failed))
	assert.Check(t, is.Equal(len(baseC.Accounts)+len(baseC.Groups)+len(baseC.Memberships), resultC.Changed), "C's installation must create its own rows, isolated from A's tenant")

	assert.Check(t, is.Equal(int64(len(baseA.Accounts)+len(baseC.Accounts)), counters.AccountCreates.Load()))
	assert.Check(t, is.Equal(int64(len(baseA.Groups)+len(baseC.Groups)), counters.GroupCreates.Load()))
	assert.Check(t, is.Equal(int64(len(baseA.Memberships)+len(baseC.Memberships)), counters.MembershipCreates.Load()))

	aAccountExternalID := baseA.Accounts[0].ExternalID
	aBefore := directoryAccountByExternalID(ctx, t, aAccountExternalID)

	changedExternalID := baseC.Accounts[0].ExternalID
	materialChange := baseC.withAccountMaterialChange(changedExternalID)

	thirdResult := ingestDirectorySnapshotFixture(ctx, t, installationC, materialChange, true)
	waitForGala(t, counters.Runtime)

	assert.Check(t, is.Equal(0, thirdResult.Failed))
	assert.Check(t, is.Equal(1, thirdResult.Changed), "a material change on C's own tenant must count as exactly one changed record")
	assert.Check(t, is.Equal(int64(1), counters.AccountUpdates.Load()), "the material change must emit exactly one account update event, scoped to C's tenant")
	assert.Check(t, is.Equal(int64(0), counters.GroupUpdates.Load()))
	assert.Check(t, is.Equal(int64(0), counters.MembershipUpdates.Load()))

	cAfter := directoryAccountByExternalID(ctx, t, changedExternalID)
	assert.Check(t, is.Equal(materialChange.account(changedExternalID).DisplayName, cAfter.DisplayName))
	assert.Check(t, is.Equal(installationC.ID, cAfter.IntegrationID))

	aAfter := directoryAccountByExternalID(ctx, t, aAccountExternalID)
	assert.Check(t, is.Equal(aBefore.DisplayName, aAfter.DisplayName), "a material change on C's tenant must not affect A's isolated tenant row")
	assert.Check(t, is.Equal(installationA.ID, aAfter.IntegrationID))
	assert.Check(t, aAfter.UpdatedAt.Equal(aBefore.UpdatedAt), "A's row must not be touched by C's ingest")
}

// TestDirectoryMembershipRemovalInference walks one installation through the full membership removal-inference lifecycle
func TestDirectoryMembershipRemovalInference(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const prefix = "removalinference"
	const tenant = "tenant-" + prefix

	installation, err := suite.Client.DB.Integration.Create().
		SetName("Membership Removal Inference Test").
		SetKind("removalinference").
		SetDefinitionID("def_dirsynctest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: tenant}}).
		Save(ctx)
	th.RequireNoError(t, err)

	cleanupDirectoryPrefix(t, ctx, []string{installation.ID}, prefix)

	// snapshot removal adds the ingesting run to the removed row's integration_runs edge, which
	// requires a real IntegrationRun id; production removal always runs under a run
	run, err := suite.Client.DB.IntegrationRun.Create().
		SetIntegrationID(installation.ID).
		SetOwnerID(installation.OwnerID).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		(&th.Cleanup[*ent.IntegrationRunDeleteOne]{Client: suite.Client.DB.IntegrationRun, ID: run.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	counters, teardown := directoryEventCounters(t)
	defer teardown()

	base := newDirectorySnapshot(prefix)
	removedInB := base.Memberships[0]
	omittedInC := base.Memberships[1]
	omittedInF := base.Memberships[3]

	t.Run("a base snapshot activates every membership under the first run", func(t *testing.T) {
		result := ingestDirectorySnapshotFixture(ctx, t, installation, base, true)
		waitForGala(t, counters.Runtime)
		assert.Check(t, is.Equal(0, result.Failed))

		for _, m := range base.Memberships {
			row := directoryMembershipByExternalIDs(ctx, t, m.DirectoryAccountID, m.DirectoryGroupID)
			assert.Check(t, row.RemovedAt == nil)
		}
	})

	t.Run("b a complete snapshot without one membership marks only that membership removed", func(t *testing.T) {
		withoutB := base.withoutMembership(removedInB.DirectoryAccountID, removedInB.DirectoryGroupID)

		result := ingestDirectorySnapshotFixture(ctx, t, installation, withoutB, true, operations.IngestOptions{RunID: run.ID})
		waitForGala(t, counters.Runtime)
		assert.Check(t, is.Equal(0, result.Failed))
		assert.Check(t, is.Equal(1, result.Removed), "the inferred removal must count as exactly one removed record")
		assert.Check(t, is.Equal(0, result.Changed), "snapshot removal is tallied separately from per-record Changed")

		removed := directoryRemovedMembershipByExternalIDs(ctx, t, removedInB.DirectoryAccountID, removedInB.DirectoryGroupID)
		assert.Check(t, removed.RemovedAt != nil)

		for _, m := range withoutB.Memberships {
			row := directoryMembershipByExternalIDs(ctx, t, m.DirectoryAccountID, m.DirectoryGroupID)
			assert.Check(t, row.RemovedAt == nil)
		}
	})

	t.Run("c an incomplete snapshot without another membership marks nothing new removed", func(t *testing.T) {
		withoutBAndC := base.
			withoutMembership(removedInB.DirectoryAccountID, removedInB.DirectoryGroupID).
			withoutMembership(omittedInC.DirectoryAccountID, omittedInC.DirectoryGroupID)

		result := ingestDirectorySnapshotFixture(ctx, t, installation, withoutBAndC, false)
		waitForGala(t, counters.Runtime)
		assert.Check(t, is.Equal(0, result.Failed))

		stillActive := directoryMembershipByExternalIDs(ctx, t, omittedInC.DirectoryAccountID, omittedInC.DirectoryGroupID)
		assert.Check(t, stillActive.RemovedAt == nil, "a partial snapshot must never authorize removal")
		assert.Check(t, is.Equal(1, directoryRemovedMembershipCount(ctx, t, installation.ID)), "only the membership removed in step b must carry removed_at")
	})

	t.Run("d re-adding a removed membership creates a new episode and keeps the old row removed", func(t *testing.T) {
		result := ingestDirectorySnapshotFixture(ctx, t, installation, base.identical(), true)
		waitForGala(t, counters.Runtime)
		assert.Check(t, is.Equal(0, result.Failed))
		assert.Check(t, is.Equal(1, result.Changed), "re-adding the removed membership must count as exactly one changed record")

		reAdded := directoryMembershipByExternalIDs(ctx, t, removedInB.DirectoryAccountID, removedInB.DirectoryGroupID)
		assert.Check(t, reAdded.RemovedAt == nil)

		oldRow := directoryRemovedMembershipByExternalIDs(ctx, t, removedInB.DirectoryAccountID, removedInB.DirectoryGroupID)
		assert.Check(t, oldRow.RemovedAt != nil)
		assert.Check(t, reAdded.ID != oldRow.ID, "the re-added membership must be a new row, not the removed episode revived")
	})

	t.Run("f a failed record vetoes removal inference for an otherwise complete snapshot", func(t *testing.T) {
		withBogusRecord := base.withoutMembership(omittedInF.DirectoryAccountID, omittedInF.DirectoryGroupID)
		withBogusRecord.Memberships = append(withBogusRecord.Memberships, directoryMembershipRecord{
			DirectoryAccountID: prefix + "-acct-missing",
			DirectoryGroupID:   base.Groups[0].ExternalID,
			Role:               enums.DirectoryMembershipRoleMember.String(),
		})

		beforeRemoved := directoryRemovedMembershipCount(ctx, t, installation.ID)

		result := ingestDirectorySnapshotFixture(ctx, t, installation, withBogusRecord, true)
		waitForGala(t, counters.Runtime)
		assert.Check(t, is.Equal(1, result.Failed), "the bogus membership reference must fail exactly one record")

		afterRemoved := directoryRemovedMembershipCount(ctx, t, installation.ID)
		assert.Check(t, is.Equal(beforeRemoved, afterRemoved), "a failed record must veto removal inference even for an otherwise complete snapshot")

		stillActive := directoryMembershipByExternalIDs(ctx, t, omittedInF.DirectoryAccountID, omittedInF.DirectoryGroupID)
		assert.Check(t, stillActive.RemovedAt == nil, "the omitted membership must not be marked removed when the run also had a failed record")
	})
}

// TestDirectoryAccountDisappearanceIsRemovalResurrectedOnReAppearance verifies a directory account
// absent from a complete snapshot is removal-inferred the same as every other ingest schema, and
// resurrected in place — the same row, not a new one — once it reappears in a later complete snapshot
func TestDirectoryAccountDisappearanceIsRemovalResurrectedOnReAppearance(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const prefix = "acctdisappear"
	const tenant = "tenant-" + prefix

	installation, err := suite.Client.DB.Integration.Create().
		SetName("Account Disappearance Test").
		SetKind("acctdisappear").
		SetDefinitionID("def_dirsynctest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: tenant}}).
		Save(ctx)
	th.RequireNoError(t, err)

	cleanupDirectoryPrefix(t, ctx, []string{installation.ID}, prefix)

	// snapshot removal adds the ingesting run to the removed row's integration_runs edge, which
	// requires a real IntegrationRun id; production removal always runs under a run
	run, err := suite.Client.DB.IntegrationRun.Create().
		SetIntegrationID(installation.ID).
		SetOwnerID(installation.OwnerID).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		(&th.Cleanup[*ent.IntegrationRunDeleteOne]{Client: suite.Client.DB.IntegrationRun, ID: run.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	base := newDirectorySnapshot(prefix)
	disappearing := base.Accounts[0]

	seeded := ingestDirectorySnapshotFixture(ctx, t, installation, base, true)
	waitForGala(t, suite.GalaRuntime)
	assert.Check(t, is.Equal(0, seeded.Failed))

	before := directoryAccountByExternalID(ctx, t, disappearing.ExternalID)

	withoutAccount := base.withoutAccount(disappearing.ExternalID)

	result := ingestDirectorySnapshotFixture(ctx, t, installation, withoutAccount, true, operations.IngestOptions{RunID: run.ID})
	waitForGala(t, suite.GalaRuntime)
	assert.Check(t, is.Equal(0, result.Failed))

	removed := directoryAccountByExternalID(ctx, t, disappearing.ExternalID)
	assert.Check(t, removed.RemovedAt != nil, "an account absent from a complete snapshot must be removal-inferred like every other ingest schema")
	assert.Check(t, is.Equal(before.ID, removed.ID), "removal must mark the existing row, not create a new one")

	for _, m := range base.Memberships {
		if m.DirectoryAccountID != disappearing.ExternalID {
			continue
		}

		membership := directoryRemovedMembershipByExternalIDs(ctx, t, m.DirectoryAccountID, m.DirectoryGroupID)
		assert.Check(t, membership.RemovedAt != nil, "a membership referencing a disappeared account must still be removal-inferred")
	}

	reappeared := ingestDirectorySnapshotFixture(ctx, t, installation, base.identical(), true)
	waitForGala(t, suite.GalaRuntime)
	assert.Check(t, is.Equal(0, reappeared.Failed))

	resurrected := directoryAccountByExternalID(ctx, t, disappearing.ExternalID)
	assert.Check(t, resurrected.RemovedAt == nil, "an account reappearing in a later complete snapshot must be resurrected")
	assert.Check(t, is.Equal(before.ID, resurrected.ID), "resurrection must reuse the existing row, not create a new one")
}

// directoryRemovedMembershipByExternalIDs loads the removed membership episode between the accounts and group identified by accountExternalID and groupExternalID
func directoryRemovedMembershipByExternalIDs(ctx context.Context, t *testing.T, accountExternalID, groupExternalID string) *ent.DirectoryMembership {
	t.Helper()

	account := directoryAccountByExternalID(ctx, t, accountExternalID)
	group := directoryGroupByExternalID(ctx, t, groupExternalID)

	membership, err := suite.Client.DB.DirectoryMembership.Query().
		Where(directorymembership.DirectoryAccountID(account.ID), directorymembership.DirectoryGroupID(group.ID), directorymembership.RemovedAtNotNil()).
		Only(ctx)
	th.RequireNoError(t, err)

	return membership
}

// directoryRemovedMembershipCount counts the installation's currently removed directory memberships
func directoryRemovedMembershipCount(ctx context.Context, t *testing.T, integrationID string) int {
	t.Helper()

	count, err := suite.Client.DB.DirectoryMembership.Query().
		Where(directorymembership.IntegrationID(integrationID), directorymembership.RemovedAtNotNil()).
		Count(ctx)
	th.RequireNoError(t, err)

	return count
}
