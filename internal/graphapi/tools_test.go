package graphapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/iam/fgax"
	mock_fga "github.com/theopenlane/iam/fgax/mockery"
	"github.com/theopenlane/riverboat/pkg/riverqueue"

	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/totp"
	"github.com/theopenlane/utils/testutils"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/internal/ent/entconfig"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/entdb"
	objmw "github.com/theopenlane/core/internal/middleware/objects"
	"github.com/theopenlane/core/pkg/objects"
	mock_objects "github.com/theopenlane/core/pkg/objects/mocks"
	"github.com/theopenlane/core/pkg/openlaneclient"
	coreutils "github.com/theopenlane/core/pkg/testutils"
)

var (
	testUser          *ent.User
	testPersonalOrgID string
	testOrgID         string
)

// TestGraphTestSuite runs all the tests in the GraphTestSuite
func TestGraphTestSuite(t *testing.T) {
	suite.Run(t, new(GraphTestSuite))
}

// GraphTestSuite handles the setup and teardown between tests
type GraphTestSuite struct {
	suite.Suite
	client *client
	tf     *testutils.TestFixture
}

// client contains all the clients the test need to interact with
type client struct {
	db           *ent.Client
	api          *openlaneclient.OpenlaneClient
	apiWithPAT   *openlaneclient.OpenlaneClient
	apiWithToken *openlaneclient.OpenlaneClient
	fga          *mock_fga.MockSdkClient
	objectStore  *objects.Objects
}

func (suite *GraphTestSuite) SetupSuite() {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	suite.tf = entdb.NewTestFixture()
}

func (suite *GraphTestSuite) SetupTest() {
	t := suite.T()

	ctx := context.Background()

	// setup fga mock
	c := &client{
		fga: mock_fga.NewMockSdkClient(t),
	}

	// create mock FGA client
	fc := fgax.NewMockFGAClient(t, c.fga)

	// setup otp manager
	otpOpts := []totp.ConfigOption{
		totp.WithCodeLength(6),
		totp.WithIssuer("theopenlane"),
		totp.WithSecret(totp.Secret{
			Version: 0,
			Key:     ulids.New().String(),
		}),
	}

	tm, err := coreutils.CreateTokenManager(15 * time.Minute) //nolint:mnd
	require.NoError(t, err)

	sm := coreutils.CreateSessionManager()
	rc := coreutils.NewRedisClient()

	sessionConfig := sessions.NewSessionConfig(
		sm,
		sessions.WithPersistence(rc),
	)

	sessionConfig.CookieConfig = &sessions.DebugOnlyCookieConfig

	otpMan := totp.NewOTP(otpOpts...)

	opts := []ent.Option{
		ent.Authz(*fc),
		ent.Emailer(&emailtemplates.Config{}), // add noop email config
		ent.TOTP(&totp.Manager{
			TOTPManager: otpMan,
		}),
		ent.TokenManager(tm),
		ent.SessionConfig(&sessionConfig),
		ent.EntConfig(&entconfig.Config{
			EntityTypes: []string{"vendor"},
		}),
	}

	// create database connection
	jobOpts := []riverqueue.Option{riverqueue.WithConnectionURI(suite.tf.URI)}

	db, err := entdb.NewTestClient(ctx, suite.tf, jobOpts, opts)
	require.NoError(t, err, "failed opening connection to database")

	c.objectStore, err = coreutils.MockObjectManager(t, objmw.Upload)
	require.NoError(t, err)

	// set the validation function
	c.objectStore.ValidationFunc = objmw.MimeTypeValidator

	// assign values
	c.db = db
	c.api, err = coreutils.TestClient(t, c.db, c.objectStore)
	require.NoError(t, err)

	// create test user
	ctx = echocontext.NewTestContext()
	testUser = (&UserBuilder{client: c}).MustNew(ctx, t)
	testPersonalOrg, err := testUser.Edges.Setting.DefaultOrg(ctx)
	require.NoError(t, err)

	testPersonalOrgID = testPersonalOrg.ID

	userCtx, err := auth.NewTestContextWithOrgID(testUser.ID, testPersonalOrgID)
	require.NoError(t, err)

	// setup a non-personal org
	testOrg := (&OrganizationBuilder{client: c}).MustNew(userCtx, t)
	testOrgID = testOrg.ID

	userCtx, err = userContext()
	require.NoError(t, err)

	// setup client with a personal access token
	pat := (&PersonalAccessTokenBuilder{client: c, OwnerID: testUser.ID, OrganizationIDs: []string{testOrgID, testPersonalOrgID}}).MustNew(userCtx, t)
	authHeaderPAT := openlaneclient.Authorization{
		BearerToken: pat.Token,
	}

	c.apiWithPAT, err = coreutils.TestClientWithAuth(t, c.db, c.objectStore, openlaneclient.WithCredentials(authHeaderPAT))
	require.NoError(t, err)

	// setup client with an API token
	apiToken := (&APITokenBuilder{client: c}).MustNew(userCtx, t)

	authHeaderAPIToken := openlaneclient.Authorization{
		BearerToken: apiToken.Token,
	}
	c.apiWithToken, err = coreutils.TestClientWithAuth(t, c.db, c.objectStore, openlaneclient.WithCredentials(authHeaderAPIToken))
	require.NoError(t, err)

	suite.client = c
}

func (suite *GraphTestSuite) TearDownTest() {
	// clear all fga mocks
	mock_fga.ClearMocks(suite.client.fga)

	err := suite.client.db.Close()
	require.NoError(suite.T(), err)
}

func (suite *GraphTestSuite) TearDownSuite() {
	testutils.TeardownFixture(suite.tf)
}

// userContext creates a new user in the database and returns a context with
// the user claims attached
func userContext() (context.Context, error) {
	return auth.NewTestContextWithOrgID(testUser.ID, testOrgID)
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

// expectUpload sets up the mock object store to expect an upload and related operations
func expectUpload(t *testing.T, mockStore objects.Storage, expectedUploads []graphql.Upload) {
	require.NotNil(t, mockStore)

	ms, ok := mockStore.(*mock_objects.MockStorage)
	require.True(t, ok)

	mockScheme := "file://"

	for _, upload := range expectedUploads {
		ms.EXPECT().GetScheme().Return(&mockScheme).Times(1)
		ms.EXPECT().Upload(mock.Anything, mock.Anything, mock.Anything).Return(&objects.UploadedFileMetadata{
			Size: upload.Size,
		}, nil).Times(1)
		ms.EXPECT().GetPresignedURL(mock.Anything, mock.Anything, mock.Anything).Return("https://presigned.url/my-file", nil).Times(1)
	}
}

// expectUploadCheckOnly sets up the mock object store to expect an upload check only operation
// but fails before the upload is attempted
func expectUploadCheckOnly(t *testing.T, mockStore objects.Storage) {
	require.NotNil(t, mockStore)

	ms, ok := mockStore.(*mock_objects.MockStorage)
	require.True(t, ok)

	mockScheme := "file://"

	ms.EXPECT().GetScheme().Return(&mockScheme).Times(1)
}
