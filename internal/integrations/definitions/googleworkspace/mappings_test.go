package googleworkspace

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// TestGoogleWorkspaceMappingsEvalMap verifies Google Workspace payloads map into directory schemas
func TestGoogleWorkspaceMappingsEvalMap(t *testing.T) {
	accountRaw, err := providerkit.EvalMap(context.Background(), mappingSpecForSchema(t, integrationgenerated.IntegrationMappingSchemaDirectoryAccount).MapExpr, types.MappingEnvelope{
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

	assert.Equal(t, "user-123", accountMapped["externalID"])
	assert.Equal(t, "alice@example.com", accountMapped["canonicalEmail"])
	assert.Equal(t, "Alice Example", accountMapped["displayName"])
	assert.Equal(t, "Alice", accountMapped["givenName"])
	assert.Equal(t, "Example", accountMapped["familyName"])
	assert.Equal(t, "/Engineering", accountMapped["organizationUnit"])
	assert.Equal(t, "ACTIVE", accountMapped["status"])
	assert.Equal(t, "ENFORCED", accountMapped["mfaState"])

	groupRaw, err := providerkit.EvalMap(context.Background(), mappingSpecForSchema(t, integrationgenerated.IntegrationMappingSchemaDirectoryGroup).MapExpr, types.MappingEnvelope{
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

	assert.Equal(t, "group-123", groupMapped["externalID"])
	assert.Equal(t, "eng@example.com", groupMapped["email"])
	assert.Equal(t, "Engineering", groupMapped["displayName"])
	assert.Equal(t, "DISTRIBUTION", groupMapped["classification"])
	assert.Equal(t, "ACTIVE", groupMapped["status"])

	membershipRaw, err := providerkit.EvalMap(context.Background(), mappingSpecForSchema(t, integrationgenerated.IntegrationMappingSchemaDirectoryMembership).MapExpr, types.MappingEnvelope{
		Resource: "eng@example.com:alice@example.com",
		Payload: json.RawMessage(`{
			"group":{"id":"group-123","email":"eng@example.com"},
			"member":{"id":"user-123","email":"alice@example.com","role":"OWNER","type":"USER"}
		}`),
	})
	require.NoError(t, err)

	membershipMapped, err := jsonx.ToMap(membershipRaw)
	require.NoError(t, err)

	assert.Equal(t, "user-123", membershipMapped["directoryAccountID"])
	assert.Equal(t, "group-123", membershipMapped["directoryGroupID"])
	assert.Equal(t, "OWNER", membershipMapped["role"])
	assert.Equal(t, "google_workspace", membershipMapped["source"])
}

// mappingSpecForSchema returns the mapping override for one schema from the Google Workspace defaults
func mappingSpecForSchema(t *testing.T, schema string) types.MappingOverride {
	t.Helper()

	for _, mapping := range googleWorkspaceMappings() {
		if mapping.Schema == schema {
			return mapping.Spec
		}
	}

	t.Fatalf("mapping not found for schema %s", schema)

	return types.MappingOverride{}
}
