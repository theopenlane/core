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

	accountsBefore := lo.Map(base.Accounts, func(a directoryAccountRecord, _ int) *ent.DirectoryAccount {
		return directoryAccountByExternalID(ctx, t, a.ExternalID)
	})
	groupsBefore := lo.Map(base.Groups, func(g directoryGroupRecord, _ int) *ent.DirectoryGroup {
		return directoryGroupByExternalID(ctx, t, g.ExternalID)
	})
	membershipsBefore := lo.Map(base.Memberships, func(m directoryMembershipRecord, _ int) *ent.DirectoryMembership {
		return directoryMembershipByExternalIDs(ctx, t, m.DirectoryAccountID, m.DirectoryGroupID)
	})

	for _, account := range accountsBefore {
		assert.Check(t, account.LastSeenAt == nil, "an account must not carry last_seen_at until a confirming sync follows its create")
	}

	resynced := ingestDirectorySnapshotFixture(ctx, t, integration, base.identical(), true)
	waitForGala(t, counters.Runtime)

	assert.Check(t, is.Equal(0, resynced.Failed))
	assert.Check(t, is.Equal(0, resynced.Changed), "a fully idle resync must not count any record as changed")
	assert.Check(t, is.Equal(int64(0), counters.AccountUpdates.Load()), "a fully idle resync must not emit an account update mutation event")
	assert.Check(t, is.Equal(int64(0), counters.GroupUpdates.Load()), "a fully idle resync must not emit a group update mutation event")
	assert.Check(t, is.Equal(int64(0), counters.MembershipUpdates.Load()), "a fully idle resync must not emit a membership update mutation event")

	for _, a := range base.Accounts {
		after := directoryAccountByExternalID(ctx, t, a.ExternalID)
		assert.Check(t, after.LastSeenAt != nil, "an idle resync must still advance last_seen_at on every account")
	}

	for i, g := range base.Groups {
		after := directoryGroupByExternalID(ctx, t, g.ExternalID)
		assert.Check(t, is.Equal(groupsBefore[i].DisplayName, after.DisplayName), "an idle resync must not change an unchanged group row")
		assert.Check(t, after.LastSeenAt != nil, "an idle resync must still advance last_seen_at on every group")
	}

	for i, m := range base.Memberships {
		after := directoryMembershipByExternalIDs(ctx, t, m.DirectoryAccountID, m.DirectoryGroupID)
		assert.Check(t, after.LastSeenAt.After(lo.FromPtr(membershipsBefore[i].LastSeenAt)), "an idle resync must still advance last_seen_at on every membership")
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

// TestDirectoryGroupProfileChangePersisted verifies a change confined to the group profile bag is a material update, and a later display-name change is one more
func TestDirectoryGroupProfileChangePersisted(t *testing.T) {
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
	assert.Check(t, is.Equal(1, churnResult.Changed), "a group profile change must count as changed")
	assert.Check(t, is.Equal(int64(1), counters.GroupUpdates.Load()), "a group profile change must emit exactly one update mutation event")

	afterChurn := directoryGroupByExternalID(ctx, t, groupExternalID)
	assert.Check(t, is.Equal(before.DisplayName, afterChurn.DisplayName), "a profile-only change must not touch other fields")
	assert.Check(t, afterChurn.Profile["lastLoginTime"] == directoryProfileChurnedLastLogin, "the stored profile must carry the new value")

	materialized := churned.withGroupMaterialChange(groupExternalID)
	materialResult := ingestDirectorySnapshotFixture(ctx, t, integration, materialized, true)
	waitForGala(t, counters.Runtime)

	assert.Check(t, is.Equal(0, materialResult.Failed))
	assert.Check(t, is.Equal(1, materialResult.Changed), "a display-name change must count as changed")
	assert.Check(t, is.Equal(int64(2), counters.GroupUpdates.Load()), "a display-name change must emit one more update mutation event")

	after := directoryGroupByExternalID(ctx, t, groupExternalID)
	assert.Check(t, is.Equal(materialized.group(groupExternalID).DisplayName, after.DisplayName))
	assert.Check(t, !after.UpdatedAt.Equal(afterChurn.UpdatedAt), "a material group change must rewrite the row")
	assert.Check(t, after.Profile["lastLoginTime"] == directoryProfileChurnedLastLogin, "the profile value written by the earlier change must survive")
}

// TestDirectoryMembershipMetadataChangePersisted verifies a change confined to membership metadata is a material update that is written and emitted
func TestDirectoryMembershipMetadataChangePersisted(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

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
	assert.Check(t, is.Equal(1, result.Changed), "a membership metadata change must count as changed")
	assert.Check(t, is.Equal(int64(1), counters.MembershipUpdates.Load()), "a membership metadata change must emit exactly one update mutation event")

	after := directoryMembershipByExternalIDs(ctx, t, membership.DirectoryAccountID, membership.DirectoryGroupID)

	assert.Check(t, is.DeepEqual(before.Metadata, membership.Metadata), "the seeded metadata is the pre-change value")
	assert.Check(t, is.DeepEqual(churned.Memberships[0].Metadata, after.Metadata), "the stored metadata must carry the new value")
}
