package handlers_test

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/iam/fgax"
	fgatest "github.com/theopenlane/iam/fgax/testutils"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/riverboat/pkg/riverqueue"
	"github.com/theopenlane/utils/testutils"

	"github.com/theopenlane/core/internal/ent/entconfig"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/entdb"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	objmw "github.com/theopenlane/core/internal/middleware/objects"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/openlaneclient"
	coreutils "github.com/theopenlane/core/pkg/testutils"
)

var (
	// commonly used vars in tests
	emptyResponse = "null\n"
	validPassword = "sup3rs3cu7e!"
)

const (
	fgaModelFile = "../../../fga/model/model.fga"
)

// HandlerTestSuite handles the setup and teardown between tests
type HandlerTestSuite struct {
	suite.Suite
	e           *echo.Echo
	db          *ent.Client
	api         *openlaneclient.OpenlaneClient
	h           *handlers.Handler
	fga         *fgax.Client
	tf          *testutils.TestFixture
	ofgaTF      *fgatest.OpenFGATestFixture
	objectStore *objects.Objects
}

// TestHandlerTestSuite runs all the tests in the HandlerTestSuite
func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func (suite *HandlerTestSuite) SetupSuite() {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	// setup db container
	suite.tf = entdb.NewTestFixture()

	// setup openFGA container
	suite.ofgaTF = fgatest.NewFGATestcontainer(context.Background(), fgatest.WithModelFile(fgaModelFile))
}

func (suite *HandlerTestSuite) SetupTest() {
	t := suite.T()

	ctx := context.Background()

	// setup fga client
	fgaClient, err := suite.ofgaTF.NewFgaClient(ctx)
	require.NoError(t, err)

	tm, err := coreutils.CreateTokenManager(15 * time.Minute) //nolint:mnd
	require.NoError(t, err)

	sm := coreutils.CreateSessionManager()
	rc := coreutils.NewRedisClient()

	sessionConfig := sessions.NewSessionConfig(
		sm,
		sessions.WithPersistence(rc),
	)

	sessionConfig.CookieConfig = &sessions.DebugOnlyCookieConfig

	opts := []ent.Option{
		ent.Authz(*fgaClient),
		ent.Emailer(&emailtemplates.Config{}),
		ent.TokenManager(tm),
		ent.SessionConfig(&sessionConfig),
		ent.EntConfig(&entconfig.Config{}),
	}

	// create database connection
	jobOpts := []riverqueue.Option{riverqueue.WithConnectionURI(suite.tf.URI)}

	db, err := entdb.NewTestClient(ctx, suite.tf, jobOpts, opts)
	require.NoError(t, err, "failed opening connection to database")

	suite.objectStore, err = coreutils.MockObjectManager(t, objmw.Upload)
	require.NoError(t, err)

	// truncate river tables
	err = db.Job.TruncateRiverTables(ctx)
	require.NoError(t, err)

	// add db to test client
	suite.db = db

	// add the client
	suite.api, err = coreutils.TestClient(t, suite.db, suite.objectStore)
	require.NoError(t, err)

	// setup handler
	suite.h = handlerSetup(t, suite.db)

	// setup echo router
	suite.e = setupEcho(suite.db)

	suite.setupTestData(ctx)
}

func (suite *HandlerTestSuite) TearDownTest() {
	if suite.db != nil {
		err := suite.db.CloseAll()
		require.NoError(suite.T(), err)
	}
}

func (suite *HandlerTestSuite) ClearTestData() {
	err := suite.db.Job.TruncateRiverTables(context.Background())
	require.NoError(suite.T(), err)
}

func (suite *HandlerTestSuite) TearDownSuite() {
	testutils.TeardownFixture(suite.tf)

	// terminate all fga containers
	err := suite.ofgaTF.TeardownFixture()
	require.NoError(suite.T(), err)
}

func setupEcho(dbClient *ent.Client) *echo.Echo {
	// create echo context with middleware
	e := echo.New()

	transactionConfig := transaction.Client{
		EntDBClient: dbClient,
	}

	e.Use(transactionConfig.Middleware)

	return e
}

// handlerSetup to be used for required references in the handler tests
func handlerSetup(t *testing.T, db *ent.Client) *handlers.Handler {
	as := authmanager.New()
	as.SetTokenManager(db.TokenManager)
	as.SetSessionConfig(db.SessionConfig)

	h := &handlers.Handler{
		IsTest:        true,
		TokenManager:  db.TokenManager,
		DBClient:      db,
		RedisClient:   db.SessionConfig.RedisClient,
		SessionConfig: db.SessionConfig,
		AuthManager:   as,
	}

	return h
}
