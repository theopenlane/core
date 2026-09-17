package graphapi_test

import (
	"context"
	"encoding/json"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorymembership"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/githubapp"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// TestDirectorySyncResolvesAccountsAcrossReinstall covers the reinstall scenario from production:
// the account row belongs to a previous installation of the same definition, and the membership
// resolver has to find it by owner + directory instance rather than the current integration id.
// External ids are seeded in clean decimal form so this exercises only the scoping change
func TestDirectorySyncResolvesAccountsAcrossReinstall(t *testing.T) {
	def, err := githubapp.Builder(githubapp.Config{})()
	assert.NilError(t, err)

	reg := registry.New()
	assert.NilError(t, reg.Register(def))

	orgUser := suite.UserBuilder(context.Background(), t)
	ctx := th.SetContext(orgUser.UserCtx, suite.Client.DB)

	oldIntegration, err := suite.Client.DB.Integration.Create().
		SetName("GitHub Old Install").
		SetKind("github").
		SetDefinitionID(def.ID).
		SetOwnerID(orgUser.OrganizationID).
		Save(ctx)
	assert.NilError(t, err)

	newIntegration, err := suite.Client.DB.Integration.Create().
		SetName("GitHub New Install").
		SetKind("github").
		SetDefinitionID(def.ID).
		SetOwnerID(orgUser.OrganizationID).
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "sfunk-dev"}}).
		Save(ctx)
	assert.NilError(t, err)

	// the account survives from the old installation, keyed by owner + instance
	seededAccount, err := suite.Client.DB.DirectoryAccount.Create().
		SetExternalID("147884153").
		SetDisplayName("sfunk").
		SetOwnerID(orgUser.OrganizationID).
		SetIntegrationID(oldIntegration.ID).
		SetSourceInstanceID("sfunk-dev").
		Save(ctx)
	assert.NilError(t, err)

	// the group survives from the old installation the same way, keyed by owner + instance
	oldGroup, err := suite.Client.DB.DirectoryGroup.Create().
		SetExternalID("17146926").
		SetDisplayName("meow").
		SetOwnerID(orgUser.OrganizationID).
		SetIntegrationID(oldIntegration.ID).
		SetSourceInstanceID("sfunk-dev").
		Save(ctx)
	assert.NilError(t, err)

	memberPayload := `{"DatabaseID":147884153,"Login":"sfunk","Name":"Sarah Funkhouser","Email":"","AvatarURL":"https://avatars.githubusercontent.com/u/147884153","OrganizationVerifiedDomainEmails":[],"Org":"sfunk-dev","CanonicalEmail":"sfunk@example.com","EmailAliases":null,"GivenName":"","FamilyName":""}`
	teamPayload := `{"DatabaseID":17146926,"Name":"meow","Slug":"meow","Description":"","Privacy":"","Org":"sfunk-dev"}`
	membershipPayload := `{"Org":"sfunk-dev","Team":{"DatabaseID":17146926,"Slug":"meow"},"Member":{"DatabaseID":147884153,"Login":"sfunk"},"Role":"MAINTAINER"}`

	ic := operations.IngestContext{
		Registry:    reg,
		DB:          suite.Client.DB,
		Integration: newIntegration,
	}

	contracts := []types.IngestContract{
		{Schema: entityops.SchemaDirectoryAccount.Name},
		{Schema: entityops.SchemaDirectoryGroup.Name},
		{Schema: entityops.SchemaDirectoryMembership.Name},
	}

	payloadSets := []types.IngestPayloadSet{
		{
			Schema:    entityops.SchemaDirectoryAccount.Name,
			Envelopes: []types.MappingEnvelope{{Resource: "sfunk-dev/sfunk", Payload: json.RawMessage(memberPayload)}},
		},
		{
			Schema:    entityops.SchemaDirectoryGroup.Name,
			Envelopes: []types.MappingEnvelope{{Resource: "sfunk-dev/meow", Payload: json.RawMessage(teamPayload)}},
		},
		{
			Schema:    entityops.SchemaDirectoryMembership.Name,
			Envelopes: []types.MappingEnvelope{{Resource: "sfunk-dev/meow", Payload: json.RawMessage(membershipPayload)}},
		},
	}

	_, err = operations.ProcessPayloadSets(ctx, ic, "DirectorySync", contracts, types.ExecutionPolicy{Snapshot: true}, payloadSets, operations.IngestOptions{})
	assert.NilError(t, err)

	// the old installation's account row was matched by owner + instance, not duplicated
	accountCount, err := suite.Client.DB.DirectoryAccount.Query().
		Where(directoryaccount.OwnerID(orgUser.OrganizationID)).
		Count(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(1, accountCount), "reinstall must not duplicate the account")

	account, err := suite.Client.DB.DirectoryAccount.Get(ctx, seededAccount.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(newIntegration.ID, account.IntegrationID), "adopted account repoints to the syncing installation")

	// the group upsert is instance scoped the same way, so the new installation adopts the
	// old row instead of creating its own
	groupCount, err := suite.Client.DB.DirectoryGroup.Query().
		Where(directorygroup.OwnerID(orgUser.OrganizationID)).
		Count(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(1, groupCount), "reinstall must not duplicate the group")

	group, err := suite.Client.DB.DirectoryGroup.Get(ctx, oldGroup.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(newIntegration.ID, group.IntegrationID), "adopted group repoints to the syncing installation")

	// this membership fails with "unresolved directory account reference" if the resolver
	// scopes by the current integration id instead of owner + instance
	membership, err := suite.Client.DB.DirectoryMembership.Query().
		Where(directorymembership.IntegrationID(newIntegration.ID)).
		Only(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(seededAccount.ID, membership.DirectoryAccountID))
	assert.Check(t, is.Equal(oldGroup.ID, membership.DirectoryGroupID))

	(&th.Cleanup[*generated.DirectoryMembershipDeleteOne]{Client: suite.Client.DB.DirectoryMembership, ID: membership.ID}).MustDelete(ctx, t)
	(&th.Cleanup[*generated.DirectoryAccountDeleteOne]{Client: suite.Client.DB.DirectoryAccount, ID: seededAccount.ID}).MustDelete(ctx, t)
	(&th.Cleanup[*generated.DirectoryGroupDeleteOne]{Client: suite.Client.DB.DirectoryGroup, ID: oldGroup.ID}).MustDelete(ctx, t)
	(&th.Cleanup[*generated.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, IDs: []string{oldIntegration.ID, newIntegration.ID}}).MustDelete(ctx, t)
}
