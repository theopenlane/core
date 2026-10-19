package objectstore

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/mappingtest"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/celx"
)

// vendorMappingSpec returns the vendor kind's mapping as the builder registers it
func vendorMappingSpec(t *testing.T) types.MappingOverride {
	t.Helper()

	def, err := Builder(&RuntimeConfig{}, "", Config{})()
	assert.NilError(t, err)

	mapping, found := lo.Find(def.Mappings, func(m types.MappingRegistration) bool {
		return m.Schema == entityMapping.Schema && m.Variant == ""
	})
	assert.Assert(t, found, "expected the builder to register the Entity mapping")
	assert.DeepEqual(t, entityMapping.Spec, mapping.Spec)

	return mapping.Spec
}

func TestMappingExpressionsValid(t *testing.T) {
	def, err := Builder(&RuntimeConfig{}, "", Config{})()
	assert.NilError(t, err)

	for _, m := range def.Mappings {
		name := m.Schema

		t.Run(name+"/filter", func(t *testing.T) {
			assert.NilError(t, providerkit.ValidateExpr(m.Spec.FilterExpr))
		})

		t.Run(name+"/map", func(t *testing.T) {
			assert.NilError(t, providerkit.ValidateExpr(m.Spec.MapExpr))
		})
	}
}

func TestEntityMapping(t *testing.T) {
	spec := vendorMappingSpec(t)

	envelope := types.MappingEnvelope{
		Resource: "vendors/1password-vendor.json#0",
		Payload:  mappingtest.LoadExample(t, "examples", "1password-vendor.json"),
	}

	assert.Assert(t, mappingtest.AssertFiltered(t, spec, envelope), "expected the vendor record to pass the vendor kind filter")

	mapped := mappingtest.EvalMap(t, spec, envelope)

	assert.Equal(t, "entity::MI2d4zsM-AnFxqsI6tNvFg", mapped["external_id"])
	assert.Equal(t, "entity::MI2d4zsM-AnFxqsI6tNvFg", mapped["system_internal_id"])
	assert.Equal(t, "1password-vendor", mapped["name"])
	assert.Equal(t, "1Password", mapped["display_name"])
	assert.DeepEqual(t, []any{"1password.com"}, mapped["domains"])
	assert.DeepEqual(t, []any{"technology", "password management"}, mapped["tags"])
	assert.Equal(t, "https://www.google.com/s2/favicons?domain=1password.com&sz=128", mapped["logo_remote_url"])
	assert.Equal(t, true, mapped["externally_visible"])
	assert.Assert(t, mapped["description"] != nil)
	assert.Assert(t, mapped["aliases"] == nil)

	mapped = mappingtest.EvalMap(t, spec, types.MappingEnvelope{
		Resource: "vendors/minimal.json#0",
		Payload:  json.RawMessage(`{"systemInternalID":"entity::minimal","name":"minimal"}`),
	})

	assert.Equal(t, "entity::minimal", mapped["external_id"])
	assert.Equal(t, "minimal", mapped["name"])
	assert.Assert(t, mapped["display_name"] == nil)
	assert.Assert(t, mapped["description"] == nil)
	assert.Assert(t, mapped["domains"] == nil)
	assert.Assert(t, mapped["aliases"] == nil)
	assert.Assert(t, mapped["tags"] == nil)
	assert.Assert(t, mapped["logo_remote_url"] == nil)
	assert.Equal(t, false, mapped["externally_visible"])

	mapped = mappingtest.EvalMap(t, spec, types.MappingEnvelope{
		Resource: "vendors/aliased.json#0",
		Payload:  json.RawMessage(`{"systemInternalID":"entity::aliased","name":"aliased","aliases":["alias-one","alias-two"],"externallyVisible":true}`),
	})

	assert.DeepEqual(t, []any{"alias-one", "alias-two"}, mapped["aliases"])
	assert.Equal(t, true, mapped["externally_visible"])
}

func TestEntityMappingFilter(t *testing.T) {
	spec := vendorMappingSpec(t)

	vendor := types.MappingEnvelope{Payload: mappingtest.LoadExample(t, "examples", "adobe.json")}
	assert.Assert(t, mappingtest.AssertFiltered(t, spec, vendor), "expected a vendor record to pass the filter")

	minimal := types.MappingEnvelope{Payload: json.RawMessage(`{"systemInternalID":"entity::u","name":"u"}`)}
	assert.Assert(t, mappingtest.AssertFiltered(t, spec, minimal), "the Entity mapping filters nothing")
}

func TestEntityMappingLinkExpression(t *testing.T) {
	spec := vendorMappingSpec(t)
	assert.Equal(t, mapExprEntity, spec.MapExpr)
	assert.Equal(t, 1, len(spec.Links))
	assert.Equal(t, entityops.SchemaEntityType.Name, spec.Links[0].TargetSchema)

	envCfg := celx.StrictEnvConfig()
	envCfg.CrossTypeNumericComparisons = true

	eval, err := celx.NewNativeEntityEvaluator(envCfg, celx.FastEvalConfig(), reflect.TypeFor[entityops.EntityTypeProjection](), reflect.TypeFor[entityops.EntityProjection]())
	assert.NilError(t, err)

	vendorType := json.RawMessage(`{"id":"et_01","name":"vendor"}`)
	customerType := json.RawMessage(`{"id":"et_02","name":"customer"}`)
	source := json.RawMessage(`{"external_id":"vendor:okta","name":"Okta"}`)

	matched, err := eval.EvaluateBoolWithSource(context.Background(), spec.Links[0].Expression, vendorType, source)
	assert.NilError(t, err)
	assert.Assert(t, matched, "expected the vendor entity type to match")

	matched, err = eval.EvaluateBoolWithSource(context.Background(), spec.Links[0].Expression, customerType, source)
	assert.NilError(t, err)
	assert.Assert(t, !matched, "expected the customer entity type not to match")
}
