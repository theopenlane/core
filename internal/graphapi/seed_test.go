package graphapi_test

import (
	"context"
	"sync"
	"testing"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	fgamodel "github.com/theopenlane/core/fga/model"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	coreutils "github.com/theopenlane/core/internal/testutils"
	authmw "github.com/theopenlane/core/pkg/middleware/auth"
)

var (
	// sharedTestUser1 is a test user with a personal org and an organization
	sharedTestUser1 testUserDetails
	// sharedTestUser2 is a test user with a personal org and an organization
	sharedTestUser2 testUserDetails
	// sharedViewOnlyUser is a test user that is a member of the first user's organization
	sharedViewOnlyUser testUserDetails
	// sharedViewOnlyUser2 is a test user that is a member of the first user's organization
	sharedViewOnlyUser2 testUserDetails
	// sharedSuperAdminUser is a test user that is a super admin of the first user's organization
	sharedSuperAdminUser testUserDetails
	// sharedAdminUser is a test user that is an admin of the first user's organization
	sharedAdminUser testUserDetails
	// sharedSystemAdminUser is a test user that is a system admin
	sharedSystemAdminUser testUserDetails
	// sharedAuditorUser is a test user that has auditor access to an organization
	sharedAuditorUser testUserDetails
)

// testUserDetails is a struct that holds the details of a test user
type testUserDetails struct {
	// ID is the ID of the user
	ID string
	// UserInfo contains all the details of the user
	UserInfo ent.User
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
func (suite *GraphTestSuite) userBuilder(ctx context.Context, t *testing.T, features ...models.OrgModule) testUserDetails {
	testUser := testUserDetails{}

	// create a test user
	testUser.UserInfo = *(&UserBuilder{client: suite.client}).MustNew(ctx, t)
	testUser.ID = testUser.UserInfo.ID

	// get the personal org for the user
	testPersonalOrg, err := testUser.UserInfo.Edges.Setting.DefaultOrg(ctx)
	requireNoError(t, err)

	testUser.PersonalOrgID = testPersonalOrg.ID

	// setup user context with the personal org
	userCtx := auth.NewTestContextWithOrgID(testUser.ID, testUser.PersonalOrgID)

	// create a non-personal test organization
	testOrg := (&OrganizationBuilder{client: suite.client, Features: features}).MustNew(userCtx, t)
	testUser.OrganizationID = testOrg.ID

	// setup user context with the org; users who create an org are owners
	testUser.UserCtx = auth.NewTestContextWithOrgID(testUser.ID, testUser.OrganizationID, auth.WithOrganizationRole(auth.OwnerRole))

	// create a group under the organization
	testGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	testUser.GroupID = testGroup.ID

	return testUser
}

var seedOnce sync.Once
var seedErr error

// setupTestData creates test users and sets up the clients with the necessary tokens
// this includes three users, two with personal orgs and organizations, and one that is a member of the first user's organization
// as well as an api token and personal access token for the first user
// all data using this should be cleaned up after each test to ensure no conflicts between tests
// if there are potential conflicts, new users should be created for the test
func (suite *GraphTestSuite) setupTestData(ctx context.Context, t *testing.T) {
	t.Helper()
	seedOnce.Do(func() {
		// create system org
		(&OrganizationBuilder{client: suite.client, SystemOrg: true}).MustNew(ctx, t)

		// create system admin user
		sharedSystemAdminUser = suite.systemAdminBuilder(ctx, t)

		sharedTestUser1 = suite.userBuilder(ctx, t)
		sharedTestUser2 = suite.userBuilder(ctx, t)

		// setup two test users that are members of the organization
		sharedViewOnlyUser = suite.userBuilder(ctx, t)
		sharedViewOnlyUser2 = suite.userBuilder(ctx, t)

		// add the user to the organization
		suite.addUserToOrganization(sharedTestUser1.UserCtx, t, &sharedViewOnlyUser, enums.RoleMember, sharedTestUser1.OrganizationID)
		suite.addUserToOrganization(sharedTestUser1.UserCtx, t, &sharedViewOnlyUser2, enums.RoleMember, sharedTestUser1.OrganizationID)

		// setup a test user that is a super admin of an organization
		sharedSuperAdminUser = suite.userBuilder(ctx, t)
		suite.addUserToOrganization(sharedTestUser1.UserCtx, t, &sharedSuperAdminUser, enums.RoleSuperAdmin, sharedTestUser1.OrganizationID)

		// setup a test user that is an admin of an organization
		sharedAdminUser = suite.userBuilder(ctx, t)
		suite.addUserToOrganization(sharedTestUser1.UserCtx, t, &sharedAdminUser, enums.RoleAdmin, sharedTestUser1.OrganizationID)

		// setup a test user that is an auditor for an organization
		sharedAuditorUser = suite.userBuilder(ctx, t)
		suite.addUserToOrganization(sharedTestUser1.UserCtx, t, &sharedAuditorUser, enums.RoleAuditor, sharedTestUser1.OrganizationID)

		suite.client.apiWithPAT = suite.setupPatClient(sharedTestUser1, t)
		suite.client.apiWithToken = suite.setupAPITokenClient(sharedTestUser1.UserCtx, t)
		suite.client.apiWithTokenOrg2 = suite.setupAPITokenClient(sharedTestUser2.UserCtx, t)
	})

	requireNoError(t, seedErr)
}

func (suite *GraphTestSuite) setupPatClient(user testUserDetails, t *testing.T) *testclient.TestClient {
	// setup client with a personal access token
	pat := (&PersonalAccessTokenBuilder{client: suite.client, OrganizationIDs: []string{user.OrganizationID, user.PersonalOrgID}}).MustNew(user.UserCtx, t)

	authHeaderPAT := testclient.Authorization{
		BearerToken: pat.Token,
	}

	apiClientPat, err := coreutils.TestClientWithAuth(suite.client.db, suite.client.objectStore,
		testclient.WithCredentials(authHeaderPAT),
		testclient.WithInterceptors(
			testclient.WithOrganizationHeader(user.OrganizationID),
		))
	requireNoError(t, err)

	return apiClientPat
}

func (suite *GraphTestSuite) setupAPITokenClient(ctx context.Context, t *testing.T) *testclient.TestClient {
	// setup client with an API token with comprehensive scopes for testing
	// Get all available scopes from the FGA model
	scopeOpts, err := fgamodel.ScopeOptions()
	requireNoError(t, err)

	var scopes []string
	for obj, verbs := range scopeOpts {
		for _, verb := range verbs {
			scopes = append(scopes, verb+":"+obj)
		}
	}

	apiToken := (&APITokenBuilder{client: suite.client, Scopes: scopes}).MustNew(ctx, t)

	authHeaderAPIToken := testclient.Authorization{
		BearerToken: apiToken.Token,
	}

	apiClientToken, err := coreutils.TestClientWithAuth(suite.client.db, suite.client.objectStore, testclient.WithCredentials(authHeaderAPIToken))
	requireNoError(t, err)

	return apiClientToken
}

// addUserToOrganization adds a user to an organization with the provided role and set's the user's organization ID and user context
// the context passed in is the context that has access to the organization the user is being added to
func (suite *GraphTestSuite) addUserToOrganization(ctx context.Context, t *testing.T, userDetails *testUserDetails, role enums.Role, organizationID string) {
	// update organization to be the read-only member of the first test organization
	(&OrgMemberBuilder{client: suite.client, UserID: userDetails.ID, Role: role.String()}).MustNew(ctx, t)

	userDetails.OrganizationID = organizationID

	// update the user context for the org member; set the role so permission checks that read
	// caller.OrganizationRole (instead of querying the DB) work correctly
	orgRole, _ := auth.ToOrganizationRoleType(role.String())
	userDetails.UserCtx = auth.NewTestContextWithOrgID(userDetails.ID, userDetails.OrganizationID, auth.WithOrganizationRole(orgRole))
}

func (suite *GraphTestSuite) systemAdminBuilder(ctx context.Context, t *testing.T) testUserDetails {
	newUser := suite.userBuilder(ctx, t)

	req := fgax.TupleRequest{
		SubjectID:   newUser.ID,
		SubjectType: auth.UserSubjectType,
		ObjectID:    authmw.SystemObjectID,
		ObjectType:  authmw.SystemObject,
		Relation:    fgax.SystemAdminRelation,
	}

	// add system admin relation for user
	_, err := suite.client.db.Authz.WriteTupleKeys(context.Background(), []fgax.TupleKey{fgax.GetTupleKey(req)}, nil)
	requireNoError(t, err)

	// set the user as a system admin
	newUser.UserCtx = auth.NewTestContextForSystemAdmin(newUser.ID, newUser.OrganizationID)

	return newUser
}

// resetContext resets the context to ensure it has not additional data that could conflict with the test
// if the context is the background context, it returns the same context
// because the context is used with a test client and we are not passing in the client here
func resetContext(ctx context.Context, t *testing.T) context.Context {
	if ctx == context.Background() {
		return ctx
	}

	caller, callerOk := auth.CallerFromContext(ctx)
	assert.Check(t, callerOk, "caller not found in context")

	// ensure system admin context is kept in the new context
	if caller.Has(auth.CapSystemAdmin) {
		return auth.NewTestContextForSystemAdmin(caller.SubjectID, caller.OrganizationID)
	}

	return auth.NewTestContextWithOrgID(caller.SubjectID, caller.OrganizationID, auth.WithOrganizationRole(caller.OrganizationRole))
}

// testOrgUsers is all available roles with api and pat clients to used with tests
type testOrgUsers struct {
	owner          *testUserDetails
	superAdmin     *testUserDetails
	admin          *testUserDetails
	member         *testUserDetails
	auditor        *testUserDetails
	adminApiClient *testclient.TestClient
	adminPatClient *testclient.TestClient
}

// testMinimalOrgUsers is a subset of org users created when all roles do not need to be tested
type testMinimalOrgUsers struct {
	owner          *testUserDetails
	admin          *testUserDetails
	member         *testUserDetails
	apiClient      *testclient.TestClient
	adminPatClient *testclient.TestClient
}

// testOwner only creates a org with a single user (owner) and api clients
type testOwner struct {
	owner     *testUserDetails
	apiClient *testclient.TestClient
	patClient *testclient.TestClient
}

// seedOrgOwner will seed the owner and api clients
func (suite *GraphTestSuite) seedOrgOwner(t *testing.T) *testOwner {
	t.Helper()
	localOwner := suite.userBuilder(context.Background(), t)

	return &testOwner{
		owner:     &localOwner,
		apiClient: suite.setupAPITokenClient(localOwner.UserCtx, t),
		patClient: suite.setupPatClient(localOwner, t),
	}
}

// seedFreshMinimalOrgUsers will seed the owner, admin, and member but leave out the super admin, auditor, and api clients
func (suite *GraphTestSuite) seedFreshMinimalOrgUsers(t *testing.T, includeClients bool) *testMinimalOrgUsers {
	t.Helper()
	localOwner := suite.userBuilder(context.Background(), t)
	localAdmin := suite.userBuilder(context.Background(), t)
	localMember := suite.userBuilder(context.Background(), t)

	suite.addUserToOrganization(localOwner.UserCtx, t, &localAdmin, enums.RoleAdmin, localOwner.OrganizationID)
	suite.addUserToOrganization(localOwner.UserCtx, t, &localMember, enums.RoleMember, localOwner.OrganizationID)

	out := &testMinimalOrgUsers{
		owner:  &localOwner,
		admin:  &localAdmin,
		member: &localMember,
	}

	if includeClients {
		out.apiClient = suite.setupAPITokenClient(localAdmin.UserCtx, t)
		out.adminPatClient = suite.setupPatClient(localAdmin, t)
	}

	return out
}

// seedFreshOrgUsers is a helper function to setup an entire new set of users that can be used when you do not want organization conflicts between tests
func (suite *GraphTestSuite) seedFreshOrgUsers(t *testing.T) *testOrgUsers {
	t.Helper()
	localOwner := suite.userBuilder(context.Background(), t)
	localSuperAdmin := suite.userBuilder(context.Background(), t)
	localAdmin := suite.userBuilder(context.Background(), t)
	localMember := suite.userBuilder(context.Background(), t)
	localAuditor := suite.userBuilder(context.Background(), t)

	suite.addUserToOrganization(localOwner.UserCtx, t, &localSuperAdmin, enums.RoleSuperAdmin, localOwner.OrganizationID)
	suite.addUserToOrganization(localOwner.UserCtx, t, &localAdmin, enums.RoleAdmin, localOwner.OrganizationID)
	suite.addUserToOrganization(localOwner.UserCtx, t, &localMember, enums.RoleMember, localOwner.OrganizationID)
	suite.addUserToOrganization(localOwner.UserCtx, t, &localAuditor, enums.RoleAuditor, localOwner.OrganizationID)

	apiTokenClient := suite.setupAPITokenClient(localAdmin.UserCtx, t)
	adminPersonalAccessTokenClient := suite.setupPatClient(localAdmin, t)

	return &testOrgUsers{
		owner:          &localOwner,
		superAdmin:     &localSuperAdmin,
		admin:          &localAdmin,
		member:         &localMember,
		auditor:        &localAuditor,
		adminApiClient: apiTokenClient,
		adminPatClient: adminPersonalAccessTokenClient,
	}
}
