package handlers_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/pkg/entitlements"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
)

// testUserDetails is a struct that holds the details of a test user
type testUserDetails struct {
	// ID is the ID of the user
	ID string
	// UserInfo contains all the details of the user
	UserInfo *ent.User
	// PersonalOrgID is the ID of the personal organization of the user
	PersonalOrgID string
	// OrganizationID is the ID of the organization of the user
	OrganizationID string
	// SubscriptionID is the ID of the subscription of the user
	SubscriptionID string
	// UserCtx is the context of the user that can be used for the test requests that require authentication
	UserCtx context.Context
}

// userBuilder creates a new test user and returns the details
// this includes a test user and an organization the user is the owner of
func (suite *HandlerTestSuite) userBuilder(ctx context.Context) testUserDetails {
	return suite.userBuilderWithInput(ctx, nil)
}

type userInput struct {
	email         string
	password      string
	confirmedUser bool
	tfaEnabled    bool
	features      []models.OrgModule
}

func (suite *HandlerTestSuite) userBuilderWithInput(ctx context.Context, input *userInput) testUserDetails {
	t := suite.T()

	testUser := testUserDetails{}

	ctx = auth.WithInternalOperationContext(ctx)

	// create a test user
	var err error
	var userSetting *ent.UserSetting

	if input != nil && input.confirmedUser {
		userSetting, err = suite.db.UserSetting.Create().
			SetEmailConfirmed(true).
			SetIsTfaEnabled(input.tfaEnabled).
			Save(ctx)
		assert.NilError(t, err)
	}

	// ensure unique email for each test user, even if the same email is provided in the input, by appending a ULID to the email
	email := gofakeit.Email()
	parts := strings.Split(email, "@")
	email = parts[0] + strings.ToLower(ulids.New().String()) + "@" + parts[1]
	if input != nil && input.email != "" {
		email = input.email
	}

	builder := suite.db.User.Create().
		SetEmail(email).
		SetFirstName(gofakeit.FirstName()).
		SetLastName(gofakeit.LastName()).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now())

	if input != nil && input.password != "" {
		builder.SetPassword(input.password)
	}

	if userSetting != nil {
		builder.SetSetting(userSetting)
	}

	testUser.UserInfo, err = builder.Save(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to create test user")
	}
	assert.NilError(t, err)

	testUser.ID = testUser.UserInfo.ID

	// get the personal org for the user
	testPersonalOrg, err := testUser.UserInfo.Edges.Setting.DefaultOrg(ctx)
	assert.NilError(t, err)

	testUser.PersonalOrgID = testPersonalOrg.ID

	// setup user context with the personal org
	userCtx := auth.NewTestContextWithOrgID(testUser.ID, testUser.PersonalOrgID)

	// set privacy allow in order to allow the creation of the users without
	// authentication in the tests seeds
	userCtx = auth.WithInternalOperationContext(userCtx)

	// add client to context, required for hooks that expect the client to be in the context
	userCtx = ent.NewContext(userCtx, suite.db)

	// create a non-personal test organization
	orgSetting, err := suite.db.OrganizationSetting.Create().
		SetBillingEmail(testUser.UserInfo.Email).
		Save(userCtx)
	assert.NilError(t, err)

	testOrg, err := suite.db.Organization.Create().
		SetName(gofakeit.Name() + " " + ulids.New().String()).
		SetSettingID(orgSetting.ID).
		Save(userCtx)

	assert.NilError(t, err)

	testUser.OrganizationID = testOrg.ID

	// setup user context with the org (and not the personal org)
	testUser.UserCtx = auth.NewTestContextWithOrgID(testUser.ID, testUser.OrganizationID)

	// set privacy allow in order to allow the creation of the users without
	// authentication in the tests seeds
	testUser.UserCtx = auth.WithInternalOperationContext(testUser.UserCtx)

	// add client to context, required for hooks that expect the client to be in the context
	testUser.UserCtx = ent.NewContext(testUser.UserCtx, suite.db)

	return testUser
}

// orgMemberBuilder adds a new user to the owner's org with the given role and optional FGA org role relations
func (suite *HandlerTestSuite) orgMemberBuilder(t *testing.T, ctx context.Context, owner testUserDetails, role enums.Role, orgRoleRelations ...string) testUserDetails {
	t.Helper()

	member := suite.userBuilderWithInput(ctx, &userInput{confirmedUser: true})

	err := suite.db.OrgMembership.Create().SetInput(ent.CreateOrgMembershipInput{
		OrganizationID: owner.OrganizationID,
		UserID:         member.ID,
		Role:           &role,
	}).Exec(owner.UserCtx)
	assert.NilError(t, err)

	tuples := make([]fgax.TupleKey, 0, len(orgRoleRelations))
	for _, relation := range orgRoleRelations {
		tuples = append(tuples, fgax.TupleKey{
			Subject:  fgax.Entity{Kind: fgax.Kind(ent.TypeUser), Identifier: member.ID},
			Object:   fgax.Entity{Kind: fgax.Kind(ent.TypeOrganization), Identifier: owner.OrganizationID},
			Relation: fgax.Relation(relation),
		})
	}

	if len(tuples) > 0 {
		_, err = suite.db.Authz.WriteTupleKeys(owner.UserCtx, tuples, nil)
		assert.NilError(t, err)
	}

	member.OrganizationID = owner.OrganizationID
	member.UserCtx = auth.NewTestContextWithOrgID(member.ID, owner.OrganizationID)

	return member
}

func (suite *HandlerTestSuite) enableModules(userID, orgID string, features []models.OrgModule) {
	// default to all modules if none provided
	if len(features) == 0 {
		features = models.AllOrgModules
	}

	newCtx := auth.NewTestContextWithOrgID(userID, orgID)
	newCtx = ent.NewContext(newCtx, suite.db)

	for _, feature := range features {
		_, err := suite.db.OrgModule.Create().
			SetOwnerID(orgID).
			SetModule(feature).
			SetActive(true).
			SetPrice(models.Price{Amount: 0, Interval: "month"}).
			Save(newCtx)
		assert.NilError(suite.T(), err)
	}

	err := entitlements.CreateFeatureTuples(newCtx, &suite.db.OrgModule.Authz, orgID, features)
	assert.NilError(suite.T(), err)
}

// setupTestData creates test users and sets up the clients with the necessary tokens
func (suite *HandlerTestSuite) setupTestData(ctx context.Context) {
	// create test users
	suite.testUser1 = suite.userBuilder(ctx)
	suite.testUser2 = suite.userBuilder(ctx)
}
