package graphapi_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/internal/ent/generated/directorymembership"
	"github.com/theopenlane/core/internal/httpserve/serveropts"
)

// TestBackfillDirectoryDuplicates seeds forked directory rows the way racing syncs stored them -
// same (integration_id, external_id) under different sync runs - and proves the backfill keeps
// the earliest row, removes the forks and their memberships, and leaves the survivor's
// membership intact
func TestBackfillDirectoryDuplicates(t *testing.T) {
	orgUser := suite.userBuilder(context.Background(), t)
	ctx := setContext(orgUser.UserCtx, suite.client.db)

	integration, err := suite.client.db.Integration.Create().
		SetName("Dup Backfill Test").
		SetKind("googleworkspace").
		SetOwnerID(orgUser.OrganizationID).
		Save(ctx)
	assert.NilError(t, err)

	runA, err := suite.client.db.DirectorySyncRun.Create().
		SetIntegrationID(integration.ID).
		SetOwnerID(orgUser.OrganizationID).
		SetStatus(enums.DirectorySyncRunStatusCompleted).
		Save(ctx)
	assert.NilError(t, err)

	runB, err := suite.client.db.DirectorySyncRun.Create().
		SetIntegrationID(integration.ID).
		SetOwnerID(orgUser.OrganizationID).
		SetStatus(enums.DirectorySyncRunStatusCompleted).
		Save(ctx)
	assert.NilError(t, err)

	newGroup := func(runID string) *generated.DirectoryGroup {
		group, err := suite.client.db.DirectoryGroup.Create().
			SetExternalID("grp-ext-1").
			SetDisplayName("engineering").
			SetOwnerID(orgUser.OrganizationID).
			SetIntegrationID(integration.ID).
			SetDirectorySyncRunID(runID).
			Save(ctx)
		assert.NilError(t, err)

		return group
	}

	survivorGroup := newGroup(runA.ID)
	forkedGroup := newGroup(runB.ID)

	newAccount := func(runID, displayName string) *generated.DirectoryAccount {
		account, err := suite.client.db.DirectoryAccount.Create().
			SetExternalID("acct-ext-1").
			SetDisplayName(displayName).
			SetOwnerID(orgUser.OrganizationID).
			SetIntegrationID(integration.ID).
			SetDirectorySyncRunID(runID).
			Save(ctx)
		assert.NilError(t, err)

		return account
	}

	survivorAccount := newAccount(runA.ID, "argo")
	forkedAccount := newAccount(runB.ID, "argo-fork")

	newMembership := func(runID, accountID, groupID string) *generated.DirectoryMembership {
		membership, err := suite.client.db.DirectoryMembership.Create().
			SetIntegrationID(integration.ID).
			SetOwnerID(orgUser.OrganizationID).
			SetDirectorySyncRunID(runID).
			SetDirectoryAccountID(accountID).
			SetDirectoryGroupID(groupID).
			Save(ctx)
		assert.NilError(t, err)

		return membership
	}

	survivorMembership := newMembership(runA.ID, survivorAccount.ID, survivorGroup.ID)
	newMembership(runB.ID, forkedAccount.ID, forkedGroup.ID)

	serveropts.BackfillDirectoryDuplicates(ctx, suite.client.db)

	groups, err := suite.client.db.DirectoryGroup.Query().
		Where(directorygroup.IntegrationID(integration.ID)).
		All(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Len(groups, 1), "forked group must be removed")
	assert.Check(t, is.Equal(survivorGroup.ID, groups[0].ID), "earliest group survives")

	accounts, err := suite.client.db.DirectoryAccount.Query().
		Where(directoryaccount.IntegrationID(integration.ID)).
		All(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Len(accounts, 1), "forked account must be removed")
	assert.Check(t, is.Equal(survivorAccount.ID, accounts[0].ID), "earliest account survives")

	memberships, err := suite.client.db.DirectoryMembership.Query().
		Where(directorymembership.IntegrationID(integration.ID)).
		All(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Len(memberships, 1), "fork memberships must be removed")
	assert.Check(t, is.Equal(survivorMembership.ID, memberships[0].ID))

	// idempotent: a clean state is untouched
	serveropts.BackfillDirectoryDuplicates(ctx, suite.client.db)

	groupCount, err := suite.client.db.DirectoryGroup.Query().
		Where(directorygroup.IntegrationID(integration.ID)).
		Count(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(1, groupCount))

	(&Cleanup[*generated.DirectoryMembershipDeleteOne]{client: suite.client.db.DirectoryMembership, ID: survivorMembership.ID}).MustDelete(ctx, t)
	(&Cleanup[*generated.DirectoryAccountDeleteOne]{client: suite.client.db.DirectoryAccount, ID: survivorAccount.ID}).MustDelete(ctx, t)
	(&Cleanup[*generated.DirectoryGroupDeleteOne]{client: suite.client.db.DirectoryGroup, ID: survivorGroup.ID}).MustDelete(ctx, t)
	(&Cleanup[*generated.IntegrationDeleteOne]{client: suite.client.db.Integration, ID: integration.ID}).MustDelete(ctx, t)
}
