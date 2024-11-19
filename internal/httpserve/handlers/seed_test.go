package handlers_test

import (
	"context"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/iam/auth"
)

var (
	testUser1 testUserDetails
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
	// UserCtx is the context of the user that can be used for the test requests that require authentication
	UserCtx context.Context
}

// userBuilder creates a new test user and returns the details
// this includes a test user and an organization the user is the owner of
func (suite *HandlerTestSuite) userBuilder(ctx context.Context) testUserDetails {
	t := suite.T()

	testUser := testUserDetails{}

	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// create a test user
	var err error
	testUser.UserInfo, err = suite.db.User.Create().
		SetEmail(gofakeit.Email()).
		SetFirstName(gofakeit.FirstName()).
		SetLastName(gofakeit.LastName()).
		Save(ctx)
	require.NoError(t, err)

	testUser.ID = testUser.UserInfo.ID

	// get the personal org for the user
	testPersonalOrg, err := testUser.UserInfo.Edges.Setting.DefaultOrg(ctx)
	require.NoError(t, err)

	testUser.PersonalOrgID = testPersonalOrg.ID

	// setup user context with the personal org
	userCtx, err := auth.NewTestContextWithOrgID(testUser.ID, testUser.PersonalOrgID)
	require.NoError(t, err)

	// set privacy allow in order to allow the creation of the users without
	// authentication in the tests seeds
	userCtx = privacy.DecisionContext(userCtx, privacy.Allow)

	// add client to context, required for hooks that expect the client to be in the context
	userCtx = ent.NewContext(userCtx, suite.db)

	// create a non-personal test organization
	testOrg := suite.db.Organization.Create().
		SetName(gofakeit.AdjectiveDescriptive() + " " + gofakeit.Noun()).
		SaveX(userCtx)

	testUser.OrganizationID = testOrg.ID

	// setup user context with the org (and not the personal org)
	testUser.UserCtx, err = auth.NewTestContextWithOrgID(testUser.ID, testUser.OrganizationID)
	require.NoError(t, err)

	return testUser
}

// setupTestData creates test users and sets up the clients with the necessary tokens
func (suite *HandlerTestSuite) setupTestData(ctx context.Context) {
	// create test users
	testUser1 = suite.userBuilder(ctx)
}
