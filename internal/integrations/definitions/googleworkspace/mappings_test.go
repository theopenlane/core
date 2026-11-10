package googleworkspace

import (
	"context"
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/mappingtest"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/utils/jsonx"
)

// testMappings returns the ingest mappings from the built Google Workspace definition
func testMappings(t *testing.T) []types.MappingRegistration {
	t.Helper()

	def, err := Builder(Config{})()
	assert.NilError(t, err)

	return def.Mappings
}

func TestMappingExpressionsValid(t *testing.T) {
	mappingtest.AssertExpressionsValid(t, testMappings(t))
}

// TestGoogleWorkspaceMappingsEvalMap verifies Google Workspace payloads map into directory schemas
func TestGoogleWorkspaceMappingsEvalMap(t *testing.T) {
	accountRaw, err := providerkit.EvalMap(context.Background(), mappingtest.MappingSpec(t, testMappings(t), entityops.SchemaDirectoryAccount.Name).MapExpr, types.MappingEnvelope{
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
	}, types.MappingInstallation{})
	assert.NilError(t, err)

	accountMapped, err := jsonx.ToMap(accountRaw)
	assert.NilError(t, err)

	assert.Check(t, is.Equal("user-123", accountMapped["external_id"]))
	assert.Check(t, is.Equal("alice@example.com", accountMapped["canonical_email"]))
	assert.Check(t, is.Equal("Alice Example", accountMapped["display_name"]))
	assert.Check(t, is.Equal("Alice", accountMapped["given_name"]))
	assert.Check(t, is.Equal("Example", accountMapped["family_name"]))
	assert.Check(t, is.Equal("/Engineering", accountMapped["organization_unit"]))
	assert.Check(t, is.Equal("ACTIVE", accountMapped["status"]))
	assert.Check(t, is.Equal("ENFORCED", accountMapped["mfa_state"]))

	groupRaw, err := providerkit.EvalMap(context.Background(), mappingtest.MappingSpec(t, testMappings(t), entityops.SchemaDirectoryGroup.Name).MapExpr, types.MappingEnvelope{
		Resource: "eng@example.com",
		Payload: json.RawMessage(`{
			"id":"group-123",
			"email":"eng@example.com",
			"name":"Engineering",
			"adminCreated":false,
			"etag":"group-etag"
		}`),
	}, types.MappingInstallation{})
	assert.NilError(t, err)

	groupMapped, err := jsonx.ToMap(groupRaw)
	assert.NilError(t, err)

	assert.Check(t, is.Equal("group-123", groupMapped["external_id"]))
	assert.Check(t, is.Equal("eng@example.com", groupMapped["email"]))
	assert.Check(t, is.Equal("Engineering", groupMapped["display_name"]))
	assert.Check(t, is.Equal("DISTRIBUTION", groupMapped["classification"]))
	assert.Check(t, is.Equal("ACTIVE", groupMapped["status"]))

	membershipRaw, err := providerkit.EvalMap(context.Background(), mappingtest.MappingSpec(t, testMappings(t), entityops.SchemaDirectoryMembership.Name).MapExpr, types.MappingEnvelope{
		Resource: "group-123",
		Payload: json.RawMessage(`{
			"email":"alice@example.com",
			"role":"OWNER",
			"type":"USER",
			"id":"member-123"
		}`),
	}, types.MappingInstallation{})
	assert.NilError(t, err)

	membershipMapped, err := jsonx.ToMap(membershipRaw)
	assert.NilError(t, err)

	assert.Check(t, is.Equal("member-123", membershipMapped["directory_account_id"]), "the membership references the member's user id, which is the account's external id")
	assert.Check(t, is.Equal("group-123", membershipMapped["directory_group_id"]), "the membership references the group id, which is the group's external id")
	assert.Check(t, is.Equal("OWNER", membershipMapped["role"]))
}

// TestGoogleWorkspaceMappingsFallbacks verifies fallback when payload fields are missing
func TestGoogleWorkspaceMappingsFallbacks(t *testing.T) {
	accountRaw, err := providerkit.EvalMap(context.Background(), mappingtest.MappingSpec(t, testMappings(t), entityops.SchemaDirectoryAccount.Name).MapExpr, types.MappingEnvelope{
		Resource: "sparse@example.com",
		Payload:  json.RawMessage(`{"id":"user-sparse","primaryEmail":"sparse@example.com"}`),
	}, types.MappingInstallation{})
	assert.NilError(t, err)

	accountMapped, err := jsonx.ToMap(accountRaw)
	assert.NilError(t, err)

	assert.Check(t, is.Equal("user-sparse", accountMapped["external_id"]))
	assert.Check(t, is.Equal("sparse@example.com", accountMapped["canonical_email"]))
	assert.Check(t, is.Equal("", accountMapped["display_name"]))
	assert.Check(t, is.Equal("", accountMapped["given_name"]))
	assert.Check(t, is.Equal("", accountMapped["family_name"]))
	assert.Check(t, is.Equal("", accountMapped["organization_unit"]))
	assert.Check(t, is.Equal("ACTIVE", accountMapped["status"]))
	assert.Check(t, is.Equal("DISABLED", accountMapped["mfa_state"]))

	membershipRaw, err := providerkit.EvalMap(context.Background(), mappingtest.MappingSpec(t, testMappings(t), entityops.SchemaDirectoryMembership.Name).MapExpr, types.MappingEnvelope{
		Resource: "eng@example.com",
		Payload:  json.RawMessage(`{"email":"norole@example.com","type":"USER","id":"member-norole"}`),
	}, types.MappingInstallation{})
	assert.NilError(t, err)

	membershipMapped, err := jsonx.ToMap(membershipRaw)
	assert.NilError(t, err)

	assert.Check(t, is.Equal("MEMBER", membershipMapped["role"]), "missing role must fall back to MEMBER so the row survives enum validation")

	emptyRoleRaw, err := providerkit.EvalMap(context.Background(), mappingtest.MappingSpec(t, testMappings(t), entityops.SchemaDirectoryMembership.Name).MapExpr, types.MappingEnvelope{
		Resource: "eng@example.com",
		Payload:  json.RawMessage(`{"email":"emptyrole@example.com","role":"","type":"USER","id":"member-emptyrole"}`),
	}, types.MappingInstallation{})
	assert.NilError(t, err)

	emptyRoleMapped, err := jsonx.ToMap(emptyRoleRaw)
	assert.NilError(t, err)

	assert.Check(t, is.Equal("MEMBER", emptyRoleMapped["role"]), "empty role must fall back to MEMBER so the row survives enum validation")
}
