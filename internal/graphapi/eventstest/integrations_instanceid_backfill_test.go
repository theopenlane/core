//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/config"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorymembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	"github.com/theopenlane/core/v2/internal/httpserve/serveropts"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	testint "github.com/theopenlane/core/v2/internal/testutils/integrations"
)

// mockProviderToken is the bearer token the mock provider server validates
const mockProviderToken = "mock-token"

// mockAccountRecord renders a directory account record the mock provider serves and ingest maps through unchanged
func mockAccountRecord(externalID string) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(`{"external_id":%q,"canonical_email":%q,"display_name":%q}`, externalID, externalID+"@example.com", externalID))
}

// mockGroupRecord renders a directory group record the mock provider serves
func mockGroupRecord(externalID string) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(`{"external_id":%q,"display_name":%q}`, externalID, externalID))
}

// mockMembershipRecord renders a directory membership record linking an account and group by external id
func mockMembershipRecord(accountExternalID, groupExternalID string) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(`{"directory_account_id":%q,"directory_group_id":%q}`, accountExternalID, groupExternalID))
}

// connectMockInstall creates and connects a mock provider installation through the production
// EnsureInstallation and Reconcile path, returning the reloaded installation once the real health
// check has passed and the instance id has been resolved from what the provider reports
func connectMockInstall(ctx context.Context, t *testing.T, server *testint.MockHTTPServer) *ent.Integration {
	t.Helper()

	def, ok := suite.IntegrationsRT.Registry().Definition(testint.MockHTTPDefinitionID.ID())
	assert.Assert(t, ok, "mock provider definition must be registered on the runtime")

	install, _, err := suite.IntegrationsRT.EnsureInstallation(ctx, th.SharedTestUser1.OrganizationID, "", def)
	th.RequireNoError(t, err)

	credential := testint.MockHTTPCredentialSet(mockProviderToken, server.URL())
	th.RequireNoError(t, suite.IntegrationsRT.Reconcile(ctx, install, nil, testint.MockHTTPCredential.ID(), &credential, nil))

	return reloadIntegration(t, ctx, install.ID)
}

// executeMockSync runs the mock provider's directory sync through the production ExecuteOperation
// entrypoint, which resolves the installation's credential and ingests whatever the provider reports
func executeMockSync(ctx context.Context, t *testing.T, install *ent.Integration) {
	t.Helper()

	operation, err := suite.IntegrationsRT.Registry().Operation(install.DefinitionID, testint.MockHTTPSyncOperation)
	th.RequireNoError(t, err)

	_, err = suite.IntegrationsRT.ExecuteOperation(ctx, install, operation, nil, nil)
	th.RequireNoError(t, err)
}

// clearInstallInstanceID clears one installation's stored instance id to simulate a legacy row
// connected before the instance-id resolver existed, a connected credentialed state the current
// connect flow can no longer produce and precisely what the gate and backfill exist to repair
func clearInstallInstanceID(ctx context.Context, t *testing.T, id string) {
	t.Helper()

	th.RequireNoError(t, suite.Client.DB.Integration.UpdateOneID(id).
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{}).
		Exec(privacy.DecisionContext(ctx, privacy.Allow)))
}

// legacyUnclaimedAccount creates a directory account FK-linked to the installation with no provenance
// columns, the pre-migration shape production ingest never produces and the provenance backfill exists
// to stamp
func legacyUnclaimedAccount(ctx context.Context, t *testing.T, install *ent.Integration, externalID string) *ent.DirectoryAccount {
	t.Helper()

	account, err := suite.Client.DB.DirectoryAccount.Create().
		SetExternalID(externalID).
		SetDisplayName(externalID).
		SetCanonicalEmail(externalID + "@example.com").
		SetOwnerID(install.OwnerID).
		SetIntegrationID(install.ID).
		Save(ctx)
	th.RequireNoError(t, err)

	return account
}

// TestMockProviderIngestStampsResolvedInstanceID drives EnsureInstallation, Reconcile, and
// ExecuteOperation so provider-supplied records flow through the real ingest pipeline, proving connect
// resolves the instance id and ingest stamps it as provenance on every record
func TestMockProviderIngestStampsResolvedInstanceID(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const prefix = "ingest-"

	server := testint.NewMockHTTPServer(mockProviderToken, "tenant-ingest")
	t.Cleanup(server.Close)

	server.SetDirectory(
		[]json.RawMessage{mockAccountRecord(prefix + "a-1"), mockAccountRecord(prefix + "a-2")},
		[]json.RawMessage{mockGroupRecord(prefix + "g-1")},
		[]json.RawMessage{mockMembershipRecord(prefix+"a-1", prefix+"g-1")},
	)

	install := connectMockInstall(ctx, t, server)

	t.Cleanup(func() {
		_, err := suite.Client.DB.DirectoryMembership.Delete().Where(directorymembership.IntegrationID(install.ID)).Exec(ctx)
		th.RequireNoError(t, err)
		_, err = suite.Client.DB.DirectoryAccount.Delete().Where(directoryaccount.ExternalIDHasPrefix(prefix)).Exec(ctx)
		th.RequireNoError(t, err)
		_, err = suite.Client.DB.DirectoryGroup.Delete().Where(directorygroup.ExternalIDHasPrefix(prefix)).Exec(ctx)
		th.RequireNoError(t, err)
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: install.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	assert.Check(t, is.Equal("tenant-ingest", install.InstallationMetadata.Display.ExternalID), "connect resolves the instance id the provider reports")
	assert.Check(t, is.Equal(enums.IntegrationStatusConnected, install.Status), "the installation connects through the real health check")

	executeMockSync(ctx, t, install)

	accounts, err := suite.Client.DB.DirectoryAccount.Query().Where(directoryaccount.ExternalIDHasPrefix(prefix)).All(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, is.Len(accounts, 2), "both provider accounts ingested through the real pipeline")

	for _, account := range accounts {
		assert.Check(t, is.Equal("tenant-ingest", account.SourceInstanceID), "ingest stamps the resolved instance id as provenance")
		assert.Check(t, is.Equal(install.ID, account.ManagedBy), "ingest stamps the managing installation")
	}

	group, err := suite.Client.DB.DirectoryGroup.Query().Where(directorygroup.ExternalID(prefix + "g-1")).Only(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal("tenant-ingest", group.SourceInstanceID), "the provider group ingested with the resolved instance id")
}

// TestInstanceIDGatesIngest proves the ingest pipeline refuses an installation with no instance id and
// resumes once the backfill re-resolves it from the provider, all through the real ExecuteOperation path
func TestInstanceIDGatesIngest(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const prefix = "gate-"

	server := testint.NewMockHTTPServer(mockProviderToken, "tenant-gate")
	t.Cleanup(server.Close)

	server.SetDirectory([]json.RawMessage{mockAccountRecord(prefix + "a-1")}, nil, nil)

	install := connectMockInstall(ctx, t, server)

	t.Cleanup(func() {
		_, err := suite.Client.DB.DirectoryAccount.Delete().Where(directoryaccount.ExternalIDHasPrefix(prefix)).Exec(ctx)
		th.RequireNoError(t, err)
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: install.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	clearInstallInstanceID(ctx, t, install.ID)

	operation, err := suite.IntegrationsRT.Registry().Operation(install.DefinitionID, testint.MockHTTPSyncOperation)
	th.RequireNoError(t, err)

	_, err = suite.IntegrationsRT.ExecuteOperation(ctx, reloadIntegration(t, ctx, install.ID), operation, nil, nil)
	assert.Check(t, errors.Is(err, operations.ErrIngestInstanceIDRequired), "ingest is gated for an installation with no instance id")

	th.RequireNoError(t, suite.IntegrationsRT.BackfillInstallationInstanceID(ctx, reloadIntegration(t, ctx, install.ID)))

	executeMockSync(ctx, t, reloadIntegration(t, ctx, install.ID))

	account, err := suite.Client.DB.DirectoryAccount.Query().Where(directoryaccount.ExternalID(prefix + "a-1")).Only(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal("tenant-gate", account.SourceInstanceID), "ingest stamps the re-resolved instance id once the gate passes")
	assert.Check(t, is.Equal(install.ID, account.ManagedBy), "the ingested record is managed by the installation")
}

// TestHealthAssessmentRefreshesChangedInstanceID proves a health assessment re-resolves and overwrites
// the stored instance id when the provider begins reporting a different one, with no direct write
func TestHealthAssessmentRefreshesChangedInstanceID(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	server := testint.NewMockHTTPServer(mockProviderToken, "tenant-v1")
	t.Cleanup(server.Close)

	install := connectMockInstall(ctx, t, server)

	t.Cleanup(func() {
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: install.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	assert.Check(t, is.Equal("tenant-v1", install.InstallationMetadata.Display.ExternalID), "connect resolves the initial instance id")

	server.SetInstanceID("tenant-v2")

	_, err := suite.IntegrationsRT.RunHealthAssessment(ctx, reloadIntegration(t, ctx, install.ID))
	th.RequireNoError(t, err)

	assert.Check(t, is.Equal("tenant-v2", reloadIntegration(t, ctx, install.ID).InstallationMetadata.Display.ExternalID), "the health assessment refreshes the instance id the provider now reports")
}

// TestReconnectRefreshesChangedInstanceID proves reconnecting with a valid credential refreshes the
// stored instance id to the one the provider now reports instead of rejecting it as an instance mismatch
func TestReconnectRefreshesChangedInstanceID(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	server := testint.NewMockHTTPServer(mockProviderToken, "tenant-v1")
	t.Cleanup(server.Close)

	install := connectMockInstall(ctx, t, server)

	t.Cleanup(func() {
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: install.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	server.SetInstanceID("tenant-v2")

	credential := testint.MockHTTPCredentialSet(mockProviderToken, server.URL())
	th.RequireNoError(t, suite.IntegrationsRT.Reconcile(ctx, reloadIntegration(t, ctx, install.ID), nil, testint.MockHTTPCredential.ID(), &credential, nil))

	assert.Check(t, is.Equal("tenant-v2", reloadIntegration(t, ctx, install.ID).InstallationMetadata.Display.ExternalID), "reconnect refreshes the changed instance id instead of rejecting it as a mismatch")
}

// TestDisconnectReinstallReclaimsIngestedRecord proves a record ingested by one installation is
// re-claimed, not duplicated, by a fresh installation of the same provider on the same instance after
// the first is disconnected, driven entirely through the production connect, disconnect, and ingest flow
func TestDisconnectReinstallReclaimsIngestedRecord(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const prefix = "reinstall-"

	server := testint.NewMockHTTPServer(mockProviderToken, "tenant-reinstall")
	t.Cleanup(server.Close)

	server.SetDirectory([]json.RawMessage{mockAccountRecord(prefix + "a-1")}, nil, nil)

	first := connectMockInstall(ctx, t, server)

	t.Cleanup(func() {
		_, err := suite.Client.DB.DirectoryAccount.Delete().Where(directoryaccount.ExternalIDHasPrefix(prefix)).Exec(ctx)
		th.RequireNoError(t, err)
	})

	executeMockSync(ctx, t, first)

	account, err := suite.Client.DB.DirectoryAccount.Query().Where(directoryaccount.ExternalID(prefix + "a-1")).Only(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal(first.ID, account.ManagedBy), "the first installation manages the ingested record")
	assert.Check(t, is.Equal("tenant-reinstall", account.SourceInstanceID), "the record carries the resolved instance id")

	_, err = suite.IntegrationsRT.Disconnect(ctx, first)
	th.RequireNoError(t, err)

	second := connectMockInstall(ctx, t, server)

	t.Cleanup(func() {
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: second.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	assert.Check(t, is.Equal("tenant-reinstall", second.InstallationMetadata.Display.ExternalID), "the reinstall resolves the same instance id")

	executeMockSync(ctx, t, second)

	count, err := suite.Client.DB.DirectoryAccount.Query().Where(directoryaccount.ExternalID(prefix + "a-1")).Count(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal(1, count), "the reinstall re-claims the existing record instead of duplicating it")

	reclaimed, err := suite.Client.DB.DirectoryAccount.Query().Where(directoryaccount.ExternalID(prefix + "a-1")).Only(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal(second.ID, reclaimed.ManagedBy), "the reinstall re-owns the record on the shared instance id")
	assert.Check(t, is.Equal("tenant-reinstall", reclaimed.SourceInstanceID), "the re-claimed record keeps the shared instance id")
}

// TestBackfillChainResolvesInstanceIDThenStampsProvenance drives the real gala startup backfill chain
// with file-backups disabled: a mock installation connects and resolves a distinct external id, its id
// is cleared to a legacy row, and one legacy unstamped record is left FK-linked to it. The chain must
// re-resolve the instance id before stamping provenance, which the record carrying the re-resolved id
// proves is the order instance-ids ran before provenance
func TestBackfillChainResolvesInstanceIDThenStampsProvenance(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	server := testint.NewMockHTTPServer(mockProviderToken, "tenant-mockhttp")
	t.Cleanup(server.Close)

	install := connectMockInstall(ctx, t, server)

	assert.Check(t, is.Equal(enums.IntegrationStatusConnected, install.Status), "the installation connects through the real health check")
	assert.Check(t, is.Equal("tenant-mockhttp", install.InstallationMetadata.Display.ExternalID), "connect resolves the distinct external id from the provider")

	account := legacyUnclaimedAccount(ctx, t, install, "chain-mock-acct")

	clearInstallInstanceID(ctx, t, install.ID)

	t.Cleanup(func() {
		(&th.Cleanup[*ent.DirectoryAccountDeleteOne]{Client: suite.Client.DB.DirectoryAccount, ID: account.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: install.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	th.RequireNoError(t, serveropts.StartBackfill(ctx, suite.GalaRuntime, config.Server{TrustCenterPreviewCnameTarget: th.PreviewCnameTargetTest}))

	waitForEvents()

	assert.Check(t, is.Equal("tenant-mockhttp", reloadIntegration(t, ctx, install.ID).InstallationMetadata.Display.ExternalID), "the backfill re-resolves the distinct external id through the provider")

	stamped, err := suite.Client.DB.DirectoryAccount.Get(ctx, account.ID)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal("tenant-mockhttp", stamped.SourceInstanceID), "provenance stamps the re-resolved id, proving instance-ids ran before provenance")
	assert.Check(t, is.Equal(install.ID, stamped.ManagedBy), "the stamped record is managed by the installation")
	assert.Check(t, "tenant-mockhttp" != install.ID, "the resolved external id is distinct from the installation id")
}
