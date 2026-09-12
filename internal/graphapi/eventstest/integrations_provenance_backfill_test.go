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
	"github.com/theopenlane/core/v2/internal/httpserve/serveropts"
)

// provenanceBackfillDefinitionID is the definition both provenance backfill installations share
const provenanceBackfillDefinitionID = "def_provbackfill"

// provenanceBackfillDefinitionVersion is the definition version the backfill must copy onto stamped rows
const provenanceBackfillDefinitionVersion = "v1"

// provenanceBackfillTenant is the primary installation's pre-resolved instance id
const provenanceBackfillTenant = "tenant-provbackfill"

// provenanceBackfillOtherTenant is the second installation's pre-resolved instance id
const provenanceBackfillOtherTenant = "tenant-provbackfill-other"

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

// TestBackfillIntegrationProvenanceStampsUnclaimedRows verifies the startup provenance backfill fills
// source_definition_id, source_definition_version, source_instance_id, and managed_by on rows linked to
// exactly one installation whose provenance is missing, leaves rows managed by another installation or
// linked to several installations untouched, and is idempotent on a second run
func TestBackfillIntegrationProvenanceStampsUnclaimedRows(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	installation, err := suite.Client.DB.Integration.Create().
		SetName("Provenance Backfill Primary").
		SetKind("provbackfillprimary").
		SetDefinitionID(provenanceBackfillDefinitionID).
		SetDefinitionVersion(provenanceBackfillDefinitionVersion).
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: provenanceBackfillTenant}}).
		Save(ctx)
	th.RequireNoError(t, err)

	other, err := suite.Client.DB.Integration.Create().
		SetName("Provenance Backfill Other").
		SetKind("provbackfillother").
		SetDefinitionID(provenanceBackfillDefinitionID).
		SetDefinitionVersion(provenanceBackfillDefinitionVersion).
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: provenanceBackfillOtherTenant}}).
		Save(ctx)
	th.RequireNoError(t, err)

	unclaimedAccount, err := suite.Client.DB.DirectoryAccount.Create().
		SetExternalID("provbackfill-acct-unclaimed").
		SetDisplayName("Unclaimed Account").
		SetCanonicalEmail("provbackfill-acct-unclaimed@example.com").
		SetOwnerID(installation.OwnerID).
		SetIntegrationID(installation.ID).
		Save(ctx)
	th.RequireNoError(t, err)

	partialAccount, err := suite.Client.DB.DirectoryAccount.Create().
		SetExternalID("provbackfill-acct-partial").
		SetDisplayName("Partial Account").
		SetCanonicalEmail("provbackfill-acct-partial@example.com").
		SetOwnerID(installation.OwnerID).
		SetIntegrationID(installation.ID).
		SetSourceDefinitionID(provenanceBackfillDefinitionID).
		SetManagedBy(installation.ID).
		Save(ctx)
	th.RequireNoError(t, err)

	foreignAccount, err := suite.Client.DB.DirectoryAccount.Create().
		SetExternalID("provbackfill-acct-foreign").
		SetDisplayName("Foreign Managed Account").
		SetCanonicalEmail("provbackfill-acct-foreign@example.com").
		SetOwnerID(installation.OwnerID).
		SetIntegrationID(installation.ID).
		SetManagedBy(other.ID).
		Save(ctx)
	th.RequireNoError(t, err)

	singleFinding, err := suite.Client.DB.Finding.Create().
		SetExternalID("provbackfill-find-single").
		SetDisplayName("Single Linked Finding").
		SetOwnerID(installation.OwnerID).
		AddIntegrationIDs(installation.ID).
		Save(allowCtx)
	th.RequireNoError(t, err)

	sharedFinding, err := suite.Client.DB.Finding.Create().
		SetExternalID("provbackfill-find-shared").
		SetDisplayName("Shared Linked Finding").
		SetOwnerID(installation.OwnerID).
		AddIntegrationIDs(installation.ID, other.ID).
		Save(allowCtx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		(&th.Cleanup[*ent.FindingDeleteOne]{Client: suite.Client.DB.Finding, IDs: []string{singleFinding.ID, sharedFinding.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*ent.DirectoryAccountDeleteOne]{Client: suite.Client.DB.DirectoryAccount, IDs: []string{unclaimedAccount.ID, partialAccount.ID, foreignAccount.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installation.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: other.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	foreignBefore := directoryAccountProvenance(ctx, t, foreignAccount.ID)
	sharedBefore := findingProvenance(ctx, t, sharedFinding.ID)

	serveropts.BackfillIntegrationProvenance(ctx, suite.Client.DB, suite.IntegrationsRT)

	unclaimedAfter := directoryAccountProvenance(ctx, t, unclaimedAccount.ID)
	assert.Check(t, is.Equal(provenanceBackfillDefinitionID, unclaimedAfter.DefinitionID), "an unclaimed FK-linked row must gain the installation's definition")
	assert.Check(t, is.Equal(provenanceBackfillDefinitionVersion, unclaimedAfter.DefinitionVersion), "an unclaimed FK-linked row must gain the installation's definition version")
	assert.Check(t, is.Equal(provenanceBackfillTenant, unclaimedAfter.InstanceID), "an unclaimed FK-linked row must gain the installation's instance id")
	assert.Check(t, is.Equal(installation.ID, unclaimedAfter.ManagedBy), "an unclaimed FK-linked row must become managed by the installation")

	partialAfter := directoryAccountProvenance(ctx, t, partialAccount.ID)
	assert.Check(t, is.Equal(provenanceBackfillDefinitionID, partialAfter.DefinitionID))
	assert.Check(t, is.Equal(provenanceBackfillDefinitionVersion, partialAfter.DefinitionVersion), "a self-managed row missing only its instance must gain the definition version")
	assert.Check(t, is.Equal(provenanceBackfillTenant, partialAfter.InstanceID), "a self-managed row missing only its instance must gain the instance id")
	assert.Check(t, is.Equal(installation.ID, partialAfter.ManagedBy))

	foreignAfter := directoryAccountProvenance(ctx, t, foreignAccount.ID)
	assert.Check(t, is.DeepEqual(foreignBefore, foreignAfter), "a row managed by another installation must be left untouched")
	assert.Check(t, is.Equal(other.ID, foreignAfter.ManagedBy))
	assert.Check(t, is.Equal("", foreignAfter.InstanceID))

	singleAfter := findingProvenance(ctx, t, singleFinding.ID)
	assert.Check(t, is.Equal(provenanceBackfillDefinitionID, singleAfter.DefinitionID), "an M2M row linked only to the installation must gain its definition")
	assert.Check(t, is.Equal(provenanceBackfillDefinitionVersion, singleAfter.DefinitionVersion))
	assert.Check(t, is.Equal(provenanceBackfillTenant, singleAfter.InstanceID), "an M2M row linked only to the installation must gain its instance id")
	assert.Check(t, is.Equal(installation.ID, singleAfter.ManagedBy))

	sharedAfter := findingProvenance(ctx, t, sharedFinding.ID)
	assert.Check(t, is.DeepEqual(sharedBefore, sharedAfter), "an M2M row linked to more than one installation must be left untouched")
	assert.Check(t, is.Equal("", sharedAfter.ManagedBy))
	assert.Check(t, is.Equal("", sharedAfter.DefinitionID))

	serveropts.BackfillIntegrationProvenance(ctx, suite.Client.DB, suite.IntegrationsRT)

	assert.Check(t, is.DeepEqual(unclaimedAfter, directoryAccountProvenance(ctx, t, unclaimedAccount.ID)), "a second run must not rewrite an already stamped FK-linked row")
	assert.Check(t, is.DeepEqual(partialAfter, directoryAccountProvenance(ctx, t, partialAccount.ID)), "a second run must not rewrite an already stamped self-managed row")
	assert.Check(t, is.DeepEqual(foreignAfter, directoryAccountProvenance(ctx, t, foreignAccount.ID)), "a second run must not touch a foreign-managed row")
	assert.Check(t, is.DeepEqual(singleAfter, findingProvenance(ctx, t, singleFinding.ID)), "a second run must not rewrite an already stamped M2M row")
	assert.Check(t, is.DeepEqual(sharedAfter, findingProvenance(ctx, t, sharedFinding.ID)), "a second run must not touch a multi-linked M2M row")
}
