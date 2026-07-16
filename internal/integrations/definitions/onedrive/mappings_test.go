package onedrive

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/internal/integrations/mappingtest"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestMappingExpressionsValid(t *testing.T) {
	for _, m := range oneDriveMappings() {
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

func TestInternalPolicyMappingWithWebURL(t *testing.T) {
	spec := mappingtest.MappingSpec(t, oneDriveMappings(), "InternalPolicy")

	payload := mappingtest.LoadExample(t, "examples", "document_with_weburl.json")

	envelope := types.MappingEnvelope{
		Payload: payload,
	}

	assert.Assert(t, mappingtest.AssertFiltered(t, spec, envelope), "expected document_with_weburl.json to pass the InternalPolicy filter")

	mapped := mappingtest.EvalMap(t, spec, envelope)

	assert.Equal(t, "Security Policy.docx", mapped["name"])
	assert.Equal(t, "01BYE5RZ6QN3ZWBTUFOFD3GSPGOHDJD36K", mapped["external_file_id"])
	assert.Equal(t, "https://contoso.sharepoint.com/sites/policies/Shared%20Documents/Security%20Policy.docx", mapped["url"])
	assert.Equal(t, "INTEGRATION", mapped["management_mode"])
	assert.Equal(t, "DRAFT", mapped["status"])
}

func TestInternalPolicyMappingWithoutWebURL(t *testing.T) {
	spec := mappingtest.MappingSpec(t, oneDriveMappings(), "InternalPolicy")

	payload := mappingtest.LoadExample(t, "examples", "document_without_weburl.json")

	envelope := types.MappingEnvelope{
		Payload: payload,
	}

	assert.Assert(t, mappingtest.AssertFiltered(t, spec, envelope), "expected document_without_weburl.json to pass the InternalPolicy filter")

	mapped := mappingtest.EvalMap(t, spec, envelope)

	assert.Equal(t, "Security Policy.docx", mapped["name"])
	assert.Equal(t, "01BYE5RZ6QN3ZWBTUFOFD3GSPGOHDJD36K", mapped["external_file_id"])
	assert.Equal(t, nil, mapped["url"])
	assert.Equal(t, "INTEGRATION", mapped["management_mode"])
	assert.Equal(t, "DRAFT", mapped["status"])
}
