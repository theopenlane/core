//go:build test

package testharness

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	fgamodel "github.com/theopenlane/core/v2/fga/model"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	coreutils "github.com/theopenlane/core/v2/internal/testutils"
	authmw "github.com/theopenlane/core/v2/pkg/middleware/auth"
)

var (
	// SharedTestUser1 is a test user with a personal org and an organization
	SharedTestUser1 TestUserDetails
	// SharedTestUser2 is a test user with a personal org and an organization
	SharedTestUser2 TestUserDetails
	// SharedViewOnlyUser is a test user that is a member of the first user's organization
	SharedViewOnlyUser TestUserDetails
	// SharedViewOnlyUser2 is a test user that is a member of the first user's organization
	SharedViewOnlyUser2 TestUserDetails
	// SharedSuperAdminUser is a test user that is a super admin of the first user's organization
	SharedSuperAdminUser TestUserDetails
	// SharedAdminUser is a test user that is an admin of the first user's organization
	SharedAdminUser TestUserDetails
	// SharedSystemAdminUser is a test user that is a system admin
	SharedSystemAdminUser TestUserDetails
	// SharedAuditorUser is a test user that has auditor access to an organization
	SharedAuditorUser TestUserDetails
	// SharedSupportCtx is a request context for an org-scoped support session (auth.NewOrgSupportCaller)
	// on SharedTestUser1's organization
	SharedSupportCtx context.Context
)

// SupportSubjectName/SupportSubjectEmail identify the synthetic
// org-scoped support caller returned by NewSupportCtx
const (
	SupportSubjectName  = "Openlane Support"
	SupportSubjectEmail = "support@theopenlane.io"
)

// NewSupportCtx builds a request context for an org-scoped support session (auth.NewOrgSupportCaller)
// on organizationID, layered on top of baseCtx. Support sessions are scoped to a single org, so a
// context built for one org cannot be used to reach another org's resources
func NewSupportCtx(baseCtx context.Context, organizationID string) context.Context {
	caller := auth.NewOrgSupportCaller(organizationID, auth.SupportSubjectID, SupportSubjectName, SupportSubjectEmail)
	return auth.WithCaller(baseCtx, caller)
}

// TestUserDetails is a struct that holds the details of a test user
type TestUserDetails struct {
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

// UserBuilder creates a new test user and returns the details
func (suite *GraphTestSuite) UserBuilder(ctx context.Context, t *testing.T, features ...models.OrgModule) TestUserDetails {
	testUser := TestUserDetails{}

	// create a test user
	testUser.UserInfo = *(&UserBuilder{Client: suite.Client}).MustNew(ctx, t)
	testUser.ID = testUser.UserInfo.ID

	// get the personal org for the user
	testPersonalOrg, err := testUser.UserInfo.Edges.Setting.DefaultOrg(ctx)
	RequireNoError(t, err)

	testUser.PersonalOrgID = testPersonalOrg.ID

	// setup user context with the personal org
	userCtx := auth.NewTestContextWithOrgID(testUser.ID, testUser.PersonalOrgID)

	// create a non-personal test organization
	testOrg := (&OrganizationBuilder{Client: suite.Client, Features: features}).MustNew(userCtx, t)
	testUser.OrganizationID = testOrg.ID

	// setup user context with the org; users who create an org are owners
	testUser.UserCtx = auth.NewTestContextWithOrgID(testUser.ID, testUser.OrganizationID, auth.WithOrganizationRole(auth.OwnerRole))

	// create a group under the organization
	testGroup := (&GroupBuilder{Client: suite.Client}).MustNew(testUser.UserCtx, t)
	testUser.GroupID = testGroup.ID

	return testUser
}

var seedOnce sync.Once
var seedErr error

// SetupTestData creates test users and sets up the clients with the necessary tokens
// this includes three users, two with personal orgs and organizations, and one that is a member of the first user's organization
// as well as an api token and personal access token for the first user
// all data using this should be cleaned up after each test to ensure no conflicts between tests
// if there are potential conflicts, new users should be created for the test
func (suite *GraphTestSuite) SetupTestData(ctx context.Context, t *testing.T) {
	t.Helper()
	seedOnce.Do(func() {
		// create system org
		(&OrganizationBuilder{Client: suite.Client, SystemOrg: true}).MustNew(ctx, t)

		// create system admin user
		SharedSystemAdminUser = suite.SystemAdminBuilder(ctx, t)

		SharedTestUser1 = suite.UserBuilder(ctx, t)
		SharedTestUser2 = suite.UserBuilder(ctx, t)

		// setup two test users that are members of the organization
		SharedViewOnlyUser = suite.UserBuilder(ctx, t)
		SharedViewOnlyUser2 = suite.UserBuilder(ctx, t)

		// add the user to the organization
		suite.AddUserToOrganization(SharedTestUser1.UserCtx, t, &SharedViewOnlyUser, enums.RoleMember, SharedTestUser1.OrganizationID)
		suite.AddUserToOrganization(SharedTestUser1.UserCtx, t, &SharedViewOnlyUser2, enums.RoleMember, SharedTestUser1.OrganizationID)

		// setup a test user that is a super admin of an organization
		SharedSuperAdminUser = suite.UserBuilder(ctx, t)
		suite.AddUserToOrganization(SharedTestUser1.UserCtx, t, &SharedSuperAdminUser, enums.RoleSuperAdmin, SharedTestUser1.OrganizationID)

		// setup a test user that is an admin of an organization
		SharedAdminUser = suite.UserBuilder(ctx, t)
		suite.AddUserToOrganization(SharedTestUser1.UserCtx, t, &SharedAdminUser, enums.RoleAdmin, SharedTestUser1.OrganizationID)

		// setup a test user that is an auditor for an organization
		SharedAuditorUser = suite.UserBuilder(ctx, t)
		suite.AddUserToOrganization(SharedTestUser1.UserCtx, t, &SharedAuditorUser, enums.RoleAuditor, SharedTestUser1.OrganizationID)

		suite.Client.APIWithPAT = suite.SetupPatClient(SharedTestUser1, t)
		suite.Client.APIWithToken = suite.SetupAPITokenClient(SharedTestUser1.UserCtx, t)
		suite.Client.APIWithTokenOrg2 = suite.SetupAPITokenClient(SharedTestUser2.UserCtx, t)

		// set up an org-scoped support session on SharedTestUser1's org
		SharedSupportCtx = NewSupportCtx(SharedTestUser1.UserCtx, SharedTestUser1.OrganizationID)
	})

	RequireNoError(t, seedErr)
}

func (suite *GraphTestSuite) SetupPatClient(user TestUserDetails, t *testing.T) *testclient.TestClient {
	// setup client with a personal access token
	pat := (&PersonalAccessTokenBuilder{Client: suite.Client, OrganizationIDs: []string{user.OrganizationID, user.PersonalOrgID}}).MustNew(user.UserCtx, t)

	authHeaderPAT := testclient.Authorization{
		BearerToken: pat.Token,
	}

	apiClientPat, err := coreutils.TestClientWithAuth(suite.Client.DB, suite.Client.ObjectStore,
		testclient.WithCredentials(authHeaderPAT),
		testclient.WithInterceptors(
			testclient.WithOrganizationHeader(user.OrganizationID),
		))
	RequireNoError(t, err)

	return apiClientPat
}

func (suite *GraphTestSuite) SetupAPITokenClient(ctx context.Context, t *testing.T) *testclient.TestClient {
	// setup client with an API token with comprehensive scopes for testing
	// Get all available scopes from the FGA model
	scopeOpts, err := fgamodel.ScopeOptions()
	RequireNoError(t, err)

	var scopes []string
	for obj, verbs := range scopeOpts {
		for _, verb := range verbs {
			scopes = append(scopes, obj+":"+verb)
		}
	}

	return SetupAPIToken(ctx, t, scopes)
}

// SetupAPIToken takes scopes and returns an api client with those scopes set
func SetupAPIToken(ctx context.Context, t *testing.T, scopes []string) *testclient.TestClient {
	apiToken := (&APITokenBuilder{Client: Suite.Client, Scopes: scopes}).MustNew(ctx, t)

	authHeaderAPIToken := testclient.Authorization{
		BearerToken: apiToken.Token,
	}

	apiClientToken, err := coreutils.TestClientWithAuth(Suite.Client.DB, Suite.Client.ObjectStore, testclient.WithCredentials(authHeaderAPIToken))
	RequireNoError(t, err)

	return apiClientToken
}

// AddUserToOrganization adds a user to an organization with the provided role and set's the user's organization ID and user context
// the context passed in is the context that has access to the organization the user is being added to
func (suite *GraphTestSuite) AddUserToOrganization(ctx context.Context, t *testing.T, userDetails *TestUserDetails, role enums.Role, organizationID string) {
	// update organization to be the read-only member of the first test organization
	(&OrgMemberBuilder{Client: suite.Client, UserID: userDetails.ID, Role: role.String()}).MustNew(ctx, t)

	userDetails.OrganizationID = organizationID

	// update the user context for the org member; set the role so permission checks that read
	// caller.OrganizationRole (instead of querying the DB) work correctly
	orgRole, _ := auth.ToOrganizationRoleType(role.String())
	userDetails.UserCtx = auth.NewTestContextWithOrgID(userDetails.ID, userDetails.OrganizationID, auth.WithOrganizationRole(orgRole))
}

func (suite *GraphTestSuite) SystemAdminBuilder(ctx context.Context, t *testing.T) TestUserDetails {
	newUser := suite.UserBuilder(ctx, t)

	req := fgax.TupleRequest{
		SubjectID:   newUser.ID,
		SubjectType: auth.UserSubjectType,
		ObjectID:    authmw.SystemObjectID,
		ObjectType:  authmw.SystemObject,
		Relation:    fgax.SystemAdminRelation,
	}

	// add system admin relation for user
	_, err := suite.Client.DB.Authz.WriteTupleKeys(context.Background(), []fgax.TupleKey{fgax.GetTupleKey(req)}, nil)
	RequireNoError(t, err)

	// set the user as a system admin
	newUser.UserCtx = auth.NewTestContextForSystemAdmin(newUser.ID, newUser.OrganizationID)

	return newUser
}

// ResetContext resets the context to ensure it has not additional data that could conflict with the test
// if the context is the background context, it returns the same context
// because the context is used with a test client and we are not passing in the client here
func ResetContext(ctx context.Context, t *testing.T) context.Context {
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

// TestOrgUsers is all available roles with api and pat clients to used with tests
type TestOrgUsers struct {
	Owner          *TestUserDetails
	SuperAdmin     *TestUserDetails
	Admin          *TestUserDetails
	Member         *TestUserDetails
	Auditor        *TestUserDetails
	AdminAPIClient *testclient.TestClient
	AdminPatClient *testclient.TestClient
}

// TestMinimalOrgUsers is a subset of org users created when all roles do not need to be tested
type TestMinimalOrgUsers struct {
	Owner          *TestUserDetails
	Admin          *TestUserDetails
	Member         *TestUserDetails
	APIClient      *testclient.TestClient
	AdminPatClient *testclient.TestClient
}

// TestOwner only creates a org with a single user (owner) and api clients
type TestOwner struct {
	Owner     *TestUserDetails
	APIClient *testclient.TestClient
	PatClient *testclient.TestClient
}

// SeedOrgOwner will seed the owner and api clients
func (suite *GraphTestSuite) SeedOrgOwner(t *testing.T) *TestOwner {
	t.Helper()
	localOwner := suite.UserBuilder(context.Background(), t)

	return &TestOwner{
		Owner:     &localOwner,
		APIClient: suite.SetupAPITokenClient(localOwner.UserCtx, t),
		PatClient: suite.SetupPatClient(localOwner, t),
	}
}

// SeedFreshMinimalOrgUsers will seed the owner, admin, and member but leave out the super admin, auditor, and api clients
func (suite *GraphTestSuite) SeedFreshMinimalOrgUsers(t *testing.T, includeClients bool) *TestMinimalOrgUsers {
	t.Helper()
	localOwner := suite.UserBuilder(context.Background(), t)
	localAdmin := suite.UserBuilder(context.Background(), t)
	localMember := suite.UserBuilder(context.Background(), t)

	suite.AddUserToOrganization(localOwner.UserCtx, t, &localAdmin, enums.RoleAdmin, localOwner.OrganizationID)
	suite.AddUserToOrganization(localOwner.UserCtx, t, &localMember, enums.RoleMember, localOwner.OrganizationID)

	out := &TestMinimalOrgUsers{
		Owner:  &localOwner,
		Admin:  &localAdmin,
		Member: &localMember,
	}

	if includeClients {
		out.APIClient = suite.SetupAPITokenClient(localAdmin.UserCtx, t)
		out.AdminPatClient = suite.SetupPatClient(localAdmin, t)
	}

	return out
}

// SeedFreshOrgUsers is a helper function to setup an entire new set of users that can be used when you do not want organization conflicts between tests
func (suite *GraphTestSuite) SeedFreshOrgUsers(t *testing.T) *TestOrgUsers {
	t.Helper()
	localOwner := suite.UserBuilder(context.Background(), t)
	localSuperAdmin := suite.UserBuilder(context.Background(), t)
	localAdmin := suite.UserBuilder(context.Background(), t)
	localMember := suite.UserBuilder(context.Background(), t)
	localAuditor := suite.UserBuilder(context.Background(), t)

	suite.AddUserToOrganization(localOwner.UserCtx, t, &localSuperAdmin, enums.RoleSuperAdmin, localOwner.OrganizationID)
	suite.AddUserToOrganization(localOwner.UserCtx, t, &localAdmin, enums.RoleAdmin, localOwner.OrganizationID)
	suite.AddUserToOrganization(localOwner.UserCtx, t, &localMember, enums.RoleMember, localOwner.OrganizationID)
	suite.AddUserToOrganization(localOwner.UserCtx, t, &localAuditor, enums.RoleAuditor, localOwner.OrganizationID)

	apiTokenClient := suite.SetupAPITokenClient(localAdmin.UserCtx, t)
	adminPersonalAccessTokenClient := suite.SetupPatClient(localAdmin, t)

	return &TestOrgUsers{
		Owner:          &localOwner,
		SuperAdmin:     &localSuperAdmin,
		Admin:          &localAdmin,
		Member:         &localMember,
		Auditor:        &localAuditor,
		AdminAPIClient: apiTokenClient,
		AdminPatClient: adminPersonalAccessTokenClient,
	}
}

// AddFunctionalRoleForUser adds the relations for the user in the organization by
// adding the tuples to FGA
func (suite *GraphTestSuite) AddFunctionalRoleForUser(ctx context.Context, t *testing.T, userID, orgID string, relations []string) {
	t.Helper()

	tuples := []fgax.TupleKey{}

	for _, rel := range relations {
		tuple := fgax.TupleKey{
			Subject: fgax.Entity{
				Kind:       fgax.Kind(generated.TypeUser),
				Identifier: userID,
			},
			Object: fgax.Entity{
				Kind:       fgax.Kind(generated.TypeOrganization),
				Identifier: orgID,
			},
			Relation: fgax.Relation(rel),
		}

		tuples = append(tuples, tuple)
	}

	_, err := suite.Client.DB.Authz.WriteTupleKeys(ctx, tuples, nil)
	require.NoError(t, err)
}
