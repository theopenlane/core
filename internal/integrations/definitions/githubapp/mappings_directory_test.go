package githubapp

import (
	"context"
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// TestGitHubDirectoryMembershipMapping verifies team membership payloads carry the provider role through to the mapped document
func TestGitHubDirectoryMembershipMapping(t *testing.T) {
	spec := githubMappingSpecForSchema(t, integrationgenerated.IntegrationMappingSchemaDirectoryMembership)

	maintainerRaw, err := providerkit.EvalMap(context.Background(), spec.MapExpr, types.MappingEnvelope{
		Resource: "acme/security",
		Payload:  json.RawMessage(`{"Org":"acme","Team":{"DatabaseID":42,"Slug":"security"},"Member":{"DatabaseID":7,"Login":"kwaters"},"Role":"MAINTAINER"}`),
	})
	assert.NilError(t, err)

	maintainerMapped, err := jsonx.ToMap(maintainerRaw)
	assert.NilError(t, err)

	assert.Check(t, is.Equal("MAINTAINER", maintainerMapped["role"]))
	assert.Check(t, is.Equal("7", maintainerMapped["directory_account_id"]))
	assert.Check(t, is.Equal("42", maintainerMapped["directory_group_id"]))

	noRoleRaw, err := providerkit.EvalMap(context.Background(), spec.MapExpr, types.MappingEnvelope{
		Resource: "acme/security",
		Payload:  json.RawMessage(`{"Org":"acme","Team":{"DatabaseID":42,"Slug":"security"},"Member":{"DatabaseID":8,"Login":"sfunk"},"Role":""}`),
	})
	assert.NilError(t, err)

	noRoleMapped, err := jsonx.ToMap(noRoleRaw)
	assert.NilError(t, err)

	assert.Check(t, is.Equal("MEMBER", noRoleMapped["role"]))
}

// TestGitHubDirectoryAccountMapping verifies confirmed email aliases flow into the mapped account document
func TestGitHubDirectoryAccountMapping(t *testing.T) {
	spec := githubMappingSpecForSchema(t, integrationgenerated.IntegrationMappingSchemaDirectoryAccount)

	withAliasesRaw, err := providerkit.EvalMap(context.Background(), spec.MapExpr, types.MappingEnvelope{
		Resource: "acme",
		Payload:  json.RawMessage(`{"DatabaseID":7,"Login":"kwaters","Name":"Kelsey Waters","Email":"","AvatarURL":"https://avatars.example.com/7","OrganizationVerifiedDomainEmails":[],"Org":"acme","CanonicalEmail":"kwaters@example.com","EmailAliases":["kelsey@example.com","kw@example.dev"],"GivenName":"Kelsey","FamilyName":"Waters"}`),
	})
	assert.NilError(t, err)

	withAliasesMapped, err := jsonx.ToMap(withAliasesRaw)
	assert.NilError(t, err)

	assert.Check(t, is.Equal("kwaters@example.com", withAliasesMapped["canonical_email"]))
	assert.Check(t, is.DeepEqual([]any{"kelsey@example.com", "kw@example.dev"}, withAliasesMapped["email_aliases"]))

	noAliasesRaw, err := providerkit.EvalMap(context.Background(), spec.MapExpr, types.MappingEnvelope{
		Resource: "acme",
		Payload:  json.RawMessage(`{"DatabaseID":8,"Login":"sfunk","Name":"Sarah Funkhouser","Email":"","AvatarURL":"","OrganizationVerifiedDomainEmails":[],"Org":"acme","CanonicalEmail":"sfunk@example.com","EmailAliases":null,"GivenName":"","FamilyName":""}`),
	})
	assert.NilError(t, err)

	noAliasesMapped, err := jsonx.ToMap(noAliasesRaw)
	assert.NilError(t, err)

	assert.Check(t, is.DeepEqual([]any{}, noAliasesMapped["email_aliases"]))
}

// TestResolveCanonicalEmail verifies the priority chain and confirmed alias collection
func TestResolveCanonicalEmail(t *testing.T) {
	samlMap := map[string]samlIdentity{
		"kwaters": {NameID: "kwaters@sso.example.com", GivenName: "Kelsey", FamilyName: "Waters"},
	}

	samlMember := &orgMemberNode{orgMemberNodeGQL: orgMemberNodeGQL{
		Login:                            "kwaters",
		Email:                            "public@example.com",
		OrganizationVerifiedDomainEmails: []string{"kwaters@example.com"},
	}}
	resolveCanonicalEmail(samlMember, samlMap)

	assert.Check(t, is.Equal("kwaters@sso.example.com", samlMember.CanonicalEmail))
	assert.Check(t, is.Equal("Kelsey", samlMember.GivenName))
	assert.Check(t, is.DeepEqual([]string{"public@example.com", "kwaters@example.com"}, samlMember.EmailAliases))

	verifiedMember := &orgMemberNode{orgMemberNodeGQL: orgMemberNodeGQL{
		Login:                            "sfunk",
		Email:                            "sfunk@example.com",
		OrganizationVerifiedDomainEmails: []string{"sfunk@example.com", "sarah@example.com"},
	}}
	resolveCanonicalEmail(verifiedMember, nil)

	assert.Check(t, is.Equal("sfunk@example.com", verifiedMember.CanonicalEmail))
	assert.Check(t, is.DeepEqual([]string{"sarah@example.com"}, verifiedMember.EmailAliases), "canonical email must be excluded from aliases")

	publicOnlyMember := &orgMemberNode{orgMemberNodeGQL: orgMemberNodeGQL{
		Login: "mando",
		Email: "MANDO@example.com",
	}}
	resolveCanonicalEmail(publicOnlyMember, nil)

	assert.Check(t, is.Equal("MANDO@example.com", publicOnlyMember.CanonicalEmail))
	assert.Check(t, is.Len(publicOnlyMember.EmailAliases, 0), "case-insensitive match with canonical must be excluded")
}

// githubMappingSpecForSchema returns the mapping override for one schema from the GitHub App defaults
func githubMappingSpecForSchema(t *testing.T, schema string) types.MappingOverride {
	t.Helper()

	for _, mapping := range githubAppMappings() {
		if mapping.Schema == schema {
			return mapping.Spec
		}
	}

	t.Fatalf("mapping not found for schema %s", schema)

	return types.MappingOverride{}
}
