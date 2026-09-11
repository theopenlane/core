//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/risk"
	"github.com/theopenlane/core/v2/internal/ent/generated/vulnerability"
	"github.com/theopenlane/core/v2/internal/graphapi"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	intregistry "github.com/theopenlane/core/v2/internal/integrations/registry"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// catalogFixtureIngestOperation names the shared ingest operation used by the catalog volatile-field regression tests
const catalogFixtureIngestOperation = "catalog.fixture.ingest"

// catalogFixtureDefinition builds a minimal ingest definition for schemaName whose mapping passes provider payloads through unchanged
func catalogFixtureDefinition(defID string, schemaName string) integrationtypes.Definition {
	passthrough := integrationtypes.MappingOverride{MapExpr: "payload"}

	return integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          defID,
			DisplayName: "Catalog Fixture Ingest Test",
			Active:      true,
		},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:  catalogFixtureIngestOperation,
				Topic: gala.TopicName("integration." + defID + "." + catalogFixtureIngestOperation),
				IngestHandle: func(context.Context, integrationtypes.OperationRequest) ([]integrationtypes.IngestPayloadSet, error) {
					return nil, nil
				},
				Ingest: []integrationtypes.IngestContract{{Schema: schemaName}},
			},
		},
		Mappings: []integrationtypes.MappingRegistration{{Schema: schemaName, Spec: passthrough}},
	}
}

// ingestCatalogPayload pushes one payload for schemaName through the synchronous catalog ingest path and returns the record-level result
func ingestCatalogPayload(ctx context.Context, t *testing.T, installation *ent.Integration, schemaName string, payload json.RawMessage) operations.IngestResult {
	t.Helper()

	def := catalogFixtureDefinition(installation.DefinitionID, schemaName)
	reg := intregistry.New()
	th.RequireNoError(t, reg.Register(def))

	result, err := operations.ProcessPayloadSets(ctx, operations.IngestContext{
		Registry:    reg,
		DB:          suite.Client.DB,
		Integration: installation,
	}, catalogFixtureIngestOperation, def.Operations[0].Ingest, def.Operations[0].Policy, []integrationtypes.IngestPayloadSet{
		{Schema: schemaName, Envelopes: []integrationtypes.MappingEnvelope{{Payload: payload}}},
	}, operations.IngestOptions{})
	th.RequireNoError(t, err)

	return result
}

// catalogListenerCounts tallies create and update mutation events observed for one schema during a test
type catalogListenerCounts struct {
	// Creates counts OpCreate mutation events
	Creates *atomic.Int64
	// Updates counts OpUpdateOne mutation events
	Updates *atomic.Int64
	// Teardown removes the mutation listener registered for these counts
	Teardown func()
}

// catalogEventCounters registers a create/update mutation listener for schema and returns the tally the caller asserts against
func catalogEventCounters(t *testing.T, schema *entityops.Schema) catalogListenerCounts {
	t.Helper()

	var creates, updates atomic.Int64

	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, []gala.Registration{
		entityops.MutationListener{
			Schema:     schema,
			Operations: []string{entityops.OpCreate, entityops.OpUpdateOne},
			Handle: func(_ entityops.Invocation, payload entityops.MutationPayload) error {
				if payload.Operation == entityops.OpCreate {
					creates.Add(1)
				} else {
					updates.Add(1)
				}

				return nil
			},
		},
	})
	assert.NilError(t, err)

	return catalogListenerCounts{Creates: &creates, Updates: &updates, Teardown: setup.Teardown}
}

// findingCatalogPayload returns the minimal Finding create-input payload for externalID
func findingCatalogPayload(externalID string) json.RawMessage {
	return json.RawMessage(`{"external_id":"` + externalID + `","display_name":"Catalog Finding","description":"initial desc"}`)
}

// vulnerabilityCatalogPayload returns the minimal Vulnerability create-input payload for externalID
func vulnerabilityCatalogPayload(externalID string) json.RawMessage {
	return json.RawMessage(`{"external_id":"` + externalID + `","display_name":"Catalog Vulnerability","description":"initial desc"}`)
}

// assetCatalogPayload returns the minimal Asset create-input payload for sourceIdentifier
func assetCatalogPayload(sourceIdentifier string) json.RawMessage {
	return json.RawMessage(`{"source_identifier":"` + sourceIdentifier + `","name":"Catalog Asset"}`)
}

// riskCatalogPayload returns the minimal Risk create-input payload for externalID
func riskCatalogPayload(externalID string) json.RawMessage {
	return json.RawMessage(`{"external_id":"` + externalID + `","name":"Catalog Risk"}`)
}

// vulnerabilityByExternalID loads one ingested vulnerability by its lookup key
func vulnerabilityByExternalID(ctx context.Context, t *testing.T, externalID string) *ent.Vulnerability {
	t.Helper()

	v, err := suite.Client.DB.Vulnerability.Query().Where(vulnerability.ExternalID(externalID)).Only(ctx)
	th.RequireNoError(t, err)

	return v
}

// riskByExternalID loads one ingested risk by its lookup key
func riskByExternalID(ctx context.Context, t *testing.T, externalID string) *ent.Risk {
	t.Helper()

	r, err := suite.Client.DB.Risk.Query().Where(risk.ExternalID(externalID)).Only(ctx)
	th.RequireNoError(t, err)

	return r
}
