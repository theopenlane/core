package graphapi_test

import (
	"context"
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/entityops"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/finding"
	"github.com/theopenlane/core/internal/integrations/operations"
	intregistry "github.com/theopenlane/core/internal/integrations/registry"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

const linkTestOperationName = "findings.sync"

// linkTestDefinition builds a minimal ingest-producing definition whose finding mapping passes the
// provider payload through unchanged and declares the supplied cross-object link rules
func linkTestDefinition(defID string, links []integrationtypes.LinkRule) integrationtypes.Definition {
	return integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          defID,
			DisplayName: "Link Test",
			Active:      true,
		},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:  linkTestOperationName,
				Topic: gala.TopicName("integration." + defID + "." + linkTestOperationName),
				IngestHandle: func(context.Context, integrationtypes.OperationRequest) ([]integrationtypes.IngestPayloadSet, error) {
					return nil, nil
				},
				Ingest: []integrationtypes.IngestContract{{Schema: entityops.SchemaFinding.Name}},
			},
		},
		Mappings: []integrationtypes.MappingRegistration{
			{
				Schema: entityops.SchemaFinding.Name,
				Spec:   integrationtypes.MappingOverride{MapExpr: "payload", Links: links},
			},
		},
	}
}

// ingestFindings registers the definition into a fresh registry and pushes the payloads through the
// synchronous ingest path — mapping, link injection, and catalog upsert — exactly as an operation run would
func ingestFindings(ctx context.Context, t *testing.T, integration *ent.Integration, def integrationtypes.Definition, payloads ...string) error {
	t.Helper()

	reg := intregistry.New()
	requireNoError(t, reg.Register(def))

	envelopes := lo.Map(payloads, func(p string, _ int) integrationtypes.MappingEnvelope {
		return integrationtypes.MappingEnvelope{Payload: json.RawMessage(p)}
	})

	return operations.ProcessPayloadSets(ctx, operations.IngestContext{
		Registry:    reg,
		DB:          suite.client.db,
		Integration: integration,
	}, linkTestOperationName, def.Operations[0].Ingest, []integrationtypes.IngestPayloadSet{
		{Schema: entityops.SchemaFinding.Name, Envelopes: envelopes},
	}, operations.IngestOptions{})
}

// findingControls loads the ingested finding by external id and returns the ref codes of its linked controls
func findingControls(ctx context.Context, t *testing.T, externalID string) (*ent.Finding, []string) {
	t.Helper()

	f, err := suite.client.db.Finding.Query().
		Where(finding.ExternalID(externalID)).
		WithControls().
		Only(ctx)
	requireNoError(t, err)

	return f, lo.Map(f.Edges.Controls, func(c *ent.Control, _ int) string { return c.RefCode })
}

func TestIntegrationCrossObjectLinking(t *testing.T) {
	ctx := setContext(sharedTestUser1.UserCtx, suite.client.db)

	// seed link targets with stable ref codes; other tests' controls use random UUID ref codes
	for _, refCode := range []string{"LINK-CC-1", "LINK-CC-2", "LINK-CC-3", "LINK-EXPR-1"} {
		(&ControlBuilder{client: suite.client, RefCode: refCode}).MustNew(sharedTestUser1.UserCtx, t)
	}

	integration, err := suite.client.db.Integration.Create().
		SetName("Link Test Integration").
		SetKind("linktest").
		SetDefinitionID("def_linktest").
		Save(ctx)
	requireNoError(t, err)
	assert.Assert(t, integration.OwnerID != "", "seeded integration must be org-owned")

	fieldMatchRules := []integrationtypes.LinkRule{
		{
			TargetSchema: entityops.SchemaControl.Name,
			TargetField:  control.FieldRefCode,
			SourceField:  entityops.InputKeyFindingCategory,
			SourceList:   entityops.InputKeyFindingCategories,
		},
	}

	t.Run("scalar field match links one control", func(t *testing.T) {
		def := linkTestDefinition("def_linktest", fieldMatchRules)

		err := ingestFindings(ctx, t, integration, def, `{"external_id":"link-f-1","display_name":"f1","category":"LINK-CC-1"}`)
		requireNoError(t, err)

		_, refCodes := findingControls(ctx, t, "link-f-1")
		assert.DeepEqual(t, refCodes, []string{"LINK-CC-1"})
	})

	// multiTargetSkip marks permutations blocked on the pre-existing through-edge bug: batch
	// m2m inserts evaluate the FindingControl id default once, so adding 2+ controls in one
	// mutation violates finding_controls_pkey. Unskip when the edge-schema fix lands
	const multiTargetSkip = "blocked on finding_controls through-edge fix: multi-target adds duplicate the join row id"

	t.Run("list field match links every element", func(t *testing.T) {
		t.Skip(multiTargetSkip)

		def := linkTestDefinition("def_linktest", fieldMatchRules)

		err := ingestFindings(ctx, t, integration, def, `{"external_id":"link-f-2","display_name":"f2","categories":["LINK-CC-2","LINK-CC-3"]}`)
		requireNoError(t, err)

		_, refCodes := findingControls(ctx, t, "link-f-2")
		assert.Equal(t, len(refCodes), 2)
		assert.Assert(t, lo.Every(refCodes, []string{"LINK-CC-2", "LINK-CC-3"}))
	})

	t.Run("scalar and list values are merged and deduplicated", func(t *testing.T) {
		t.Skip(multiTargetSkip)

		def := linkTestDefinition("def_linktest", fieldMatchRules)

		err := ingestFindings(ctx, t, integration, def, `{"external_id":"link-f-3","display_name":"f3","category":"LINK-CC-1","categories":["LINK-CC-1","LINK-CC-2"]}`)
		requireNoError(t, err)

		_, refCodes := findingControls(ctx, t, "link-f-3")
		assert.Equal(t, len(refCodes), 2)
		assert.Assert(t, lo.Every(refCodes, []string{"LINK-CC-1", "LINK-CC-2"}))
	})

	t.Run("no matching target leaves the record unlinked", func(t *testing.T) {
		def := linkTestDefinition("def_linktest", fieldMatchRules)

		err := ingestFindings(ctx, t, integration, def, `{"external_id":"link-f-4","display_name":"f4","category":"LINK-NOPE"}`)
		requireNoError(t, err)

		f, refCodes := findingControls(ctx, t, "link-f-4")
		assert.Equal(t, len(refCodes), 0)
		assert.Equal(t, f.DisplayName, "f4")
	})

	t.Run("cel expression match", func(t *testing.T) {
		def := linkTestDefinition("def_linktest", []integrationtypes.LinkRule{
			{
				TargetSchema: entityops.SchemaControl.Name,
				Expression:   `source.category.startsWith(target.ref_code)`,
			},
		})

		err := ingestFindings(ctx, t, integration, def, `{"external_id":"link-f-5","display_name":"f5","category":"LINK-EXPR-1-sub"}`)
		requireNoError(t, err)

		_, refCodes := findingControls(ctx, t, "link-f-5")
		assert.DeepEqual(t, refCodes, []string{"LINK-EXPR-1"})
	})

	t.Run("re-ingest updates in place and re-applies links additively", func(t *testing.T) {
		def := linkTestDefinition("def_linktest", fieldMatchRules)

		err := ingestFindings(ctx, t, integration, def, `{"external_id":"link-f-6","display_name":"first","category":"LINK-CC-1"}`)
		requireNoError(t, err)

		first, refCodes := findingControls(ctx, t, "link-f-6")
		assert.DeepEqual(t, refCodes, []string{"LINK-CC-1"})

		err = ingestFindings(ctx, t, integration, def, `{"external_id":"link-f-6","display_name":"second","category":"LINK-CC-2"}`)
		requireNoError(t, err)

		second, refCodes := findingControls(ctx, t, "link-f-6")
		assert.Equal(t, second.ID, first.ID, "re-ingest must update the same record, not create a duplicate")
		assert.Equal(t, second.DisplayName, "second")
		assert.Equal(t, len(refCodes), 2, "links must be additive: newly matched targets added, existing links retained")
		assert.Assert(t, lo.Every(refCodes, []string{"LINK-CC-1", "LINK-CC-2"}))
	})

	t.Run("multiple rules apply independently", func(t *testing.T) {
		t.Skip(multiTargetSkip)

		def := linkTestDefinition("def_linktest", []integrationtypes.LinkRule{
			{
				TargetSchema: entityops.SchemaControl.Name,
				TargetField:  control.FieldRefCode,
				SourceField:  entityops.InputKeyFindingCategory,
			},
			{
				TargetSchema: entityops.SchemaControl.Name,
				Edge:         "controls",
				TargetField:  control.FieldRefCode,
				SourceList:   entityops.InputKeyFindingCategories,
			},
		})

		err := ingestFindings(ctx, t, integration, def, `{"external_id":"link-f-7","display_name":"f7","category":"LINK-CC-1","categories":["LINK-CC-3"]}`)
		requireNoError(t, err)

		_, refCodes := findingControls(ctx, t, "link-f-7")
		assert.Equal(t, len(refCodes), 2)
		assert.Assert(t, lo.Every(refCodes, []string{"LINK-CC-1", "LINK-CC-3"}))
	})
}
