package objectstore

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/mappingtest"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/utils/jsonx"
)

func builderMappings(t *testing.T) []types.MappingRegistration {
	t.Helper()

	def, err := Builder(&RuntimeConfig{}, "", Config{})()
	assert.NilError(t, err)

	return def.Mappings
}

func TestMappingExpressionsValid(t *testing.T) {
	mappingtest.AssertExpressionsValid(t, builderMappings(t))
}

func TestEntityMapping(t *testing.T) {
	spec := mappingtest.MappingSpec(t, builderMappings(t), entityops.SchemaEntity.Name)

	envelope := types.MappingEnvelope{
		Resource: "vendors/1password-vendor.json#0",
		Payload:  mappingtest.LoadExample(t, "examples", "1password-vendor.json"),
	}

	assert.Check(t, mappingtest.AssertFiltered(t, spec, envelope))

	mapped := mappingtest.EvalMap(t, spec, envelope)

	assert.Equal(t, "entity::MI2d4zsM-AnFxqsI6tNvFg", mapped["external_id"])
	assert.Equal(t, "entity::MI2d4zsM-AnFxqsI6tNvFg", mapped["system_internal_id"])
	assert.Equal(t, "1password-vendor", mapped["name"])
	assert.Equal(t, "1Password", mapped["display_name"])
	assert.DeepEqual(t, []any{"1password.com"}, mapped["domains"])
	assert.DeepEqual(t, []any{"technology", "password management"}, mapped["tags"])
	assert.Equal(t, "https://www.google.com/s2/favicons?domain=1password.com&sz=128", mapped["logo_remote_url"])
	assert.Equal(t, true, mapped["externally_visible"])
	assert.Equal(t, vendorEntityTypeName, mapped[recordEntityTypeKey])
}

func TestEntityExamplesUseInputKeys(t *testing.T) {
	spec := mappingtest.MappingSpec(t, builderMappings(t), entityops.SchemaEntity.Name)

	inputKeys := lo.Map(entityops.SchemaEntity.Fields, func(field entityops.FieldDescriptor, _ int) string { return field.InputKey })
	linkKeys := lo.Map(spec.Links, func(link types.LinkRule, _ int) string { return lo.CoalesceOrEmpty(link.SourceList, link.SourceField) })

	examples, err := filepath.Glob(filepath.Join("examples", "*.json"))
	assert.NilError(t, err)
	assert.Check(t, len(examples) > 0)

	for _, example := range examples {
		t.Run(filepath.Base(example), func(t *testing.T) {
			record, err := jsonx.ToRawMap(mappingtest.LoadExample(t, "examples", filepath.Base(example)))
			assert.NilError(t, err)

			for key := range record {
				assert.Check(t, lo.Contains(inputKeys, key) || lo.Contains(linkKeys, key), "%s is not an entity input or link key", key)
			}
		})
	}
}

func TestEntityMappingEntityType(t *testing.T) {
	spec := mappingtest.MappingSpec(t, builderMappings(t), entityops.SchemaEntity.Name)

	testCases := []struct {
		name     string
		payload  string
		expected string
	}{
		{name: "unnamed type links the vendor type", payload: `{"external_id":"entity::a","name":"a"}`, expected: vendorEntityTypeName},
		{name: "named type is kept", payload: `{"external_id":"entity::b","name":"b","entity_type_name":"customer"}`, expected: "customer"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Payload: json.RawMessage(tc.payload)})

			assert.Equal(t, tc.expected, mapped[recordEntityTypeKey])
		})
	}
}

func TestRecordMappings(t *testing.T) {
	mappings := builderMappings(t)

	mappingtest.AssertExpressionsValid(t, mappings)
	assert.Equal(t, len(recordSchemas), len(mappings))

	for _, schema := range recordSchemas {
		count := lo.CountBy(mappings, func(m types.MappingRegistration) bool { return m.Schema == schema.Name })
		assert.Check(t, count == 1, "%s registered %d times", schema.Name, count)
	}

	for schema, links := range recordLinks {
		_, found := entityops.LookupSchema(schema)
		assert.Check(t, found, "link source %s", schema)

		for _, link := range links {
			assert.Check(t, link.PayloadKeys, "%s %s", schema, link.SourceList)
		}
	}

	for schema := range recordMapExprs {
		_, found := entityops.LookupSchema(schema)
		assert.Check(t, found, "map expression for %s", schema)
	}
}

func TestRecordMappingEdgeKeys(t *testing.T) {
	mappings := builderMappings(t)

	testCases := []struct {
		schema *entityops.Schema
		key    string
	}{
		{schema: entityops.SchemaSystemDetail, key: "platforms_external_ids"},
		{schema: entityops.SchemaPlatform, key: "assets_external_ids"},
		{schema: entityops.SchemaPlatform, key: "entities_external_ids"},
		{schema: entityops.SchemaEntity, key: "assets_external_ids"},
		{schema: entityops.SchemaContact, key: "entities_external_ids"},
		{schema: entityops.SchemaInternalPolicy, key: "programs_external_ids"},
		{schema: entityops.SchemaProcedure, key: "internal_policies_external_ids"},
		{schema: entityops.SchemaEvidence, key: "tasks_external_ids"},
		{schema: entityops.SchemaRemediation, key: "vulnerabilities_external_ids"},
		{schema: entityops.SchemaAssessmentPolicy, key: "assessment_external_id"},
		{schema: entityops.SchemaAssessmentPolicy, key: "internal_policy_external_id"},
		{schema: entityops.SchemaCampaign, key: "email_template_external_id"},
		{schema: entityops.SchemaCampaign, key: "identity_holders_external_ids"},
		{schema: entityops.SchemaInternalPolicy, key: "approver_name"},
		{schema: entityops.SchemaProcedure, key: "approver_name"},
		{schema: entityops.SchemaControl, key: "control_owner_name"},
		{schema: entityops.SchemaControl, key: "delegate_name"},
		{schema: entityops.SchemaEntity, key: "entity_type_name"},
	}

	for _, tc := range testCases {
		t.Run(tc.schema.Snake+"/"+tc.key, func(t *testing.T) {
			spec := mappingtest.MappingSpec(t, mappings, tc.schema.Name)

			found := lo.ContainsBy(spec.Links, func(link types.LinkRule) bool {
				return lo.CoalesceOrEmpty(link.SourceList, link.SourceField) == tc.key
			})
			assert.Check(t, found)
		})
	}
}

func TestRecordMappingFilter(t *testing.T) {
	spec := mappingtest.MappingSpec(t, builderMappings(t), entityops.SchemaControl.Name)

	assert.Check(t, mappingtest.AssertFiltered(t, spec, types.MappingEnvelope{Payload: json.RawMessage(`{"external_id":"control::1","ref_code":"MIN-01"}`)}))
	assert.Check(t, !mappingtest.AssertFiltered(t, spec, types.MappingEnvelope{Payload: json.RawMessage(`{"ref_code":"MIN-01"}`)}), "a record without an external id is filtered")
}
