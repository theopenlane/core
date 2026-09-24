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

func TestMappingExpressionsValid(t *testing.T) {
	def, err := Builder(&RuntimeConfig{}, "", Config{})()
	assert.NilError(t, err)

	for _, m := range def.Mappings {
		name := m.Schema
		if m.Variant != "" {
			name += "/" + m.Variant
		}

		t.Run(name+"/filter", func(t *testing.T) {
			assert.NilError(t, providerkit.ValidateExpr(m.Spec.FilterExpr))
		})

		t.Run(name+"/map", func(t *testing.T) {
			assert.NilError(t, providerkit.ValidateExpr(m.Spec.MapExpr))
		})
	}
}

func TestEntityMapping(t *testing.T) {
	def, err := Builder(&RuntimeConfig{}, "", Config{})()
	assert.NilError(t, err)

	spec := mappingtest.MappingSpec(t, def.Mappings, "Entity")

	envelope := types.MappingEnvelope{
		Resource: "vendors/1password-vendor.json#0",
		Payload:  mappingtest.LoadExample(t, "examples", "1password-vendor.json"),
	}

	assert.Assert(t, mappingtest.AssertFiltered(t, spec, envelope), "expected the vendor record to pass the Entity filter")

	mapped := mappingtest.EvalMap(t, spec, envelope)

	assert.Equal(t, "entity::MI2d4zsM-AnFxqsI6tNvFg", mapped["external_id"])
	assert.Equal(t, "entity::MI2d4zsM-AnFxqsI6tNvFg", mapped["system_internal_id"])
	assert.Equal(t, "1password-vendor", mapped["name"])
	assert.Equal(t, "1Password", mapped["display_name"])
	assert.DeepEqual(t, []any{"1password.com"}, mapped["domains"])
	assert.DeepEqual(t, []any{"technology", "password management"}, mapped["tags"])
	assert.Equal(t, "https://www.google.com/s2/favicons?domain=1password.com&sz=128", mapped["logo_remote_url"])
	assert.Equal(t, false, mapped["externally_visible"])
	assert.Assert(t, mapped["description"] != nil)

	mapped = mappingtest.EvalMap(t, spec, types.MappingEnvelope{
		Resource: "vendors/minimal.json#0",
		Payload:  json.RawMessage(`{"systemInternalID":"entity::minimal","name":"minimal"}`),
	})

	assert.Equal(t, "entity::minimal", mapped["external_id"])
	assert.Equal(t, "minimal", mapped["name"])
	assert.Assert(t, mapped["display_name"] == nil)
	assert.Assert(t, mapped["description"] == nil)
	assert.Assert(t, mapped["domains"] == nil)
	assert.Assert(t, mapped["tags"] == nil)
	assert.Assert(t, mapped["logo_remote_url"] == nil)
	assert.Equal(t, false, mapped["externally_visible"])

	mapped = mappingtest.EvalMap(t, spec, types.MappingEnvelope{
		Resource: "vendors/visible.json#0",
		Payload:  json.RawMessage(`{"systemInternalID":"entity::visible","name":"visible","externallyVisible":true}`),
	})

	assert.Equal(t, true, mapped["externally_visible"])
}

func TestVendorVariantFilter(t *testing.T) {
	def, err := Builder(&RuntimeConfig{}, "", Config{})()
	assert.NilError(t, err)

	variant, found := lo.Find(def.Mappings, func(m types.MappingRegistration) bool {
		return m.Schema == entityops.SchemaEntity.Name && m.Variant == variantVendor
	})
	assert.Assert(t, found, "expected a vendor mapping variant for Entity")

	vendor := types.MappingEnvelope{Variant: variantVendor, Payload: mappingtest.LoadExample(t, "examples", "adobe.json")}
	assert.Assert(t, mappingtest.AssertFiltered(t, variant.Spec, vendor), "expected a vendor-typed record to pass the vendor filter")

	customer := types.MappingEnvelope{Variant: variantVendor, Payload: json.RawMessage(`{"systemInternalID":"entity::c","name":"c","entityTypeName":"customer"}`)}
	assert.Assert(t, !mappingtest.AssertFiltered(t, variant.Spec, customer), "expected a non-vendor record to be filtered out")

	untyped := types.MappingEnvelope{Variant: variantVendor, Payload: json.RawMessage(`{"systemInternalID":"entity::u","name":"u"}`)}
	assert.Assert(t, !mappingtest.AssertFiltered(t, variant.Spec, untyped), "expected an untyped record to be filtered out")
}

func TestVendorVariantLinkExpression(t *testing.T) {
	def, err := Builder(&RuntimeConfig{}, "", Config{})()
	assert.NilError(t, err)

	variant, found := lo.Find(def.Mappings, func(m types.MappingRegistration) bool {
		return m.Schema == entityops.SchemaEntity.Name && m.Variant == variantVendor
	})
	assert.Assert(t, found, "expected a vendor mapping variant for Entity")
	assert.Equal(t, mapExprEntity, variant.Spec.MapExpr)
	assert.Equal(t, 1, len(variant.Spec.Links))
	assert.Equal(t, entityops.SchemaEntityType.Name, variant.Spec.Links[0].TargetSchema)

	envCfg := celx.StrictEnvConfig()
	envCfg.CrossTypeNumericComparisons = true

	eval, err := celx.NewNativeEntityEvaluator(envCfg, celx.FastEvalConfig(), reflect.TypeFor[entityops.EntityTypeProjection](), reflect.TypeFor[entityops.EntityProjection]())
	assert.NilError(t, err)

	vendorType := json.RawMessage(`{"id":"et_01","name":"vendor"}`)
	customerType := json.RawMessage(`{"id":"et_02","name":"customer"}`)
	source := json.RawMessage(`{"external_id":"vendor:okta","name":"Okta"}`)

	matched, err := eval.EvaluateBoolWithSource(context.Background(), variant.Spec.Links[0].Expression, vendorType, source)
	assert.NilError(t, err)
	assert.Assert(t, matched, "expected the vendor entity type to match")

	matched, err = eval.EvaluateBoolWithSource(context.Background(), variant.Spec.Links[0].Expression, customerType, source)
	assert.NilError(t, err)
	assert.Assert(t, !matched, "expected the customer entity type not to match")
}
