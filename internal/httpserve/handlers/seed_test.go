package handlers_test

import (
	"context"

	"github.com/stretchr/testify/require"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/auth"
)

var (
	testUser1 testUserDetails
)

type testUserDetails struct {
	ID             string
	UserInfo       *ent.User
	PersonalOrgID  string
	OrganizationID string
	UserCtx        context.Context
}

func (suite *HandlerTestSuite) userBuilder(ctx context.Context) testUserDetails {
	t := suite.T()

	testUser := testUserDetails{}

	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// create a test user
	var err error
	testUser.UserInfo, err = suite.db.User.Create().
		SetEmail("marco@theopenlane.io").
		SetFirstName("Marco").
		SetLastName("Polo").
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
		SetName("mitb").
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

// userContextWithID creates a new user context with the provided user ID
// and adds it to a new echo context
func userContextWithID(userID string) (context.Context, error) {
	// Use that user to create the organization
	ec, err := auth.NewTestEchoContextWithValidUser(userID)
	if err != nil {
		return nil, err
	}

	reqCtx := context.WithValue(ec.Request().Context(), echocontext.EchoContextKey, ec)

	ec.SetRequest(ec.Request().WithContext(reqCtx))

	return reqCtx, nil
}
