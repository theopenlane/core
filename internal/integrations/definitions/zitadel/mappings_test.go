package zitadel

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/internal/integrations/mappingtest"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestMappingExpressionsValid(t *testing.T) {
	for _, m := range zitadelMappings() {
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

func TestExamplePayloads(t *testing.T) {
	mappings := zitadelMappings()
	accountSpec := mappingtest.MappingSpec(t, mappings, "DirectoryAccount")

	t.Run("user_json", func(t *testing.T) {
		envelope := types.MappingEnvelope{
			Payload: mappingtest.LoadExample(t, "examples", "user.json"),
		}

		assert.Assert(t, mappingtest.AssertFiltered(t, accountSpec, envelope), "expected user.json to pass the DirectoryAccount filter")

		mapped := mappingtest.EvalMap(t, accountSpec, envelope)

		assert.Equal(t, "8413c0cc-69a7-4726-a922-63e892264adc", mapped["externalID"])
		assert.Equal(t, "john.doe@company.com", mapped["canonicalEmail"])
		assert.Equal(t, "John Doe", mapped["displayName"])
		assert.Equal(t, "John", mapped["givenName"])
		assert.Equal(t, "Doe", mapped["familyName"])
		assert.Equal(t, "ACTIVE", mapped["status"])
		assert.Equal(t, "USER", mapped["accountType"])
		assert.Equal(t, "2026-01-15T10:00:00Z", mapped["addedAt"])
	})

	t.Run("machine_user_json", func(t *testing.T) {
		envelope := types.MappingEnvelope{
			Payload: mappingtest.LoadExample(t, "examples", "machine_user.json"),
		}

		assert.Assert(t, mappingtest.AssertFiltered(t, accountSpec, envelope), "expected machine_user.json to pass the DirectoryAccount filter")

		mapped := mappingtest.EvalMap(t, accountSpec, envelope)

		assert.Equal(t, "19cac1ac-19e6-43ed-bf3b-90ced88d3548", mapped["externalID"])
		assert.Equal(t, "SERVICE", mapped["accountType"])
		assert.Equal(t, "ACTIVE", mapped["status"])
	})

	t.Run("inactive_user", func(t *testing.T) {
		payload, err := json.Marshal(map[string]any{
			"user_id":  "abc-123",
			"username": "inactive-user",
			"state":    2,
			"profile": map[string]any{
				"given_name":  "Inactive",
				"family_name": "User",
			},
		})
		assert.NilError(t, err)

		envelope := types.MappingEnvelope{Payload: json.RawMessage(payload)}
		mapped := mappingtest.EvalMap(t, accountSpec, envelope)

		assert.Equal(t, "INACTIVE", mapped["status"])
	})

	t.Run("locked_user", func(t *testing.T) {
		payload, err := json.Marshal(map[string]any{
			"user_id":  "def-456",
			"username": "locked-user",
			"state":    4,
			"profile": map[string]any{
				"given_name":  "Locked",
				"family_name": "User",
			},
		})
		assert.NilError(t, err)

		envelope := types.MappingEnvelope{Payload: json.RawMessage(payload)}
		mapped := mappingtest.EvalMap(t, accountSpec, envelope)

		assert.Equal(t, "SUSPENDED", mapped["status"])
	})

	t.Run("user_fallback_to_username", func(t *testing.T) {
		payload, err := json.Marshal(map[string]any{
			"user_id":  "ghi-789",
			"username": "janedoe",
			"state":    1,
		})
		assert.NilError(t, err)

		envelope := types.MappingEnvelope{Payload: json.RawMessage(payload)}
		mapped := mappingtest.EvalMap(t, accountSpec, envelope)

		assert.Equal(t, "janedoe", mapped["displayName"])
		assert.Equal(t, "SERVICE", mapped["accountType"])
	})
}