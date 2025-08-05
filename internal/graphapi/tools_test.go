package graphapi_test

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"gotest.tools/v3/assert"

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
	"github.com/theopenlane/core/internal/graphapi/testclient"
	objmw "github.com/theopenlane/core/internal/middleware/objects"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/objects"
	mock_objects "github.com/theopenlane/core/pkg/objects/mocks"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/core/pkg/summarizer"
	coreutils "github.com/theopenlane/core/pkg/testutils"

	// import generated runtime which is required to prevent cyclical dependencies
	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
)

const (
	fgaModelFile = "../../fga/model/model.fga"

	redacted = "*****************************"

	// common error message strings
	notFoundErrorMsg      = "not found"
	notExistsErrorMsg     = "does not exist"
	notAuthorizedErrorMsg = "you are not authorized to perform this action"
	couldNotFindUser      = "could not identify authenticated user in request"
)

// GraphTestSuite handles the setup and teardown between tests
type GraphTestSuite struct {
	client *client
	tf     *testutils.TestFixture
	ofgaTF *fgatest.OpenFGATestFixture
}

// client contains all the clients the test need to interact with
type client struct {
	db           *ent.Client
	api          *testclient.TestClient
	apiWithPAT   *testclient.TestClient
	apiWithToken *testclient.TestClient
	fga          *fgax.Client
	objectStore  *objects.Objects
}

var suite = &GraphTestSuite{}

func TestMain(m *testing.M) {
	flag.Parse()

	// Create a new testing.T instance
	// Note: this is only to seed data; you should not use this instance for actual tests
	t := &testing.T{}

	// Setup code here (e.g., initialize database connection)
	suite.SetupSuite(t)

	// Setup test data, most tests can reuse this same data
	suite.setupTestData(context.Background(), t)

	// Run the tests
	exitCode := m.Run()

	// Teardown code here (e.g., close database connection)
	suite.TearDownSuite(t)

	// Exit with the result of the tests
	os.Exit(exitCode)
}

func (suite *GraphTestSuite) SetupSuite(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	if testing.Verbose() {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}

	// setup db container
	suite.tf = entdb.NewTestFixture()

	// setup openFGA container
	suite.ofgaTF = fgatest.NewFGATestcontainer(context.Background(), fgatest.WithModelFile(fgaModelFile))

	ctx := context.Background()

	// setup fga client
	fgaClient, err := suite.ofgaTF.NewFgaClient(ctx)
	assert.NilError(t, err)

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
	assert.NilError(t, err)

	sm := coreutils.CreateSessionManager()
	rc := coreutils.NewRedisClient()

	sessionConfig := sessions.NewSessionConfig(
		sm,
		sessions.WithPersistence(rc),
	)

	sessionConfig.CookieConfig = sessions.DebugOnlyCookieConfig

	otpMan := totp.NewOTP(otpOpts...)

	entCfg := &entconfig.Config{
		EntityTypes: []string{"vendor"},
		Summarizer: entconfig.Summarizer{
			Type:             entconfig.SummarizerTypeLexrank,
			MaximumSentences: 60,
		},
	}

	summarizerClient, err := summarizer.NewSummarizer(*entCfg)
	assert.NilError(t, err)

	pool := soiree.NewPondPool(
		soiree.WithMaxWorkers(100), //nolint:mnd
		soiree.WithName("ent_client_pool"),
	)

	opts := []ent.Option{
		ent.Authz(*fgaClient),
		ent.Emailer(&emailtemplates.Config{}), // add noop email config
		ent.TOTP(&totp.Client{
			Manager: otpMan,
		}),
		ent.TokenManager(tm),
		ent.SessionConfig(&sessionConfig),
		ent.EntConfig(entCfg),
		ent.Summarizer(summarizerClient),
		ent.PondPool(pool),
	}

	// create database connection
	jobOpts := []riverqueue.Option{riverqueue.WithConnectionURI(suite.tf.URI)}

	db, err := entdb.NewTestClient(ctx, suite.tf, jobOpts, opts)
	assert.NilError(t, err)

	c.objectStore, err = coreutils.MockObjectManager(t, objmw.Upload)
	assert.NilError(t, err)

	// set the validation function
	c.objectStore.ValidationFunc = objmw.MimeTypeValidator

	// assign values
	c.db = db
	c.api, err = coreutils.TestClient(c.db, c.objectStore)
	assert.NilError(t, err)

	suite.client = c
}

func (suite *GraphTestSuite) TearDownSuite(t *testing.T) {
	// close the database connection
	err := suite.client.db.Close()
	assert.NilError(t, err)

	// close the database container
	testutils.TeardownFixture(suite.tf)

	// terminate all fga containers
	err = suite.ofgaTF.TeardownFixture()
	assert.NilError(t, err)
}

// expectUpload sets up the mock object store to expect an upload and related operations
func expectUpload(t *testing.T, mockStore objects.Storage, expectedUploads []graphql.Upload) {
	assert.Assert(t, mockStore != nil)

	ms, ok := mockStore.(*mock_objects.MockStorage)
	assert.Assert(t, ok)

	mockScheme := "file://"

	for _, upload := range expectedUploads {
		ms.EXPECT().GetScheme().Return(&mockScheme).Times(1)
		ms.EXPECT().Upload(mock.Anything, mock.Anything, mock.Anything).Return(&objects.UploadedFileMetadata{
			Size: upload.Size,
		}, nil).Times(1)
	}
}

// expectUploadNillable sets up the mock object store to expect an upload and related operations
func expectUploadNillable(t *testing.T, mockStore objects.Storage, expectedUploads []*graphql.Upload) {
	assert.Check(t, mockStore != nil)

	ms, ok := mockStore.(*mock_objects.MockStorage)
	assert.Assert(t, ok)

	mockScheme := "file://"

	for _, upload := range expectedUploads {
		ms.EXPECT().GetScheme().Return(&mockScheme).Times(1)
		ms.EXPECT().Upload(mock.Anything, mock.Anything, mock.Anything).Return(&objects.UploadedFileMetadata{
			Size: upload.Size,
		}, nil).Times(1)
	}
}

// expectUploadCheckOnly sets up the mock object store to expect an upload check only operation
// but fails before the upload is attempted
func expectUploadCheckOnly(t *testing.T, mockStore objects.Storage) {
	assert.Assert(t, mockStore != nil)

	ms, ok := mockStore.(*mock_objects.MockStorage)
	assert.Assert(t, ok)

	mockScheme := "file://"

	ms.EXPECT().GetScheme().Return(&mockScheme).Times(1)
}

// parseClientError parses the error response from the client and returns a slice of gqlerror.Error
func parseClientError(t *testing.T, err error) []*gqlerror.Error {
	t.Helper()

	if err == nil {
		return nil
	}

	errResp, ok := err.(*clientv2.ErrorResponse)
	assert.Check(t, ok)
	assert.Check(t, errResp.HasErrors())

	gqlErrors := []*gqlerror.Error{}

	errors := errResp.GqlErrors.Unwrap()

	for _, e := range errors {
		customErr, ok := e.(*gqlerror.Error)
		assert.Check(t, ok)
		gqlErrors = append(gqlErrors, customErr)
	}

	return gqlErrors
}

// assertErrorCode checks if the error code matches the expected code
func assertErrorCode(t *testing.T, err *gqlerror.Error, code string) {
	t.Helper()

	assert.Equal(t, code, openlaneclient.GetErrorCode(err))
}

// assertErrorMessage checks if the error message matches the expected message
func assertErrorMessage(t *testing.T, err *gqlerror.Error, msg string) {
	t.Helper()

	assert.Equal(t, msg, openlaneclient.GetErrorMessage(err))
}
