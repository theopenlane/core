//go:build test

package eventstest_test

import (
	"context"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	openapi "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	intruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	testint "github.com/theopenlane/core/v2/internal/testutils/integrations"
)

// provenanceConversionDefinitionID is the shared test integration definition every conversion
// fixture installation installs under; it is registered on suite.IntegrationsRT at harness setup
// so EnsureInstallationConverted can resolve a definition for both installations
var provenanceConversionDefinitionID = testint.DefinitionID.ID()

// provenanceConversionDefinitionVersion is the definition version the conversion must copy onto stamped rows
const provenanceConversionDefinitionVersion = "v1"

// provenanceConversionTenant is the primary installation's pre-resolved instance id
const provenanceConversionTenant = "tenant-provconv"

// provenanceConversionOtherTenant is the second installation's pre-resolved instance id
const provenanceConversionOtherTenant = "tenant-provconv-other"

// provenanceConversionStaleTenant is an old-format instance id already stored on a managed row, distinct
// from the installation's current resolved instance id
const provenanceConversionStaleTenant = "tenant-provconv-stale"

// provenanceSnapshot is the provenance state of one row plus its updated_at, for before/after comparison
type provenanceSnapshot struct {
	// DefinitionID is the row's source_definition_id
	DefinitionID string
	// DefinitionVersion is the row's source_definition_version
	DefinitionVersion string
	// InstanceID is the row's source_instance_id
	InstanceID string
	// ManagedBy is the row's managed_by
	ManagedBy string
	// UpdatedAt is the row's updated_at
	UpdatedAt time.Time
}

// directoryAccountProvenance reloads the directory account and returns its provenance snapshot
func directoryAccountProvenance(ctx context.Context, t *testing.T, id string) provenanceSnapshot {
	t.Helper()

	row, err := suite.Client.DB.DirectoryAccount.Get(ctx, id)
	th.RequireNoError(t, err)

	return provenanceSnapshot{
		DefinitionID:      row.SourceDefinitionID,
		DefinitionVersion: row.SourceDefinitionVersion,
		InstanceID:        row.SourceInstanceID,
		ManagedBy:         row.ManagedBy,
		UpdatedAt:         row.UpdatedAt,
	}
}

// findingProvenance reloads the finding and returns its provenance snapshot
func findingProvenance(ctx context.Context, t *testing.T, id string) provenanceSnapshot {
	t.Helper()

	row, err := suite.Client.DB.Finding.Get(ctx, id)
	th.RequireNoError(t, err)

	return provenanceSnapshot{
		DefinitionID:      row.SourceDefinitionID,
		DefinitionVersion: row.SourceDefinitionVersion,
		InstanceID:        row.SourceInstanceID,
		ManagedBy:         row.ManagedBy,
		UpdatedAt:         row.UpdatedAt,
	}
}

// installationProvenanceVersion reloads the installation and returns the provenance version recorded
// in its provider state for the shared test integration definition
func installationProvenanceVersion(ctx context.Context, t *testing.T, id string) string {
	t.Helper()

	row := reloadIntegration(t, ctx, id)

	def, ok := suite.IntegrationsRT.Registry().Definition(provenanceConversionDefinitionID)
	assert.Assert(t, ok, "shared test integration definition must be registered on suite.IntegrationsRT")

	state, err := def.ProviderState(row.ProviderState)
	th.RequireNoError(t, err)

	return state.ProvenanceVersion
}

// provenanceConversionProviderState returns the provider state every fixture installation must
// persist so it resolves as connected through the shared definition's OAuth connection instead of
// as never-connected, which is what lets EnsureInstallationConverted stamp its rows
func provenanceConversionProviderState(t *testing.T) openapi.IntegrationProviderState {
	t.Helper()

	def, ok := suite.IntegrationsRT.Registry().Definition(provenanceConversionDefinitionID)
	assert.Assert(t, ok, "shared test integration definition must be registered on suite.IntegrationsRT")

	state, err := def.WithProviderState(openapi.IntegrationProviderState{}, integrationtypes.DefinitionProviderState{CredentialRef: testint.OAuthCredential.ID()})
	th.RequireNoError(t, err)

	return state
}

// TestEnsureInstallationConvertedStampsOrganizationRows verifies EnsureInstallationConverted fills
// source_definition_id, source_definition_version, source_instance_id, and managed_by on rows linked
// to exactly one installation whose provenance is missing, re-stamps rows already managed by any
// installation in the caller's organization onto that installation's current resolved instance id
// regardless of which installation they are linked to, leaves rows linked to several installations
// untouched, persists the provenance conversion marker and resolved instance id on every installation
// it converts in the caller's organization, and is idempotent on a second call
func TestEnsureInstallationConvertedStampsOrganizationRows(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	providerState := provenanceConversionProviderState(t)

	installation, err := suite.Client.DB.Integration.Create().
		SetName("Provenance Conversion Primary").
		SetKind("provconvprimary").
		SetDefinitionID(provenanceConversionDefinitionID).
		SetDefinitionVersion(provenanceConversionDefinitionVersion).
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: provenanceConversionTenant}}).
		SetProviderState(providerState).
		Save(ctx)
	th.RequireNoError(t, err)

	other, err := suite.Client.DB.Integration.Create().
		SetName("Provenance Conversion Other").
		SetKind("provconvother").
		SetDefinitionID(provenanceConversionDefinitionID).
		SetDefinitionVersion(provenanceConversionDefinitionVersion).
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: provenanceConversionOtherTenant}}).
		SetProviderState(providerState).
		Save(ctx)
	th.RequireNoError(t, err)

	unclaimedAccount, err := suite.Client.DB.DirectoryAccount.Create().
		SetExternalID("provconv-acct-unclaimed").
		SetDisplayName("Unclaimed Account").
		SetCanonicalEmail("provconv-acct-unclaimed@example.com").
		SetOwnerID(installation.OwnerID).
		SetIntegrationID(installation.ID).
		Save(ctx)
	th.RequireNoError(t, err)

	partialAccount, err := suite.Client.DB.DirectoryAccount.Create().
		SetExternalID("provconv-acct-partial").
		SetDisplayName("Partial Account").
		SetCanonicalEmail("provconv-acct-partial@example.com").
		SetOwnerID(installation.OwnerID).
		SetIntegrationID(installation.ID).
		SetSourceDefinitionID(provenanceConversionDefinitionID).
		SetManagedBy(installation.ID).
		Save(ctx)
	th.RequireNoError(t, err)

	foreignAccount, err := suite.Client.DB.DirectoryAccount.Create().
		SetExternalID("provconv-acct-foreign").
		SetDisplayName("Foreign Managed Account").
		SetCanonicalEmail("provconv-acct-foreign@example.com").
		SetOwnerID(installation.OwnerID).
		SetIntegrationID(installation.ID).
		SetManagedBy(other.ID).
		Save(ctx)
	th.RequireNoError(t, err)

	staleManagedAccount, err := suite.Client.DB.DirectoryAccount.Create().
		SetExternalID("provconv-acct-stale-managed").
		SetDisplayName("Stale Instance Managed Account").
		SetCanonicalEmail("provconv-acct-stale-managed@example.com").
		SetOwnerID(installation.OwnerID).
		SetIntegrationID(installation.ID).
		SetSourceDefinitionID(provenanceConversionDefinitionID).
		SetSourceDefinitionVersion(provenanceConversionDefinitionVersion).
		SetSourceInstanceID(provenanceConversionStaleTenant).
		SetManagedBy(installation.ID).
		Save(ctx)
	th.RequireNoError(t, err)

	staleForeignAccount, err := suite.Client.DB.DirectoryAccount.Create().
		SetExternalID("provconv-acct-stale-foreign").
		SetDisplayName("Stale Instance Foreign Account").
		SetCanonicalEmail("provconv-acct-stale-foreign@example.com").
		SetOwnerID(installation.OwnerID).
		SetIntegrationID(installation.ID).
		SetSourceDefinitionID(provenanceConversionDefinitionID).
		SetSourceDefinitionVersion(provenanceConversionDefinitionVersion).
		SetSourceInstanceID(provenanceConversionStaleTenant).
		SetManagedBy(other.ID).
		Save(ctx)
	th.RequireNoError(t, err)

	staleManagedCrossLinkedAccount, err := suite.Client.DB.DirectoryAccount.Create().
		SetExternalID("provconv-acct-stale-managed-cross-linked").
		SetDisplayName("Stale Instance Managed Cross-Linked Account").
		SetCanonicalEmail("provconv-acct-stale-managed-cross-linked@example.com").
		SetOwnerID(installation.OwnerID).
		SetIntegrationID(other.ID).
		SetSourceDefinitionID(provenanceConversionDefinitionID).
		SetSourceDefinitionVersion(provenanceConversionDefinitionVersion).
		SetSourceInstanceID(provenanceConversionStaleTenant).
		SetManagedBy(installation.ID).
		Save(ctx)
	th.RequireNoError(t, err)

	singleFinding, err := suite.Client.DB.Finding.Create().
		SetExternalID("provconv-find-single").
		SetDisplayName("Single Linked Finding").
		SetOwnerID(installation.OwnerID).
		AddIntegrationIDs(installation.ID).
		Save(allowCtx)
	th.RequireNoError(t, err)

	sharedFinding, err := suite.Client.DB.Finding.Create().
		SetExternalID("provconv-find-shared").
		SetDisplayName("Shared Linked Finding").
		SetOwnerID(installation.OwnerID).
		AddIntegrationIDs(installation.ID, other.ID).
		Save(allowCtx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		(&th.Cleanup[*ent.FindingDeleteOne]{Client: suite.Client.DB.Finding, IDs: []string{singleFinding.ID, sharedFinding.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*ent.DirectoryAccountDeleteOne]{Client: suite.Client.DB.DirectoryAccount, IDs: []string{unclaimedAccount.ID, partialAccount.ID, foreignAccount.ID, staleManagedAccount.ID, staleForeignAccount.ID, staleManagedCrossLinkedAccount.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installation.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: other.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	sharedBefore := findingProvenance(ctx, t, sharedFinding.ID)

	th.RequireNoError(t, suite.IntegrationsRT.EnsureInstallationConverted(ctx, installation))

	assert.Check(t, is.Equal(intruntime.ProvenanceSchemeVersion, installationProvenanceVersion(ctx, t, installation.ID)), "the caller installation must persist the provenance conversion marker")
	assert.Check(t, is.Equal(intruntime.ProvenanceSchemeVersion, installationProvenanceVersion(ctx, t, other.ID)), "every other installation in the caller's organization must also be converted")

	assert.Check(t, is.Equal(installation.ID, reloadIntegration(t, ctx, installation.ID).InstallationMetadata.Display.ExternalID), "the caller installation's stored external id must become its own instance id")
	assert.Check(t, is.Equal(other.ID, reloadIntegration(t, ctx, other.ID).InstallationMetadata.Display.ExternalID), "every other installation's stored external id must also become its own instance id")

	unclaimedAfter := directoryAccountProvenance(ctx, t, unclaimedAccount.ID)
	assert.Check(t, is.Equal(provenanceConversionDefinitionID, unclaimedAfter.DefinitionID), "an unclaimed FK-linked row must gain the installation's definition")
	assert.Check(t, is.Equal(provenanceConversionDefinitionVersion, unclaimedAfter.DefinitionVersion), "an unclaimed FK-linked row must gain the installation's definition version")
	assert.Check(t, is.Equal(installation.ID, unclaimedAfter.InstanceID), "an unclaimed FK-linked row must gain the installation's instance id")
	assert.Check(t, is.Equal(installation.ID, unclaimedAfter.ManagedBy), "an unclaimed FK-linked row must become managed by the installation")

	partialAfter := directoryAccountProvenance(ctx, t, partialAccount.ID)
	assert.Check(t, is.Equal(provenanceConversionDefinitionID, partialAfter.DefinitionID))
	assert.Check(t, is.Equal(provenanceConversionDefinitionVersion, partialAfter.DefinitionVersion), "a self-managed row missing only its instance must gain the definition version")
	assert.Check(t, is.Equal(installation.ID, partialAfter.InstanceID), "a self-managed row missing only its instance must gain the instance id")
	assert.Check(t, is.Equal(installation.ID, partialAfter.ManagedBy))

	foreignAfter := directoryAccountProvenance(ctx, t, foreignAccount.ID)
	assert.Check(t, is.Equal(provenanceConversionDefinitionID, foreignAfter.DefinitionID), "a row managed by another installation in the same organization is stamped by that installation's own conversion")
	assert.Check(t, is.Equal(provenanceConversionDefinitionVersion, foreignAfter.DefinitionVersion))
	assert.Check(t, is.Equal(other.ID, foreignAfter.InstanceID), "a row managed by another installation gains that installation's instance id, not the caller's")
	assert.Check(t, is.Equal(other.ID, foreignAfter.ManagedBy))

	staleManagedAfter := directoryAccountProvenance(ctx, t, staleManagedAccount.ID)
	assert.Check(t, is.Equal(provenanceConversionDefinitionID, staleManagedAfter.DefinitionID))
	assert.Check(t, is.Equal(provenanceConversionDefinitionVersion, staleManagedAfter.DefinitionVersion))
	assert.Check(t, is.Equal(installation.ID, staleManagedAfter.InstanceID), "a row managed by the installation with a stale instance id must be re-stamped to the installation's current instance id")
	assert.Check(t, is.Equal(installation.ID, staleManagedAfter.ManagedBy))

	staleForeignAfter := directoryAccountProvenance(ctx, t, staleForeignAccount.ID)
	assert.Check(t, is.Equal(provenanceConversionDefinitionID, staleForeignAfter.DefinitionID))
	assert.Check(t, is.Equal(provenanceConversionDefinitionVersion, staleForeignAfter.DefinitionVersion))
	assert.Check(t, is.Equal(other.ID, staleForeignAfter.InstanceID), "a row managed by another installation with a stale instance id is re-stamped to that installation's own current instance id by its own conversion")
	assert.Check(t, is.Equal(other.ID, staleForeignAfter.ManagedBy))

	staleManagedCrossLinkedAfter := directoryAccountProvenance(ctx, t, staleManagedCrossLinkedAccount.ID)
	assert.Check(t, is.Equal(provenanceConversionDefinitionID, staleManagedCrossLinkedAfter.DefinitionID))
	assert.Check(t, is.Equal(provenanceConversionDefinitionVersion, staleManagedCrossLinkedAfter.DefinitionVersion))
	assert.Check(t, is.Equal(installation.ID, staleManagedCrossLinkedAfter.InstanceID), "a row managed by the installation with a stale instance id must be re-stamped to the installation's current instance id regardless of which installation its integration link points at")
	assert.Check(t, is.Equal(installation.ID, staleManagedCrossLinkedAfter.ManagedBy))

	singleAfter := findingProvenance(ctx, t, singleFinding.ID)
	assert.Check(t, is.Equal(provenanceConversionDefinitionID, singleAfter.DefinitionID), "an M2M row linked only to the installation must gain its definition")
	assert.Check(t, is.Equal(provenanceConversionDefinitionVersion, singleAfter.DefinitionVersion))
	assert.Check(t, is.Equal(installation.ID, singleAfter.InstanceID), "an M2M row linked only to the installation must gain its instance id")
	assert.Check(t, is.Equal(installation.ID, singleAfter.ManagedBy))

	sharedAfter := findingProvenance(ctx, t, sharedFinding.ID)
	assert.Check(t, is.DeepEqual(sharedBefore, sharedAfter), "an M2M row linked to more than one installation must be left untouched")
	assert.Check(t, is.Equal("", sharedAfter.ManagedBy))
	assert.Check(t, is.Equal("", sharedAfter.DefinitionID))

	th.RequireNoError(t, suite.IntegrationsRT.EnsureInstallationConverted(ctx, installation))

	assert.Check(t, is.DeepEqual(unclaimedAfter, directoryAccountProvenance(ctx, t, unclaimedAccount.ID)), "a second call must not rewrite an already stamped FK-linked row")
	assert.Check(t, is.DeepEqual(partialAfter, directoryAccountProvenance(ctx, t, partialAccount.ID)), "a second call must not rewrite an already stamped self-managed row")
	assert.Check(t, is.DeepEqual(foreignAfter, directoryAccountProvenance(ctx, t, foreignAccount.ID)), "a second call must not touch a foreign-managed row")
	assert.Check(t, is.DeepEqual(staleManagedAfter, directoryAccountProvenance(ctx, t, staleManagedAccount.ID)), "a second call must not rewrite a row already re-stamped onto the installation's current instance id")
	assert.Check(t, is.DeepEqual(staleForeignAfter, directoryAccountProvenance(ctx, t, staleForeignAccount.ID)), "a second call must not touch a foreign-managed row even when its instance id is stale")
	assert.Check(t, is.DeepEqual(staleManagedCrossLinkedAfter, directoryAccountProvenance(ctx, t, staleManagedCrossLinkedAccount.ID)), "a second call must not rewrite a cross-linked row already re-stamped onto the installation's current instance id")
	assert.Check(t, is.DeepEqual(singleAfter, findingProvenance(ctx, t, singleFinding.ID)), "a second call must not rewrite an already stamped M2M row")
	assert.Check(t, is.DeepEqual(sharedAfter, findingProvenance(ctx, t, sharedFinding.ID)), "a second call must not touch a multi-linked M2M row")
}
