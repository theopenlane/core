//go:build test

package eventstest_test

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	openapi "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
)

// directoryScopeTestInstance is the shared external instance id two installations of one definition connect to
const directoryScopeTestInstance = "tenant-dirscope"

// newDirectoryScopeInstallation creates one installation of the directory sync test definition against the shared scope instance
func newDirectoryScopeInstallation(t *testing.T, name string) *ent.Integration {
	t.Helper()

	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	integration, err := suite.Client.DB.Integration.Create().
		SetName(name).
		SetKind("dirscopetest").
		SetDefinitionID("def_dirscopetest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: directoryScopeTestInstance}}).
		Save(ctx)
	th.RequireNoError(t, err)

	return integration
}

// TestDirectorySnapshotRemovalScopedToManagingInstallation verifies a complete snapshot from one installation only marks its own rows removed, never rows another installation of the same definition and instance manages
func TestDirectorySnapshotRemovalScopedToManagingInstallation(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const prefix = "dirscope"

	first := newDirectoryScopeInstallation(t, "Scope First Install")
	second := newDirectoryScopeInstallation(t, "Scope Second Install")

	cleanupDirectoryPrefix(t, ctx, []string{first.ID, second.ID}, prefix)

	firstSnapshot := directorySnapshot{Prefix: prefix, Accounts: []directoryAccountRecord{newDirectoryAccountRecord(prefix+"-first-1", "First User")}}
	secondSnapshot := directorySnapshot{Prefix: prefix, Accounts: []directoryAccountRecord{newDirectoryAccountRecord(prefix+"-second-1", "Second User")}}

	seeded := ingestDirectorySnapshotFixture(ctx, t, first, firstSnapshot, true)
	assert.Check(t, is.Equal(0, seeded.Failed))
	assert.Check(t, is.Equal(1, seeded.Changed))

	other := ingestDirectorySnapshotFixture(ctx, t, second, secondSnapshot, true)
	assert.Check(t, is.Equal(0, other.Failed))
	assert.Check(t, is.Equal(0, other.Removed), "a second installation's complete snapshot must not remove rows it does not manage")

	firstAccount := directoryAccountByExternalID(ctx, t, prefix+"-first-1")
	assert.Check(t, firstAccount.RemovedAt == nil, "the first installation's account must stay active after the second installation's snapshot")

	emptied := ingestDirectorySnapshotFixture(ctx, t, first, directorySnapshot{Prefix: prefix}, true)
	assert.Check(t, is.Equal(0, emptied.Failed))
	assert.Check(t, is.Equal(1, emptied.Removed), "the managing installation's complete snapshot still removes its own absent rows")

	removed, err := suite.Client.DB.DirectoryAccount.Get(ctx, firstAccount.ID)
	th.RequireNoError(t, err)
	assert.Check(t, removed.RemovedAt != nil)

	secondAccount := directoryAccountByExternalID(ctx, t, prefix+"-second-1")
	assert.Check(t, secondAccount.RemovedAt == nil, "the other installation's account must be untouched by the first installation's snapshot")
}

// TestDirectoryIngestExcludedRecordSkipsPersist verifies a record tracked as failing from an earlier run is skipped without a write, its attempt counter advances on the installation's health, and only that record is shielded from snapshot removal
func TestDirectoryIngestExcludedRecordSkipsPersist(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const prefix = "direxcluded"

	excludedExternalID := prefix + "-acct-1"
	otherExternalID := prefix + "-acct-2"

	integration, err := suite.Client.DB.Integration.Create().
		SetName("Excluded Record Test").
		SetKind("direxcludedtest").
		SetDefinitionID("def_direxcludedtest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-direxcluded"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	cleanupDirectoryPrefix(t, ctx, []string{integration.ID}, prefix)

	both := directorySnapshot{Prefix: prefix, Accounts: []directoryAccountRecord{
		newDirectoryAccountRecord(excludedExternalID, "Excluded User"),
		newDirectoryAccountRecord(otherExternalID, "Included User"),
	}}

	seeded := ingestDirectorySnapshotFixture(ctx, t, integration, both, true)
	assert.Check(t, is.Equal(2, seeded.Changed), "both accounts are created before any exclusion is tracked")

	tracked, err := suite.Client.DB.Integration.UpdateOneID(integration.ID).
		SetHealth(models.IntegrationHealth{FailedRecords: []models.FailedRecord{{Schema: "directory_account", Key: excludedExternalID, Attempts: 1, LastError: "seeded failure"}}}).
		Save(ctx)
	th.RequireNoError(t, err)

	result := ingestDirectorySnapshotFixture(ctx, t, tracked, both.withAccountMaterialChange(excludedExternalID), true)
	assert.Check(t, is.Equal(1, result.Excluded), "the tracked record must be counted excluded")
	assert.Check(t, is.Equal(0, result.Changed), "the excluded record's change must not be written")
	assert.Check(t, is.Equal(0, result.Removed), "an excluded record present in the snapshot must not be removed")
	assert.Check(t, is.Equal(0, result.Failed))

	excluded := directoryAccountByExternalID(ctx, t, excludedExternalID)
	assert.Check(t, is.Equal("Excluded User", excluded.DisplayName), "an excluded record must keep its stored values")
	assert.Check(t, excluded.RemovedAt == nil)

	refreshed, err := suite.Client.DB.Integration.Get(ctx, integration.ID)
	th.RequireNoError(t, err)
	assert.Check(t, is.Len(refreshed.Health.FailedRecords, 1))
	assert.Check(t, is.Equal(2, refreshed.Health.FailedRecords[0].Attempts), "skipping an excluded record advances its attempt counter")

	onlyExcluded := both.withoutAccount(otherExternalID)

	removal := ingestDirectorySnapshotFixture(ctx, t, refreshed, onlyExcluded, true)
	assert.Check(t, is.Equal(1, removal.Excluded))
	assert.Check(t, is.Equal(1, removal.Removed), "an absent untracked record is still removed while another record is excluded")

	other := directoryAccountByExternalID(ctx, t, otherExternalID)
	assert.Check(t, other.RemovedAt != nil, "the absent record must be marked removed")

	stillExcluded := directoryAccountByExternalID(ctx, t, excludedExternalID)
	assert.Check(t, stillExcluded.RemovedAt == nil, "the excluded record present in the snapshot must stay active")
}

// TestDirectorySyncAdoptsLegacyScientificKeys verifies an account row still keyed by the scientific notation form of its numeric external id is re-keyed in place by the next sync instead of duplicated
func TestDirectorySyncAdoptsLegacyScientificKeys(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const (
		canonicalExternalID = "147884153"
		legacyExternalID    = "1.47884153e+08"
	)

	integration, err := suite.Client.DB.Integration.Create().
		SetName("Legacy Key Adoption Test").
		SetKind("dirlegacykeytest").
		SetDefinitionID("def_dirlegacykeytest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-dirlegacykey"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	seeded, err := suite.Client.DB.DirectoryAccount.Create().
		SetExternalID(legacyExternalID).
		SetDisplayName("Legacy Keyed User").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SetIntegrationID(integration.ID).
		SetManagedBy(integration.ID).
		SetSourceDefinitionID(integration.DefinitionID).
		SetSourceInstanceID("tenant-dirlegacykey").
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		(&th.Cleanup[*ent.DirectoryAccountDeleteOne]{Client: suite.Client.DB.DirectoryAccount, ID: seeded.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: integration.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	snapshot := directorySnapshot{Accounts: []directoryAccountRecord{newDirectoryAccountRecord(canonicalExternalID, "Legacy Keyed User")}}

	result := ingestDirectorySnapshotFixture(ctx, t, integration, snapshot, true)
	assert.Check(t, is.Equal(0, result.Failed))
	assert.Check(t, is.Equal(0, result.Removed), "the adopted row is seen by the snapshot, never removed")

	count, err := suite.Client.DB.DirectoryAccount.Query().
		Where(directoryaccount.ExternalIDIn(canonicalExternalID, legacyExternalID), directoryaccount.OwnerID(th.SharedTestUser1.OrganizationID)).
		Count(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal(1, count), "adoption must not create a second row")

	adopted, err := suite.Client.DB.DirectoryAccount.Get(ctx, seeded.ID)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal(canonicalExternalID, adopted.ExternalID), "the legacy key is repaired in place")
	assert.Check(t, adopted.RemovedAt == nil)
}

// TestRetryRunCreatesFreshPendingRun verifies a re-executed attempt of a terminal run continues under a new pending run that records the run it retries
func TestRetryRunCreatesFreshPendingRun(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	integration, err := suite.Client.DB.Integration.Create().
		SetName("Retry Run Test").
		SetKind("retryruntest").
		SetDefinitionID("def_retryruntest").
		Save(ctx)
	th.RequireNoError(t, err)

	failed, err := suite.Client.DB.IntegrationRun.Create().
		SetIntegrationID(integration.ID).
		SetOwnerID(integration.OwnerID).
		SetOperationName("directory.sync").
		SetOperationKind(enums.IntegrationOperationKindSync).
		SetRunType(enums.IntegrationRunTypeReconcile).
		SetStatus(enums.IntegrationRunStatusFailed).
		Save(ctx)
	th.RequireNoError(t, err)

	retry, err := operations.RetryRun(ctx, suite.Client.DB, failed)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		(&th.Cleanup[*ent.IntegrationRunDeleteOne]{Client: suite.Client.DB.IntegrationRun, IDs: []string{failed.ID, retry.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: integration.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	assert.Check(t, retry.ID != failed.ID)
	assert.Check(t, is.Equal(enums.IntegrationRunStatusPending, retry.Status))
	assert.Check(t, is.Equal(failed.IntegrationID, retry.IntegrationID))
	assert.Check(t, is.Equal(failed.OperationName, retry.OperationName))
	assert.Check(t, is.Equal(failed.OperationKind, retry.OperationKind))
	assert.Check(t, is.Equal(failed.RunType, retry.RunType))
	assert.Check(t, is.Equal(failed.ID, retry.Metrics["retry_of"]), "the retry records the run it continues")
}
