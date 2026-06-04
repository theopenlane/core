package keycloak

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/internal/integrations/mappingtest"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestMappingExpressionsValid(t *testing.T) {
	for _, m := range keycloakMappings() {
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
	mappings := keycloakMappings()
	accountSpec := mappingtest.MappingSpec(t, mappings, "DirectoryAccount")
	groupSpec := mappingtest.MappingSpec(t, mappings, "DirectoryGroup")
	membershipSpec := mappingtest.MappingSpec(t, mappings, "DirectoryMembership")

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
	})

	t.Run("service_account_user_json", func(t *testing.T) {
		envelope := types.MappingEnvelope{
			Payload: mappingtest.LoadExample(t, "examples", "service_account_user.json"),
		}

		assert.Assert(t, mappingtest.AssertFiltered(t, accountSpec, envelope), "expected service_account_user.json to pass the DirectoryAccount filter")

		mapped := mappingtest.EvalMap(t, accountSpec, envelope)

		assert.Equal(t, "19cac1ac-19e6-43ed-bf3b-90ced88d3548", mapped["externalID"])
		assert.Equal(t, "SERVICE", mapped["accountType"])
		assert.Equal(t, "ACTIVE", mapped["status"])
	})

	t.Run("group_json", func(t *testing.T) {
		envelope := types.MappingEnvelope{
			Payload: mappingtest.LoadExample(t, "examples", "group.json"),
		}

		assert.Assert(t, mappingtest.AssertFiltered(t, groupSpec, envelope), "expected group.json to pass the DirectoryGroup filter")

		mapped := mappingtest.EvalMap(t, groupSpec, envelope)

		assert.Equal(t, "488f8a0c-c0b6-4dce-bf79-82db3af7cdac", mapped["externalID"])
		assert.Equal(t, "Engineering", mapped["displayName"])
		assert.Equal(t, "ACTIVE", mapped["status"])
	})

	t.Run("member_json", func(t *testing.T) {
		groupID := "488f8a0c-c0b6-4dce-bf79-82db3af7cdac"

		envelope := types.MappingEnvelope{
			Payload:  mappingtest.LoadExample(t, "examples", "member.json"),
			Resource: groupID,
		}

		assert.Assert(t, mappingtest.AssertFiltered(t, membershipSpec, envelope), "expected member.json to pass the DirectoryMembership filter")

		mapped := mappingtest.EvalMap(t, membershipSpec, envelope)

		assert.Equal(t, "8413c0cc-69a7-4726-a922-63e892264adc", mapped["directoryAccountID"])
		assert.Equal(t, "488f8a0c-c0b6-4dce-bf79-82db3af7cdac", mapped["directoryGroupID"])
		assert.Equal(t, "MEMBER", mapped["role"])
	})

	t.Run("inactive_user", func(t *testing.T) {
		payload, err := json.Marshal(map[string]any{
			"id":       "abc-123",
			"username": "inactive-user",
			"enabled":  false,
			"email":    "inactive@example.com",
		})
		assert.NilError(t, err)

		envelope := types.MappingEnvelope{Payload: json.RawMessage(payload)}
		mapped := mappingtest.EvalMap(t, accountSpec, envelope)

		assert.Equal(t, "INACTIVE", mapped["status"])
	})

	t.Run("user_no_name_falls_back_to_username", func(t *testing.T) {
		payload, err := json.Marshal(map[string]any{
			"id":       "def-456",
			"username": "janedoe",
			"enabled":  true,
			"email":    "jane@example.com",
		})
		assert.NilError(t, err)

		envelope := types.MappingEnvelope{Payload: json.RawMessage(payload)}
		mapped := mappingtest.EvalMap(t, accountSpec, envelope)

		assert.Equal(t, "janedoe", mapped["displayName"])
		assert.Equal(t, "USER", mapped["accountType"])
	})
}
