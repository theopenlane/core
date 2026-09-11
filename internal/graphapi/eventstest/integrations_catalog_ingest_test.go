//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/asset"
	"github.com/theopenlane/core/v2/internal/ent/generated/integrationrun"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	intregistry "github.com/theopenlane/core/v2/internal/integrations/registry"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

const catalogIngestTestOperation = "catalog.ingest"

// catalogIngestTestDefinition builds a minimal asset ingest definition whose mapping passes
// provider payloads through unchanged
func catalogIngestTestDefinition(defID string) integrationtypes.Definition {
	passthrough := integrationtypes.MappingOverride{MapExpr: "payload"}

	return integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          defID,
			DisplayName: "Catalog Ingest Test",
			Active:      true,
		},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:  catalogIngestTestOperation,
				Topic: gala.TopicName("integration." + defID + "." + catalogIngestTestOperation),
				IngestHandle: func(context.Context, integrationtypes.OperationRequest) ([]integrationtypes.IngestPayloadSet, error) {
					return nil, nil
				},
				Ingest: []integrationtypes.IngestContract{{Schema: entityops.SchemaAsset.Name}},
			},
		},
		Mappings: []integrationtypes.MappingRegistration{{Schema: entityops.SchemaAsset.Name, Spec: passthrough}},
	}
}

// ingestAssetPayloads pushes asset payloads through the synchronous catalog ingest path and
// returns the record-level result
func ingestAssetPayloads(ctx context.Context, t *testing.T, integration *ent.Integration, payloads ...string) operations.IngestResult {
	t.Helper()

	def := catalogIngestTestDefinition(integration.DefinitionID)
	reg := intregistry.New()
	th.RequireNoError(t, reg.Register(def))

	envelopes := lo.Map(payloads, func(p string, _ int) integrationtypes.MappingEnvelope {
		return integrationtypes.MappingEnvelope{Payload: json.RawMessage(p)}
	})

	result, err := operations.ProcessPayloadSets(ctx, operations.IngestContext{
		Registry:    reg,
		DB:          suite.Client.DB,
		Integration: integration,
	}, catalogIngestTestOperation, def.Operations[0].Ingest, def.Operations[0].Policy, []integrationtypes.IngestPayloadSet{
		{Schema: entityops.SchemaAsset.Name, Envelopes: envelopes},
	}, operations.IngestOptions{})
	th.RequireNoError(t, err)

	return result
}

// catalogAssetBySourceIdentifier loads one ingested asset by its lookup key
func catalogAssetBySourceIdentifier(ctx context.Context, t *testing.T, sourceIdentifier string) *ent.Asset {
	t.Helper()

	a, err := suite.Client.DB.Asset.Query().
		Where(asset.SourceIdentifier(sourceIdentifier)).
		Only(ctx)
	th.RequireNoError(t, err)

	return a
}

func TestCatalogUpsertUnchangedGate(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	integration, err := suite.Client.DB.Integration.Create().
		SetName("Catalog Upsert Gate Test").
		SetKind("catgatetest").
		SetDefinitionID("def_catgatetest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-catgatetest"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		assets, err := suite.Client.DB.Asset.Query().Where(asset.SourceIdentifierHasPrefix("catgate-")).All(ctx)
		th.RequireNoError(t, err)

		if len(assets) > 0 {
			(&th.Cleanup[*ent.AssetDeleteOne]{Client: suite.Client.DB.Asset, IDs: lo.Map(assets, func(a *ent.Asset, _ int) string { return a.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: integration.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	payload := `{"source_identifier":"catgate-asset-1","name":"Gate Asset"}`

	result := ingestAssetPayloads(ctx, t, integration, payload)
	assert.Check(t, is.Equal(1, result.Persisted))
	assert.Check(t, is.Equal(0, result.Failed))
	assert.Check(t, is.Equal(1, result.Changed), "a created row must count as changed")

	created := catalogAssetBySourceIdentifier(ctx, t, "catgate-asset-1")

	result = ingestAssetPayloads(ctx, t, integration, payload)
	assert.Check(t, is.Equal(1, result.Persisted))
	assert.Check(t, is.Equal(0, result.Changed), "an unchanged payload must not count as changed")

	unchanged := catalogAssetBySourceIdentifier(ctx, t, "catgate-asset-1")
	assert.Check(t, unchanged.UpdatedAt.Equal(created.UpdatedAt), "an unchanged payload must not rewrite the row")

	result = ingestAssetPayloads(ctx, t, integration, `{"source_identifier":"catgate-asset-1","name":"Gate Asset Two"}`)
	assert.Check(t, is.Equal(1, result.Persisted))
	assert.Check(t, is.Equal(1, result.Changed), "a changed payload must count as changed")

	after := catalogAssetBySourceIdentifier(ctx, t, "catgate-asset-1")
	assert.Check(t, is.Equal("Gate Asset Two", after.Name))
}

// TestCatalogUpsertDuplicateKeyInOneRunConvergesOnOneRow verifies a duplicate lookup value within
// one run converges on the same row: the first record creates it and the second record's lookup
// sees the row the first created, updating it instead of racing a second create
func TestCatalogUpsertDuplicateKeyInOneRunConvergesOnOneRow(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	integration, err := suite.Client.DB.Integration.Create().
		SetName("Catalog Dup Key Test").
		SetKind("catdupkey").
		SetDefinitionID("def_catdupkey").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-catdupkey"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		assets, err := suite.Client.DB.Asset.Query().Where(asset.SourceIdentifierHasPrefix("catdup-")).All(ctx)
		th.RequireNoError(t, err)

		if len(assets) > 0 {
			(&th.Cleanup[*ent.AssetDeleteOne]{Client: suite.Client.DB.Asset, IDs: lo.Map(assets, func(a *ent.Asset, _ int) string { return a.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: integration.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	first := `{"source_identifier":"catdup-asset-1","name":"First"}`
	second := `{"source_identifier":"catdup-asset-1","name":"Second"}`

	result := ingestAssetPayloads(ctx, t, integration, first, second)
	assert.Check(t, is.Equal(0, result.Failed))
	assert.Check(t, is.Equal(2, result.Persisted), "both records in the preloaded run must persist against the same row")
	assert.Check(t, is.Equal(2, result.Changed), "the create and the second record's material update both count as changed")

	rows, err := suite.Client.DB.Asset.Query().Where(asset.SourceIdentifier("catdup-asset-1")).All(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal(1, len(rows)), "the duplicate lookup value in one run must converge on one row, not create a second")
	assert.Check(t, is.Equal("Second", rows[0].Name), "the second record in the run must update the row the first record created")
}

// TestCatalogClaimUnclaimedRowTakenOverInOneWrite verifies an ingest payload takes over a row that
// carries no recorded source definition, in the same write that applies its other field changes
func TestCatalogClaimUnclaimedRowTakenOverInOneWrite(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	integration, err := suite.Client.DB.Integration.Create().
		SetName("Catalog Claim Unclaimed Test").
		SetKind("catclaimunclaimed").
		SetDefinitionID("def_catclaimunclaimed").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-catclaimunclaimed"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	unclaimed, err := suite.Client.DB.Asset.Create().
		SetSourceIdentifier("catclaim-unclaimed-1").
		SetName("Pre-existing Asset").
		Save(privacy.DecisionContext(ctx, privacy.Allow))
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		(&th.Cleanup[*ent.AssetDeleteOne]{Client: suite.Client.DB.Asset, ID: unclaimed.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: integration.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	result := ingestAssetPayloads(ctx, t, integration, `{"source_identifier":"catclaim-unclaimed-1","name":"Claimed Asset"}`)
	assert.Check(t, is.Equal(0, result.Failed))
	assert.Check(t, is.Equal(1, result.Persisted))
	assert.Check(t, is.Equal(1, result.Changed), "the name change and the claim land in the same write")

	after := catalogAssetBySourceIdentifier(ctx, t, "catclaim-unclaimed-1")
	assert.Check(t, is.Equal(unclaimed.ID, after.ID), "the takeover must update the existing row, not create a new one")
	assert.Check(t, is.Equal("Claimed Asset", after.Name))
	assert.Check(t, is.Equal(integration.ID, after.IntegrationID))
	assert.Check(t, is.Equal(integration.ID, after.ManagedBy))
}

// TestCatalogClaimActiveOtherInstallationKeepsPointers verifies a row managed by another still-active
// installation of the same definition keeps its ownership pointers while its other fields still diff
// and update normally
func TestCatalogClaimActiveOtherInstallationKeepsPointers(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const sharedDefinitionID = "def_catclaimactive"

	installationA, err := suite.Client.DB.Integration.Create().
		SetName("Catalog Claim Active A").
		SetKind("catclaimactivea").
		SetDefinitionID(sharedDefinitionID).
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-catclaimactivea"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	installationB, err := suite.Client.DB.Integration.Create().
		SetName("Catalog Claim Active B").
		SetKind("catclaimactiveb").
		SetDefinitionID(sharedDefinitionID).
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-catclaimactivea"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		assets, err := suite.Client.DB.Asset.Query().Where(asset.SourceIdentifierHasPrefix("catclaim-active-")).All(ctx)
		th.RequireNoError(t, err)

		if len(assets) > 0 {
			(&th.Cleanup[*ent.AssetDeleteOne]{Client: suite.Client.DB.Asset, IDs: lo.Map(assets, func(a *ent.Asset, _ int) string { return a.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installationA.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installationB.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	resultA := ingestAssetPayloads(ctx, t, installationA, `{"source_identifier":"catclaim-active-1","name":"Owned By A"}`)
	assert.Check(t, is.Equal(1, resultA.Changed))

	resultB := ingestAssetPayloads(ctx, t, installationB, `{"source_identifier":"catclaim-active-1","name":"Renamed By B"}`)
	assert.Check(t, is.Equal(0, resultB.Failed))
	assert.Check(t, is.Equal(1, resultB.Changed), "the other field's material change must still apply")

	after := catalogAssetBySourceIdentifier(ctx, t, "catclaim-active-1")
	assert.Check(t, is.Equal(installationA.ID, after.IntegrationID), "an active other installation must keep the ownership pointer")
	assert.Check(t, is.Equal(installationA.ID, after.ManagedBy), "an active other installation must keep managed_by")
	assert.Check(t, is.Equal("Renamed By B", after.Name), "the non-ownership field change must still be applied")
}

// TestCatalogClaimGoneOtherInstallationRepoints verifies a row whose managing installation no longer
// exists is repointed to the new installation in the same update as any other field change
func TestCatalogClaimGoneOtherInstallationRepoints(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const sharedDefinitionID = "def_catclaimgone"

	installationA, err := suite.Client.DB.Integration.Create().
		SetName("Catalog Claim Gone A").
		SetKind("catclaimgonea").
		SetDefinitionID(sharedDefinitionID).
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-catclaimgonea"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	installationB, err := suite.Client.DB.Integration.Create().
		SetName("Catalog Claim Gone B").
		SetKind("catclaimgoneb").
		SetDefinitionID(sharedDefinitionID).
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-catclaimgonea"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		assets, err := suite.Client.DB.Asset.Query().Where(asset.SourceIdentifierHasPrefix("catclaim-gone-")).All(ctx)
		th.RequireNoError(t, err)

		if len(assets) > 0 {
			(&th.Cleanup[*ent.AssetDeleteOne]{Client: suite.Client.DB.Asset, IDs: lo.Map(assets, func(a *ent.Asset, _ int) string { return a.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installationB.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	resultA := ingestAssetPayloads(ctx, t, installationA, `{"source_identifier":"catclaim-gone-1","name":"Owned By A"}`)
	assert.Check(t, is.Equal(1, resultA.Changed))

	th.RequireNoError(t, suite.Client.DB.Integration.DeleteOneID(installationA.ID).Exec(ctx))

	resultB := ingestAssetPayloads(ctx, t, installationB, `{"source_identifier":"catclaim-gone-1","name":"Adopted By B"}`)
	assert.Check(t, is.Equal(0, resultB.Failed))
	assert.Check(t, is.Equal(1, resultB.Changed), "the repoint and the name change land in the same write")

	after := catalogAssetBySourceIdentifier(ctx, t, "catclaim-gone-1")
	assert.Check(t, is.Equal(installationB.ID, after.IntegrationID), "a gone manager must be repointed to the adopting installation")
	assert.Check(t, is.Equal(installationB.ID, after.ManagedBy))
	assert.Check(t, is.Equal("Adopted By B", after.Name))
}

// TestCatalogIntegrationRunsEdgeOnChangedRow verifies a material catalog ingest change adds the
// ingesting run to the row's integration_runs edge, not just its scalar integration_run_id field
func TestCatalogIntegrationRunsEdgeOnChangedRow(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	integration, err := suite.Client.DB.Integration.Create().
		SetName("Catalog Run Edge Test").
		SetKind("catrunedge").
		SetDefinitionID("def_catrunedge").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-catrunedge"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	run, err := suite.Client.DB.IntegrationRun.Create().
		SetIntegrationID(integration.ID).
		SetOwnerID(integration.OwnerID).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		assets, err := suite.Client.DB.Asset.Query().Where(asset.SourceIdentifierHasPrefix("catrunedge-")).All(ctx)
		th.RequireNoError(t, err)

		if len(assets) > 0 {
			(&th.Cleanup[*ent.AssetDeleteOne]{Client: suite.Client.DB.Asset, IDs: lo.Map(assets, func(a *ent.Asset, _ int) string { return a.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationRunDeleteOne]{Client: suite.Client.DB.IntegrationRun, ID: run.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: integration.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	def := catalogIngestTestDefinition(integration.DefinitionID)
	reg := intregistry.New()
	th.RequireNoError(t, reg.Register(def))

	result, err := operations.ProcessPayloadSets(ctx, operations.IngestContext{
		Registry:    reg,
		DB:          suite.Client.DB,
		Integration: integration,
	}, catalogIngestTestOperation, def.Operations[0].Ingest, def.Operations[0].Policy, []integrationtypes.IngestPayloadSet{
		{Schema: entityops.SchemaAsset.Name, Envelopes: []integrationtypes.MappingEnvelope{{Payload: json.RawMessage(`{"source_identifier":"catrunedge-1","name":"Run Edge Asset"}`)}}},
	}, operations.IngestOptions{RunID: run.ID})
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal(1, result.Changed))

	created := catalogAssetBySourceIdentifier(ctx, t, "catrunedge-1")
	assert.Check(t, is.Equal(run.ID, created.IntegrationRunID), "the created row must carry the ingesting run's id")

	onEdge, err := created.QueryIntegrationRuns().Where(integrationrun.ID(run.ID)).Exist(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, onEdge, "a changed row must add the ingesting run to its integration_runs edge, not just the scalar field")
}
