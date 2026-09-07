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

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/asset"
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
		VirtualUser: testVirtualUser(),
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
	}, catalogIngestTestOperation, def.Operations[0].Ingest, []integrationtypes.IngestPayloadSet{
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
