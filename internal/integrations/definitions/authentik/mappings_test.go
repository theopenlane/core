package authentik

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/internal/integrations/mappingtest"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestMappingExpressionsValid(t *testing.T) {
	for _, m := range authentikMappings() {
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
	mappings := authentikMappings()
	accountSpec := mappingtest.MappingSpec(t, mappings, "DirectoryAccount")
	groupSpec := mappingtest.MappingSpec(t, mappings, "DirectoryGroup")
	membershipSpec := mappingtest.MappingSpec(t, mappings, "DirectoryMembership")

	t.Run("user_json", func(t *testing.T) {
		envelope := types.MappingEnvelope{
			Payload: mappingtest.LoadExample(t, "examples", "user.json"),
		}

		assert.Assert(t, mappingtest.AssertFiltered(t, accountSpec, envelope), "expected user.json to pass the DirectoryAccount filter")

		mapped := mappingtest.EvalMap(t, accountSpec, envelope)

		assert.Equal(t, "772416add93e0b563b4ceb5c534f9b064f356bb2782d61d31bf1dc2c7cf1e3a5", mapped["external_id"])
		assert.Equal(t, "authentik Default Admin", mapped["display_name"])
		assert.Equal(t, "manab@gmail.com", mapped["canonical_email"])
		assert.Equal(t, "ACTIVE", mapped["status"])
		assert.Equal(t, "USER", mapped["account_type"])
		assert.Equal(t, "2026-05-08T01:54:15.179825Z", mapped["added_at"])
		assert.Equal(t, "2026-05-09T08:11:37.941933Z", mapped["last_seen_at"])
		assert.Equal(t, "2026-05-08T01:55:21.345631Z", mapped["observed_at"])
	})

	t.Run("group_json", func(t *testing.T) {
		envelope := types.MappingEnvelope{
			Payload: mappingtest.LoadExample(t, "examples", "group.json"),
		}

		assert.Assert(t, mappingtest.AssertFiltered(t, groupSpec, envelope), "expected group.json to pass the DirectoryGroup filter")

		mapped := mappingtest.EvalMap(t, groupSpec, envelope)

		assert.Equal(t, "488f8a0c-c0b6-4dce-bf79-82db3af7cdac", mapped["external_id"])
		assert.Equal(t, "authentik Admins", mapped["display_name"])
		assert.Equal(t, "ACTIVE", mapped["status"])
	})

	t.Run("member_json", func(t *testing.T) {
		groupPK := "488f8a0c-c0b6-4dce-bf79-82db3af7cdac"

		payload := mappingtest.LoadExample(t, "examples", "member.json")

		envelope := types.MappingEnvelope{
			Payload:  payload,
			Resource: groupPK,
		}

		assert.Assert(t, mappingtest.AssertFiltered(t, membershipSpec, envelope), "expected member.json to pass the DirectoryMembership filter")

		mapped := mappingtest.EvalMap(t, membershipSpec, envelope)

		assert.Equal(t, "772416add93e0b563b4ceb5c534f9b064f356bb2782d61d31bf1dc2c7cf1e3a5", mapped["directory_account_id"])
		assert.Equal(t, "488f8a0c-c0b6-4dce-bf79-82db3af7cdac", mapped["directory_group_id"])
		assert.Equal(t, "MEMBER", mapped["role"])
	})

	t.Run("inactive_user", func(t *testing.T) {
		payload, err := json.Marshal(map[string]any{
			"pk":           99,
			"username":     "inactive-user",
			"name":         "Inactive User",
			"is_active":    false,
			"email":        "inactive@example.com",
			"uid":          "abc123",
			"type":         "internal",
			"date_joined":  "2026-01-01T00:00:00Z",
			"last_updated": "2026-01-01T00:00:00Z",
			"attributes":   map[string]any{},
		})
		assert.NilError(t, err)

		envelope := types.MappingEnvelope{Payload: json.RawMessage(payload)}
		mapped := mappingtest.EvalMap(t, accountSpec, envelope)

		assert.Equal(t, "INACTIVE", mapped["status"])
	})

	t.Run("service_account_user", func(t *testing.T) {
		payload, err := json.Marshal(map[string]any{
			"pk":           5,
			"username":     "openlane-svc",
			"name":         "Open Lane Service",
			"is_active":    true,
			"email":        "svc@example.com",
			"uid":          "svcuid123",
			"type":         "service_account",
			"date_joined":  "2026-01-01T00:00:00Z",
			"last_updated": "2026-01-01T00:00:00Z",
			"attributes":   map[string]any{},
		})
		assert.NilError(t, err)

		envelope := types.MappingEnvelope{Payload: json.RawMessage(payload)}
		mapped := mappingtest.EvalMap(t, accountSpec, envelope)

		assert.Equal(t, "SERVICE", mapped["account_type"])
		assert.Equal(t, "ACTIVE", mapped["status"])
	})
}
