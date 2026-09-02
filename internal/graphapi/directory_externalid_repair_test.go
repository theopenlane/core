package graphapi_test

import (
	"context"
	"encoding/json"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"entgo.io/ent/dialect/sql"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
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

// TestDirectoryAccountExternalIDRepair runs the exact query and Modify update the backfill and
// ingest adopt paths use to repair scientific notation external ids; external_id is immutable so
// the write only works through the sql modifier, and this proves that against a real seeded row
func TestDirectoryAccountExternalIDRepair(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	da := (&th.DirectoryAccountBuilder{
		Client:      suite.Client,
		DisplayName: "Notation Repair User",
		OwnerID:     th.SharedTestUser1.OrganizationID,
		ExternalID:  "1.47884153e+08",
	}).MustNew(ctx, t)

	matched, err := suite.Client.DB.DirectoryAccount.Query().
		Where(directoryaccount.ExternalIDContains("e+"), directoryaccount.IDEQ(da.ID)).
		Only(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal("1.47884153e+08", matched.ExternalID))

	err = suite.Client.DB.DirectoryAccount.UpdateOneID(da.ID).
		Modify(func(u *sql.UpdateBuilder) {
			u.Set(directoryaccount.FieldExternalID, "147884153")
		}).
		Exec(ctx)
	assert.NilError(t, err)

	repaired, err := suite.Client.DB.DirectoryAccount.Get(ctx, da.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal("147884153", repaired.ExternalID))

	(&th.Cleanup[*generated.DirectoryAccountDeleteOne]{Client: suite.Client.DB.DirectoryAccount, ID: da.ID}).MustDelete(ctx, t)
}

// TestDirectorySyncAdoptsLegacyScientificKeys seeds directory rows the way the broken CEL
// conversion stored them, then runs a real DirectorySync ingest through ProcessPayloadSets with the
// actual githubapp definition and mappings. The sync should adopt and repair the existing rows
// instead of creating duplicates, and the membership should resolve against the repaired account
func TestDirectorySyncAdoptsLegacyScientificKeys(t *testing.T) {
	def, err := githubapp.Builder(githubapp.Config{})()
	assert.NilError(t, err)

	reg := registry.New()
	assert.NilError(t, reg.Register(def))

	orgUser := suite.UserBuilder(context.Background(), t)
	ctx := th.SetContext(orgUser.UserCtx, suite.Client.DB)

	integration, err := suite.Client.DB.Integration.Create().
		SetName("GitHub Legacy Key Test").
		SetKind("github").
		SetDefinitionID(def.ID).
		SetOwnerID(orgUser.OrganizationID).
		Save(ctx)
	assert.NilError(t, err)

	// seed the rows exactly as a pre-fix sync stored them: scientific notation external ids
	priorRun, err := suite.Client.DB.DirectorySyncRun.Create().
		SetIntegrationID(integration.ID).
		SetOwnerID(orgUser.OrganizationID).
		SetStatus(enums.DirectorySyncRunStatusCompleted).
		Save(ctx)
	assert.NilError(t, err)

	seededAccount, err := suite.Client.DB.DirectoryAccount.Create().
		SetExternalID("1.47884153e+08").
		SetDisplayName("sfunk").
		SetOwnerID(orgUser.OrganizationID).
		SetIntegrationID(integration.ID).
		SetDirectoryInstanceID("sfunk-dev").
		Save(ctx)
	assert.NilError(t, err)

	seededGroup, err := suite.Client.DB.DirectoryGroup.Create().
		SetExternalID("1.7146926e+07").
		SetDisplayName("meow").
		SetOwnerID(orgUser.OrganizationID).
		SetIntegrationID(integration.ID).
		SetDirectoryInstanceID("sfunk-dev").
		SetDirectorySyncRunID(priorRun.ID).
		Save(ctx)
	assert.NilError(t, err)

	// the same member and team as the provider returns them on the next sync
	memberPayload := `{"DatabaseID":147884153,"Login":"sfunk","Name":"Sarah Funkhouser","Email":"","AvatarURL":"https://avatars.githubusercontent.com/u/147884153","OrganizationVerifiedDomainEmails":[],"Org":"sfunk-dev","CanonicalEmail":"sfunk@example.com","EmailAliases":null,"GivenName":"","FamilyName":""}`
	teamPayload := `{"DatabaseID":17146926,"Name":"meow","Slug":"meow","Description":"","Privacy":"","Org":"sfunk-dev"}`
	membershipPayload := `{"Org":"sfunk-dev","Team":{"DatabaseID":17146926,"Slug":"meow"},"Member":{"DatabaseID":147884153,"Login":"sfunk"},"Role":"MAINTAINER"}`

	ic := operations.IngestContext{
		Registry:    reg,
		DB:          suite.Client.DB,
		Integration: integration,
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

	_, err = operations.ProcessPayloadSets(ctx, ic, "DirectorySync", contracts, payloadSets, operations.IngestOptions{})
	assert.NilError(t, err)

	// the seeded account was adopted and repaired in place, not duplicated
	account, err := suite.Client.DB.DirectoryAccount.Get(ctx, seededAccount.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal("147884153", account.ExternalID))

	accountCount, err := suite.Client.DB.DirectoryAccount.Query().
		Where(directoryaccount.OwnerID(orgUser.OrganizationID)).
		Count(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(1, accountCount), "sync must not create a duplicate account")

	// same for the group
	group, err := suite.Client.DB.DirectoryGroup.Get(ctx, seededGroup.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal("17146926", group.ExternalID))

	groupCount, err := suite.Client.DB.DirectoryGroup.Query().
		Where(directorygroup.OwnerID(orgUser.OrganizationID)).
		Count(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(1, groupCount), "sync must not create a duplicate group")

	// the membership resolved its refs against the repaired rows
	membership, err := suite.Client.DB.DirectoryMembership.Query().
		Where(directorymembership.IntegrationID(integration.ID)).
		Only(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(seededAccount.ID, membership.DirectoryAccountID))
	assert.Check(t, is.Equal(seededGroup.ID, membership.DirectoryGroupID))

	(&th.Cleanup[*generated.DirectoryMembershipDeleteOne]{Client: suite.Client.DB.DirectoryMembership, ID: membership.ID}).MustDelete(ctx, t)
	(&th.Cleanup[*generated.DirectoryAccountDeleteOne]{Client: suite.Client.DB.DirectoryAccount, ID: seededAccount.ID}).MustDelete(ctx, t)
	(&th.Cleanup[*generated.DirectoryGroupDeleteOne]{Client: suite.Client.DB.DirectoryGroup, ID: seededGroup.ID}).MustDelete(ctx, t)
	(&th.Cleanup[*generated.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: integration.ID}).MustDelete(ctx, t)
}

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
		Save(ctx)
	assert.NilError(t, err)

	priorRun, err := suite.Client.DB.DirectorySyncRun.Create().
		SetIntegrationID(oldIntegration.ID).
		SetOwnerID(orgUser.OrganizationID).
		SetStatus(enums.DirectorySyncRunStatusCompleted).
		Save(ctx)
	assert.NilError(t, err)

	// the account survives from the old installation, keyed by owner + instance
	seededAccount, err := suite.Client.DB.DirectoryAccount.Create().
		SetExternalID("147884153").
		SetDisplayName("sfunk").
		SetOwnerID(orgUser.OrganizationID).
		SetIntegrationID(oldIntegration.ID).
		SetDirectoryInstanceID("sfunk-dev").
		Save(ctx)
	assert.NilError(t, err)

	// the group upsert is integration scoped, so the old row stays behind and the new
	// installation gets its own
	oldGroup, err := suite.Client.DB.DirectoryGroup.Create().
		SetExternalID("17146926").
		SetDisplayName("meow").
		SetOwnerID(orgUser.OrganizationID).
		SetIntegrationID(oldIntegration.ID).
		SetDirectoryInstanceID("sfunk-dev").
		SetDirectorySyncRunID(priorRun.ID).
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

	_, err = operations.ProcessPayloadSets(ctx, ic, "DirectorySync", contracts, payloadSets, operations.IngestOptions{})
	assert.NilError(t, err)

	// the old installation's account row was matched by owner + instance, not duplicated
	accountCount, err := suite.Client.DB.DirectoryAccount.Query().
		Where(directoryaccount.OwnerID(orgUser.OrganizationID)).
		Count(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(1, accountCount), "reinstall must not duplicate the account")

	account, err := suite.Client.DB.DirectoryAccount.Get(ctx, seededAccount.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(oldIntegration.ID, account.IntegrationID), "adopted account keeps its original integration edge")

	// the new installation created its own group row
	newGroup, err := suite.Client.DB.DirectoryGroup.Query().
		Where(directorygroup.IntegrationID(newIntegration.ID)).
		Only(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal("17146926", newGroup.ExternalID))

	// this membership fails with "unresolved directory account reference" if the resolver
	// scopes by the current integration id instead of owner + instance
	membership, err := suite.Client.DB.DirectoryMembership.Query().
		Where(directorymembership.IntegrationID(newIntegration.ID)).
		Only(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(seededAccount.ID, membership.DirectoryAccountID))
	assert.Check(t, is.Equal(newGroup.ID, membership.DirectoryGroupID))

	(&th.Cleanup[*generated.DirectoryMembershipDeleteOne]{Client: suite.Client.DB.DirectoryMembership, ID: membership.ID}).MustDelete(ctx, t)
	(&th.Cleanup[*generated.DirectoryAccountDeleteOne]{Client: suite.Client.DB.DirectoryAccount, ID: seededAccount.ID}).MustDelete(ctx, t)
	(&th.Cleanup[*generated.DirectoryGroupDeleteOne]{Client: suite.Client.DB.DirectoryGroup, IDs: []string{oldGroup.ID, newGroup.ID}}).MustDelete(ctx, t)
	(&th.Cleanup[*generated.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, IDs: []string{oldIntegration.ID, newIntegration.ID}}).MustDelete(ctx, t)
}
