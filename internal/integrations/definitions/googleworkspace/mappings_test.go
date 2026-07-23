package googleworkspace

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// testMappings returns the ingest mappings from the built Google Workspace definition
func testMappings(t *testing.T) []types.MappingRegistration {
	t.Helper()

	def, err := Builder(Config{})()
	require.NoError(t, err)

	return def.Mappings
}

func TestMappingExpressionsValid(t *testing.T) {
	for _, m := range testMappings(t) {
		name := m.Schema
		if m.Variant != "" {
			name += "/" + m.Variant
		}

		t.Run(name+"/filter", func(t *testing.T) {
			assert.NoError(t, providerkit.ValidateExpr(m.Spec.FilterExpr))
		})

		t.Run(name+"/map", func(t *testing.T) {
			assert.NoError(t, providerkit.ValidateExpr(m.Spec.MapExpr))
		})
	}
}

// TestGoogleWorkspaceMappingsEvalMap verifies Google Workspace payloads map into directory schemas
func TestGoogleWorkspaceMappingsEvalMap(t *testing.T) {
	accountRaw, err := providerkit.EvalMap(context.Background(), mappingSpecForSchema(t, entityops.SchemaDirectoryAccount.Name).MapExpr, types.MappingEnvelope{
		Resource: "alice@example.com",
		Payload: json.RawMessage(`{
			"id":"user-123",
			"primaryEmail":"alice@example.com",
			"name":{"fullName":"Alice Example","givenName":"Alice","familyName":"Example"},
			"orgUnitPath":"/Engineering",
			"suspended":false,
			"archived":false,
			"isEnforcedIn2Sv":true,
			"isEnrolledIn2Sv":true,
			"lastLoginTime":"2026-03-15T10:00:00Z",
			"customerId":"C123"
		}`),
	})
	require.NoError(t, err)

	accountMapped, err := jsonx.ToMap(accountRaw)
	require.NoError(t, err)

	assert.Equal(t, "user-123", accountMapped["external_id"])
	assert.Equal(t, "alice@example.com", accountMapped["canonical_email"])
	assert.Equal(t, "Alice Example", accountMapped["display_name"])
	assert.Equal(t, "Alice", accountMapped["given_name"])
	assert.Equal(t, "Example", accountMapped["family_name"])
	assert.Equal(t, "/Engineering", accountMapped["organization_unit"])
	assert.Equal(t, "ACTIVE", accountMapped["status"])
	assert.Equal(t, "ENFORCED", accountMapped["mfa_state"])

	groupRaw, err := providerkit.EvalMap(context.Background(), mappingSpecForSchema(t, entityops.SchemaDirectoryGroup.Name).MapExpr, types.MappingEnvelope{
		Resource: "eng@example.com",
		Payload: json.RawMessage(`{
			"id":"group-123",
			"email":"eng@example.com",
			"name":"Engineering",
			"adminCreated":false,
			"etag":"group-etag"
		}`),
	})
	require.NoError(t, err)

	groupMapped, err := jsonx.ToMap(groupRaw)
	require.NoError(t, err)

	assert.Equal(t, "group-123", groupMapped["external_id"])
	assert.Equal(t, "eng@example.com", groupMapped["email"])
	assert.Equal(t, "Engineering", groupMapped["display_name"])
	assert.Equal(t, "DISTRIBUTION", groupMapped["classification"])
	assert.Equal(t, "ACTIVE", groupMapped["status"])

	membershipRaw, err := providerkit.EvalMap(context.Background(), mappingSpecForSchema(t, entityops.SchemaDirectoryMembership.Name).MapExpr, types.MappingEnvelope{
		Resource: "eng@example.com",
		Payload: json.RawMessage(`{
			"email":"alice@example.com",
			"role":"OWNER",
			"type":"USER",
			"id":"member-123"
		}`),
	})
	require.NoError(t, err)

	membershipMapped, err := jsonx.ToMap(membershipRaw)
	require.NoError(t, err)

	assert.Equal(t, "alice@example.com", membershipMapped["directory_account_id"])
	assert.Equal(t, "eng@example.com", membershipMapped["directory_group_id"])
	assert.Equal(t, "OWNER", membershipMapped["role"])
}

// TestGoogleWorkspaceMappingsFallbacks verifies graceful fallback when fields are missing from the payload
func TestGoogleWorkspaceMappingsFallbacks(t *testing.T) {
	accountRaw, err := providerkit.EvalMap(context.Background(), mappingSpecForSchema(t, entityops.SchemaDirectoryAccount.Name).MapExpr, types.MappingEnvelope{
		Resource: "sparse@example.com",
		Payload:  json.RawMessage(`{"id":"user-sparse","primaryEmail":"sparse@example.com"}`),
	})
	require.NoError(t, err)

	accountMapped, err := jsonx.ToMap(accountRaw)
	require.NoError(t, err)

	assert.Equal(t, "user-sparse", accountMapped["external_id"])
	assert.Equal(t, "sparse@example.com", accountMapped["canonical_email"])
	assert.Equal(t, "", accountMapped["display_name"])
	assert.Equal(t, "", accountMapped["given_name"])
	assert.Equal(t, "", accountMapped["family_name"])
	assert.Equal(t, "", accountMapped["organization_unit"])
	assert.Equal(t, "ACTIVE", accountMapped["status"])
	assert.Equal(t, "DISABLED", accountMapped["mfa_state"])

	membershipRaw, err := providerkit.EvalMap(context.Background(), mappingSpecForSchema(t, entityops.SchemaDirectoryMembership.Name).MapExpr, types.MappingEnvelope{
		Resource: "eng@example.com",
		Payload:  json.RawMessage(`{"email":"norole@example.com","type":"USER","id":"member-norole"}`),
	})
	require.NoError(t, err)

	membershipMapped, err := jsonx.ToMap(membershipRaw)
	require.NoError(t, err)

	assert.Equal(t, "MEMBER", membershipMapped["role"], "missing role must fall back to MEMBER so the row survives enum validation")

	emptyRoleRaw, err := providerkit.EvalMap(context.Background(), mappingSpecForSchema(t, entityops.SchemaDirectoryMembership.Name).MapExpr, types.MappingEnvelope{
		Resource: "eng@example.com",
		Payload:  json.RawMessage(`{"email":"emptyrole@example.com","role":"","type":"USER","id":"member-emptyrole"}`),
	})
	require.NoError(t, err)

	emptyRoleMapped, err := jsonx.ToMap(emptyRoleRaw)
	require.NoError(t, err)

	assert.Equal(t, "MEMBER", emptyRoleMapped["role"], "empty role must fall back to MEMBER so the row survives enum validation")
}

// mappingSpecForSchema returns the mapping override for one schema from the Google Workspace defaults
func mappingSpecForSchema(t *testing.T, schema string) types.MappingOverride {
	t.Helper()

	for _, mapping := range testMappings(t) {
		if mapping.Schema == schema {
			return mapping.Spec
		}
	}

	t.Fatalf("mapping not found for schema %s", schema)

	return types.MappingOverride{}
}
