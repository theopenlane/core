//go:build test

package eventstest_test

import (
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	openapi "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
)

// TestDirectoryGroupDisappearanceIsRemovalRevivedOnReAppearance verifies a directory group absent
// from a complete snapshot is removal-inferred together with every membership referencing it, every
// other group and membership stays active, and a later complete snapshot revives the group on the
// same row while its memberships come back as new rows
func TestDirectoryGroupDisappearanceIsRemovalRevivedOnReAppearance(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const prefix = "grpdisappear"
	const tenant = "tenant-" + prefix

	installation, err := suite.Client.DB.Integration.Create().
		SetName("Group Disappearance Test").
		SetKind("grpdisappear").
		SetDefinitionID("def_dirsynctest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: tenant}}).
		Save(ctx)
	th.RequireNoError(t, err)

	cleanupDirectoryPrefix(t, ctx, []string{installation.ID}, prefix)

	run, err := suite.Client.DB.IntegrationRun.Create().
		SetIntegrationID(installation.ID).
		SetOwnerID(installation.OwnerID).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		(&th.Cleanup[*ent.IntegrationRunDeleteOne]{Client: suite.Client.DB.IntegrationRun, ID: run.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	base := newDirectorySnapshot(prefix)
	disappearing := base.Groups[0]

	omittedMemberships := lo.Filter(base.Memberships, func(m directoryMembershipRecord, _ int) bool {
		return m.DirectoryGroupID == disappearing.ExternalID
	})

	seeded := ingestDirectorySnapshotFixture(ctx, t, installation, base, true)
	waitForGala(t, suite.GalaRuntime)
	assert.Check(t, is.Equal(0, seeded.Failed))

	before := directoryGroupByExternalID(ctx, t, disappearing.ExternalID)

	withoutGroup := base.withoutGroup(disappearing.ExternalID)

	result := ingestDirectorySnapshotFixture(ctx, t, installation, withoutGroup, true, operations.IngestOptions{RunID: run.ID})
	waitForGala(t, suite.GalaRuntime)
	assert.Check(t, is.Equal(0, result.Failed))
	assert.Check(t, is.Equal(1+len(omittedMemberships), result.Removed), "the omitted group and every membership referencing it must each count as one removed record")

	removed := directoryGroupByExternalID(ctx, t, disappearing.ExternalID)
	assert.Check(t, removed.RemovedAt != nil, "a group absent from a complete snapshot must be removal-inferred like every other ingest schema")
	assert.Check(t, is.Equal(before.ID, removed.ID), "removal must mark the existing row, not create a new one")

	for _, m := range omittedMemberships {
		membership := directoryRemovedMembershipByExternalIDs(ctx, t, m.DirectoryAccountID, m.DirectoryGroupID)
		assert.Check(t, membership.RemovedAt != nil, "a membership referencing a disappeared group must be removal-inferred")
	}

	for _, g := range withoutGroup.Groups {
		row := directoryGroupByExternalID(ctx, t, g.ExternalID)
		assert.Check(t, row.RemovedAt == nil, "a group still present in the complete snapshot must stay active")
	}

	for _, m := range withoutGroup.Memberships {
		row := directoryMembershipByExternalIDs(ctx, t, m.DirectoryAccountID, m.DirectoryGroupID)
		assert.Check(t, row.RemovedAt == nil, "a membership still present in the complete snapshot must stay active")
	}

	reappeared := ingestDirectorySnapshotFixture(ctx, t, installation, base.identical(), true)
	waitForGala(t, suite.GalaRuntime)
	assert.Check(t, is.Equal(0, reappeared.Failed))

	revived := directoryGroupByExternalID(ctx, t, disappearing.ExternalID)
	assert.Check(t, revived.RemovedAt == nil, "a group reappearing in a later complete snapshot must be revived")
	assert.Check(t, is.Equal(before.ID, revived.ID), "revival must reuse the existing row, not create a new one")

	for _, m := range omittedMemberships {
		reAdded := directoryMembershipByExternalIDs(ctx, t, m.DirectoryAccountID, m.DirectoryGroupID)
		assert.Check(t, reAdded.RemovedAt == nil)

		oldRow := directoryRemovedMembershipByExternalIDs(ctx, t, m.DirectoryAccountID, m.DirectoryGroupID)
		assert.Check(t, oldRow.RemovedAt != nil)
		assert.Check(t, reAdded.ID != oldRow.ID, "the re-added membership must be a new row, not the removed episode revived")
	}
}
