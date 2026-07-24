package slack

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	slackgo "github.com/slack-go/slack"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// testMappings builds the Slack definition and returns its registered mappings
func testMappings(t *testing.T) []types.MappingRegistration {
	t.Helper()

	def, err := Builder(Config{}, nil, false)()
	assert.NilError(t, err)

	return def.Mappings
}

func TestMappingExpressionsValid(t *testing.T) {
	for _, m := range testMappings(t) {
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

// TestSlackMappingsUserExample tests the directory account mapping against examples/user.json
func TestSlackMappingsUserExample(t *testing.T) {
	raw, err := os.ReadFile("examples/user.json")
	assert.NilError(t, err)

	var outer struct {
		User slackgo.User `json:"user"`
	}
	assert.NilError(t, json.Unmarshal(raw, &outer))

	payload := normalizeUser(outer.User)
	rawPayload, err := json.Marshal(payload)
	assert.NilError(t, err)

	envelope := providerkit.RawEnvelope(outer.User.TeamID+"/"+outer.User.ID, rawPayload)
	result, err := providerkit.EvalMap(context.Background(), testMappings(t)[0].Spec.MapExpr, envelope)
	assert.NilError(t, err)

	mapped, err := jsonx.ToMap(result)
	assert.NilError(t, err)

	assert.Equal(t, "U123ABC456", mapped["external_id"])
	assert.Equal(t, "sholmes@example.com", mapped["canonical_email"])
	assert.Equal(t, "sherlock", mapped["display_name"])
	assert.Equal(t, "Sherlock", mapped["given_name"])
	assert.Equal(t, "Holmes", mapped["family_name"])
	assert.Equal(t, "Senior Detective", mapped["job_title"])
	assert.Equal(t, "DISABLED", mapped["mfa_state"])
	assert.Equal(t, "ACTIVE", mapped["status"])
	assert.Equal(t, "T123ABC456", mapped["directory_instance_id"])
	assert.Equal(t, "USER", mapped["account_type"])
}

// TestSlackMappingsServiceExample tests the directory account mapping against examples/service.json
func TestSlackMappingsServiceExample(t *testing.T) {
	raw, err := os.ReadFile("examples/service.json")
	assert.NilError(t, err)

	var outer struct {
		User slackgo.User `json:"user"`
	}
	assert.NilError(t, json.Unmarshal(raw, &outer))

	payload := normalizeUser(outer.User)
	rawPayload, err := json.Marshal(payload)
	assert.NilError(t, err)

	envelope := providerkit.RawEnvelope(outer.User.TeamID+"/"+outer.User.ID, rawPayload)
	result, err := providerkit.EvalMap(context.Background(), testMappings(t)[0].Spec.MapExpr, envelope)
	assert.NilError(t, err)

	mapped, err := jsonx.ToMap(result)
	assert.NilError(t, err)

	assert.Equal(t, "U123ABC456", mapped["external_id"])
	assert.Equal(t, "sholmes@example.com", mapped["canonical_email"])
	assert.Equal(t, "sherlock", mapped["display_name"])
	assert.Equal(t, "ACTIVE", mapped["status"])
	assert.Equal(t, "T123ABC456", mapped["directory_instance_id"])
	assert.Equal(t, "SERVICE", mapped["account_type"])
}
