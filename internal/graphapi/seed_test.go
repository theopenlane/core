package graphapi_test

import (
	"context"

	"github.com/stretchr/testify/require"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
	coreutils "github.com/theopenlane/core/pkg/testutils"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/auth"
)

var (
	// testUser1 is a test user with a personal org and an organization
	testUser1 testUserDetails
	// testUser2 is a test user with a personal org and an organization
	testUser2 testUserDetails
	// viewOnlyUser is a test user that is a member of the first user's organization
	viewOnlyUser testUserDetails
	// adminUser is a test user that is an admin of the first user's organization
	adminUser testUserDetails
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
	// GroupID is the ID of the group created under the organization
	GroupID string
	// UserCtx is the context of the user that should be used for the test requests
	UserCtx context.Context
}

// userBuilder creates a new test user and returns the details
func (suite *GraphTestSuite) userBuilder(ctx context.Context) testUserDetails {
	t := suite.T()

	testUser := testUserDetails{}

	// create a test user
	testUser.UserInfo = (&UserBuilder{client: suite.client}).MustNew(ctx, t)
	testUser.ID = testUser.UserInfo.ID

	// get the personal org for the user
	testPersonalOrg, err := testUser.UserInfo.Edges.Setting.DefaultOrg(ctx)
	require.NoError(t, err)

	testUser.PersonalOrgID = testPersonalOrg.ID

	// setup user context with the personal org
	userCtx, err := auth.NewTestContextWithOrgID(testUser.ID, testUser.PersonalOrgID)
	require.NoError(t, err)

	// create a non-personal test organization
	testOrg := (&OrganizationBuilder{client: suite.client}).MustNew(userCtx, t)
	testUser.OrganizationID = testOrg.ID

	// setup user context with the org (and not the personal org)
	testUser.UserCtx, err = auth.NewTestContextWithOrgID(testUser.ID, testUser.OrganizationID)
	require.NoError(t, err)

	// create a group under the organization
	testGroup := (&GroupBuilder{client: suite.client, Owner: testUser.OrganizationID}).MustNew(testUser.UserCtx, t)
	testUser.GroupID = testGroup.ID

	return testUser
}

// setupTestData creates test users and sets up the clients with the necessary tokens
// this includes three users, two with personal orgs and organizations, and one that is a member of the first user's organization
// as well as an api token and personal access token for the first user
func (suite *GraphTestSuite) setupTestData(ctx context.Context) {
	t := suite.T()

	// create test users
	testUser1 = suite.userBuilder(ctx)
	testUser2 = suite.userBuilder(ctx)

	// setup a test user that is a member of an organization
	viewOnlyUser = suite.userBuilder(ctx)

	// add the user to the organization
	suite.addUserToOrganization(&viewOnlyUser, enums.RoleMember, testUser1.OrganizationID)

	// setup a test user that is an admin of an organization
	adminUser = suite.userBuilder(ctx)
	suite.addUserToOrganization(&adminUser, enums.RoleAdmin, testUser1.OrganizationID)

	// setup client with a personal access token
	pat := (&PersonalAccessTokenBuilder{
		client:          suite.client,
		OwnerID:         testUser1.ID,
		OrganizationIDs: []string{testUser1.OrganizationID, testUser1.OrganizationID}}).
		MustNew(testUser1.UserCtx, t)

	authHeaderPAT := openlaneclient.Authorization{
		BearerToken: pat.Token,
	}

	var err error

	suite.client.apiWithPAT, err = coreutils.TestClientWithAuth(t,
		suite.client.db,
		suite.client.objectStore,
		openlaneclient.WithCredentials(authHeaderPAT))
	require.NoError(t, err)

	// setup client with an API token
	apiToken := (&APITokenBuilder{client: suite.client, OwnerID: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t)

	authHeaderAPIToken := openlaneclient.Authorization{
		BearerToken: apiToken.Token,
	}

	suite.client.apiWithToken, err = coreutils.TestClientWithAuth(t, suite.client.db, suite.client.objectStore, openlaneclient.WithCredentials(authHeaderAPIToken))
	require.NoError(t, err)
}

// addUserToOrganization adds a user to an organization with the provided role and set's the user's organization ID and user context
func (suite *GraphTestSuite) addUserToOrganization(userDetails *testUserDetails, role enums.Role, organizationID string) {
	t := suite.T()

	// update organization to be the read-only member of the first test organization
	(&OrgMemberBuilder{client: suite.client, UserID: userDetails.ID, OrgID: organizationID, Role: role.String()}).MustNew(userDetails.UserCtx, t)

	userDetails.OrganizationID = organizationID

	// update the user context for the org member
	var err error
	userDetails.UserCtx, err = auth.NewTestContextWithOrgID(userDetails.ID, userDetails.OrganizationID)
	require.NoError(t, err)
}

// userContextWithID creates a new user context with the provided user ID
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
