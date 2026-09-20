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

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorymembership"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
)

// directoryMembershipByExternalIDs loads the active membership between the accounts and group identified by accountExternalID and groupExternalID
func directoryMembershipByExternalIDs(ctx context.Context, t *testing.T, accountExternalID, groupExternalID string) *ent.DirectoryMembership {
	t.Helper()

	account := directoryAccountByExternalID(ctx, t, accountExternalID)
	group := directoryGroupByExternalID(ctx, t, groupExternalID)

	membership, err := suite.Client.DB.DirectoryMembership.Query().
		Where(directorymembership.DirectoryAccountID(account.ID), directorymembership.DirectoryGroupID(group.ID), directorymembership.RemovedAtIsNil()).
		Only(ctx)
	th.RequireNoError(t, err)

	return membership
}

// TestDirectoryFullSnapshotIdleResyncWritesNothing verifies an unchanged resync emits zero update events, and one material change produces exactly one
func TestDirectoryFullSnapshotIdleResyncWritesNothing(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const prefix = "fullidle"

	integration, err := suite.Client.DB.Integration.Create().
		SetName("Full Snapshot Idle Resync Test").
		SetKind("fullidleresync").
		SetDefinitionID("def_dirsynctest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-fullidleresync"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	cleanupDirectoryPrefix(t, ctx, []string{integration.ID}, prefix)

	counters, teardown := directoryEventCounters(t)
	defer teardown()

	base := newDirectorySnapshot(prefix)

	seeded := ingestDirectorySnapshotFixture(ctx, t, integration, base, true)
	waitForGala(t, counters.Runtime)

	assert.Check(t, is.Equal(0, seeded.Failed))
	assert.Check(t, is.Equal(len(base.Accounts)+len(base.Groups)+len(base.Memberships), seeded.Changed), "the first sync must create every account, group, and membership")

	groupsBefore := lo.Map(base.Groups, func(g directoryGroupRecord, _ int) *ent.DirectoryGroup {
		return directoryGroupByExternalID(ctx, t, g.ExternalID)
	})

	resynced := ingestDirectorySnapshotFixture(ctx, t, integration, base.identical(), true)
	waitForGala(t, counters.Runtime)

	assert.Check(t, is.Equal(0, resynced.Failed))
	assert.Check(t, is.Equal(0, resynced.Changed), "a fully idle resync must not count any record as changed")
	assert.Check(t, is.Equal(int64(0), counters.AccountUpdates.Load()), "a fully idle resync must not emit an account update mutation event")
	assert.Check(t, is.Equal(int64(0), counters.GroupUpdates.Load()), "a fully idle resync must not emit a group update mutation event")
	assert.Check(t, is.Equal(int64(0), counters.MembershipUpdates.Load()), "a fully idle resync must not emit a membership update mutation event")

	for i, g := range base.Groups {
		after := directoryGroupByExternalID(ctx, t, g.ExternalID)
		assert.Check(t, is.Equal(groupsBefore[i].DisplayName, after.DisplayName), "an idle resync must not change an unchanged group row")
	}

	changedExternalID := base.Accounts[0].ExternalID
	materialChange := base.withAccountMaterialChange(changedExternalID)

	thirdRun := ingestDirectorySnapshotFixture(ctx, t, integration, materialChange, true)
	waitForGala(t, counters.Runtime)

	assert.Check(t, is.Equal(0, thirdRun.Failed))
	assert.Check(t, is.Equal(1, thirdRun.Changed), "a single material account change must count as exactly one changed record")
	assert.Check(t, is.Equal(int64(1), counters.AccountUpdates.Load()), "a single material account change must emit exactly one account update mutation event")

	after := directoryAccountByExternalID(ctx, t, changedExternalID)
	assert.Check(t, is.Equal(materialChange.account(changedExternalID).DisplayName, after.DisplayName))
}

// TestDirectoryGroupProfileChurnRidesAlongMaterialChange verifies a change confined to the group profile bag is not written, counted, or emitted on its own, and lands once a display-name change follows
func TestDirectoryGroupProfileChurnRidesAlongMaterialChange(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const prefix = "grpchurn"

	integration, err := suite.Client.DB.Integration.Create().
		SetName("Group Profile Churn Test").
		SetKind("grpprofilechurn").
		SetDefinitionID("def_dirsynctest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-grpprofilechurn"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	cleanupDirectoryPrefix(t, ctx, []string{integration.ID}, prefix)

	counters, teardown := directoryEventCounters(t)
	defer teardown()

	base := newDirectorySnapshot(prefix)
	groupExternalID := base.Groups[0].ExternalID

	seeded := ingestDirectorySnapshotFixture(ctx, t, integration, base, true)
	waitForGala(t, counters.Runtime)
	assert.Check(t, is.Equal(0, seeded.Failed))

	before := directoryGroupByExternalID(ctx, t, groupExternalID)

	churned := base.withGroupProfileChurn(groupExternalID)
	churnResult := ingestDirectorySnapshotFixture(ctx, t, integration, churned, true)
	waitForGala(t, counters.Runtime)

	assert.Check(t, is.Equal(0, churnResult.Failed))
	assert.Check(t, is.Equal(0, churnResult.Changed), "a group profile-only change must not count as changed")
	assert.Check(t, is.Equal(int64(0), counters.GroupUpdates.Load()), "a group profile-only change must not emit an update mutation event")

	afterChurn := directoryGroupByExternalID(ctx, t, groupExternalID)
	assert.Check(t, is.Equal(before.DisplayName, afterChurn.DisplayName), "a profile-only change must not touch other fields")
	assert.Check(t, is.Equal(before.Profile["lastLoginTime"], afterChurn.Profile["lastLoginTime"]), "a profile-only change must not be written")
	assert.Check(t, afterChurn.UpdatedAt.Equal(before.UpdatedAt), "a profile-only change must not rewrite the row")

	materialized := churned.withGroupMaterialChange(groupExternalID)
	materialResult := ingestDirectorySnapshotFixture(ctx, t, integration, materialized, true)
	waitForGala(t, counters.Runtime)

	assert.Check(t, is.Equal(0, materialResult.Failed))
	assert.Check(t, is.Equal(1, materialResult.Changed), "a display-name change must count as changed")
	assert.Check(t, is.Equal(int64(1), counters.GroupUpdates.Load()), "a display-name change must emit exactly one update mutation event")

	after := directoryGroupByExternalID(ctx, t, groupExternalID)
	assert.Check(t, is.Equal(materialized.group(groupExternalID).DisplayName, after.DisplayName))
	assert.Check(t, !after.UpdatedAt.Equal(afterChurn.UpdatedAt), "a material group change must rewrite the row")
	assert.Check(t, after.Profile["lastLoginTime"] == directoryProfileChurnedLastLogin, "the profile must ride along the material change")
}

// TestDirectoryMembershipMetadataChangePersisted verifies metadata persistence follows its volatility
// descriptor and that a material role change persists the latest metadata in either configuration.
func TestDirectoryMembershipMetadataChangePersisted(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	field, ok := entityops.SchemaDirectoryMembership.FieldByName(directorymembership.FieldMetadata)
	assert.Assert(t, ok, "membership metadata must have a field descriptor")

	const prefix = "memchurn"

	integration, err := suite.Client.DB.Integration.Create().
		SetName("Membership Metadata Churn Test").
		SetKind("memmetadatachurn").
		SetDefinitionID("def_dirsynctest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-memmetadatachurn"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	cleanupDirectoryPrefix(t, ctx, []string{integration.ID}, prefix)

	counters, teardown := directoryEventCounters(t)
	defer teardown()

	base := newDirectorySnapshot(prefix)
	membership := base.Memberships[0]

	seeded := ingestDirectorySnapshotFixture(ctx, t, integration, base, true)
	waitForGala(t, counters.Runtime)
	assert.Check(t, is.Equal(0, seeded.Failed))

	before := directoryMembershipByExternalIDs(ctx, t, membership.DirectoryAccountID, membership.DirectoryGroupID)

	churned := base.withMembershipMetadataChurn(membership.DirectoryAccountID, membership.DirectoryGroupID)
	result := ingestDirectorySnapshotFixture(ctx, t, integration, churned, true)
	waitForGala(t, counters.Runtime)

	assert.Check(t, is.Equal(0, result.Failed))

	afterOperation := directoryMembershipByExternalIDs(ctx, t, membership.DirectoryAccountID, membership.DirectoryGroupID)

	assert.Check(t, is.DeepEqual(before.Metadata, membership.Metadata), "the seeded metadata is the pre-change value")

	expectedResults := []int64{1, 1}

	if field.Volatile {
		expectedResults = []int64{0, 0}
		assert.Check(t, afterOperation.UpdatedAt.Equal(before.UpdatedAt), "a volatile metadata-only change must not rewrite the row")
	}

	receivedResults := []int64{int64(result.Changed), counters.MembershipUpdates.Load()}
	assert.Check(t, is.DeepEqual(expectedResults, receivedResults), "changed records and update events must respect metadata volatility")

	materialized := churned.clone()
	materialized.Memberships[0].Role = enums.DirectoryMembershipRoleOwner.String()
	updatesBefore := counters.MembershipUpdates.Load()

	materialResult := ingestDirectorySnapshotFixture(ctx, t, integration, materialized, true)
	waitForGala(t, counters.Runtime)

	assert.Check(t, is.Equal(0, materialResult.Failed))
	assert.Check(t, is.Equal(1, materialResult.Changed), "a role change must count as changed")
	assert.Check(t, is.Equal(int64(1), counters.MembershipUpdates.Load()-updatesBefore), "a role change must emit exactly one update mutation event")

	after := directoryMembershipByExternalIDs(ctx, t, membership.DirectoryAccountID, membership.DirectoryGroupID)
	assert.Check(t, is.Equal(enums.DirectoryMembershipRoleOwner, after.Role))
	assert.Check(t, !after.UpdatedAt.Equal(afterOperation.UpdatedAt), "a material membership change must rewrite the row")
	assert.Check(t, is.DeepEqual(materialized.Memberships[0].Metadata, after.Metadata), "the metadata must be with the material change")
}
