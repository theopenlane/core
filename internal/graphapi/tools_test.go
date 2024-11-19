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
	fgatest "github.com/theopenlane/iam/fgax/testutils"
	"github.com/theopenlane/riverboat/pkg/riverqueue"

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

const (
	fgaModelFile = "../../fga/model/model.fga"

	redacted = "*****************************"

	// common error message strings
	notFoundErrorMsg      = "not found"
	notAuthorizedErrorMsg = "you are not authorized to perform this action"
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
	ofgaTF *fgatest.OpenFGATestFixture
}

// client contains all the clients the test need to interact with
type client struct {
	db           *ent.Client
	api          *openlaneclient.OpenlaneClient
	apiWithPAT   *openlaneclient.OpenlaneClient
	apiWithToken *openlaneclient.OpenlaneClient
	fga          *fgax.Client
	objectStore  *objects.Objects
}

func (suite *GraphTestSuite) SetupSuite() {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	// setup db container
	suite.tf = entdb.NewTestFixture()

	// setup openFGA container
	suite.ofgaTF = fgatest.NewFGATestcontainer(context.Background(), fgatest.WithModelFile(fgaModelFile))
}

func (suite *GraphTestSuite) SetupTest() {
	t := suite.T()

	ctx := context.Background()

	// setup fga client
	fgaClient, err := suite.ofgaTF.NewFgaClient(ctx)
	require.NoError(t, err)

	c := &client{
		fga: fgaClient,
	}

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
		ent.Authz(*fgaClient),
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

	suite.client = c

	// setup test data
	suite.setupTestData(ctx)
}

func (suite *GraphTestSuite) TearDownTest() {
	err := suite.client.db.Close()
	require.NoError(suite.T(), err)
}

func (suite *GraphTestSuite) TearDownSuite() {
	testutils.TeardownFixture(suite.tf)

	// terminate all fga containers
	err := suite.ofgaTF.TeardownFixture()
	require.NoError(suite.T(), err)
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
