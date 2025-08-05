package handlers_test

import (
	"context"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/redis/go-redis/v9"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/iam/fgax"
	fgatest "github.com/theopenlane/iam/fgax/testutils"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/iam/totp"
	"github.com/theopenlane/riverboat/pkg/riverqueue"
	"github.com/theopenlane/utils/testutils"

	"github.com/theopenlane/core/internal/ent/entconfig"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/entdb"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/internal/httpserve/route"
	"github.com/theopenlane/core/internal/httpserve/server"
	objmw "github.com/theopenlane/core/internal/middleware/objects"
	"github.com/theopenlane/core/pkg/entitlements/mocks"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/middleware/cors"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/core/pkg/objects"
	coreutils "github.com/theopenlane/core/pkg/testutils"

	// import generated runtime which is required to prevent cyclical dependencies
	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
)

// TestOperations consolidates all test operations for easier access
type TestOperations struct {
	Account struct {
		Access   *openapi3.Operation
		Roles    *openapi3.Operation
		Features *openapi3.Operation
	}
	Auth struct {
		Login    *openapi3.Operation
		Register *openapi3.Operation
		Refresh  *openapi3.Operation
	}
	Organization struct {
		Switch *openapi3.Operation
		Invite *openapi3.Operation
	}
	Email struct {
		Verify *openapi3.Operation
		Resend *openapi3.Operation
	}
}

var (
	// commonly used vars in tests
	emptyResponse    = "null\n"
	validPassword    = "sup3rs3cu7e!"
	otpManagerSecret = totp.Secret{
		Version: 0,
		Key:     "9f0c6da662f018b58b04a093e2dbb2e1",
	}
	webhookSecret = "whsec_test_secret"
)

const (
	fgaModelFile = "../../../fga/model/model.fga"
)

// HandlerTestSuite handles the setup and teardown between tests
type HandlerTestSuite struct {
	suite.Suite
	e                    *echo.Echo
	db                   *ent.Client
	api                  *testclient.TestClient
	h                    *handlers.Handler
	router               *route.Router
	fga                  *fgax.Client
	tf                   *testutils.TestFixture
	ofgaTF               *fgatest.OpenFGATestFixture
	stripeMockBackend    *mocks.MockStripeBackend
	objectStore          *objects.Objects
	sharedTokenManager   *tokens.TokenManager
	sharedRedisClient    *redis.Client
	sharedSessionManager sessions.Store[map[string]any]
	sharedFGAClient      *fgax.Client
	sharedOTPManager     *totp.Client
	sharedPondPool       *soiree.PondPool

	// OpenAPI operations for reuse in tests
	startImpersonationOp *openapi3.Operation
	endImpersonationOp   *openapi3.Operation
}

// TestHandlerTestSuite runs all the tests in the HandlerTestSuite
func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func (suite *HandlerTestSuite) SetupSuite() {
	if testing.Verbose() {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}

	// setup db container
	suite.tf = entdb.NewTestFixture()

	// setup openFGA container
	suite.ofgaTF = fgatest.NewFGATestcontainer(context.Background(), fgatest.WithModelFile(fgaModelFile))

	// create shared instances once to avoid expensive recreation in each test
	var err error

	// shared token manager to avoid RSA key generation
	suite.sharedTokenManager, err = coreutils.CreateTokenManager(15 * time.Minute) //nolint:mnd
	require.NoError(suite.T(), err)

	// shared redis client to avoid miniredis server startup
	suite.sharedRedisClient = coreutils.NewRedisClient()

	// shared session manager to avoid random key generation
	suite.sharedSessionManager = coreutils.CreateSessionManager()

	// shared FGA client to avoid repeated container connections
	suite.sharedFGAClient, err = suite.ofgaTF.NewFgaClient(context.Background())
	require.NoError(suite.T(), err)

	// shared OTP manager
	otpOpts := []totp.ConfigOption{
		totp.WithCodeLength(6),
		totp.WithIssuer("authenticator.local"),
		totp.WithSecret(otpManagerSecret),
		totp.WithRedis(suite.sharedRedisClient),
	}
	otpMan := totp.NewOTP(otpOpts...)
	suite.sharedOTPManager = &totp.Client{
		Manager: otpMan,
	}

	// shared pond pool to avoid worker pool creation
	suite.sharedPondPool = soiree.NewPondPool(
		soiree.WithMaxWorkers(100), //nolint:mnd
		soiree.WithName("ent_client_pool"),
	)
}

func (suite *HandlerTestSuite) SetupTest() {
	t := suite.T()

	ctx := context.Background()

	// use all shared instances to avoid expensive recreation
	sessionConfig := sessions.NewSessionConfig(
		suite.sharedSessionManager,
		sessions.WithPersistence(suite.sharedRedisClient),
	)

	sessionConfig.CookieConfig = sessions.DebugOnlyCookieConfig

	opts := []ent.Option{
		ent.Authz(*suite.sharedFGAClient),
		ent.Emailer(&emailtemplates.Config{
			CompanyName: "Meow Inc.",
		}),
		ent.TokenManager(suite.sharedTokenManager),
		ent.SessionConfig(&sessionConfig),
		ent.EntConfig(&entconfig.Config{}),
		ent.TOTP(suite.sharedOTPManager),
		ent.PondPool(suite.sharedPondPool),
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
	suite.api, err = coreutils.TestClient(suite.db, suite.objectStore)
	require.NoError(t, err)

	// setup router with schema registry
	suite.router, err = setupRouter()
	require.NoError(t, err)

	// setup handler
	suite.h = handlerSetup(suite.db)

	// use shared OTP manager
	suite.h.OTPManager = suite.sharedOTPManager

	// setup echo router with transaction middleware
	suite.e = suite.router.Echo

	// Add transaction middleware to router's echo instance for tests
	transactionConfig := transaction.Client{
		EntDBClient: suite.db,
	}
	suite.e.Use(transactionConfig.Middleware)

	// Setup reusable OpenAPI operations
	suite.startImpersonationOp = suite.createImpersonationOperation("StartImpersonationHandler", "Test start impersonation")
	suite.endImpersonationOp = suite.createImpersonationOperation("EndImpersonationHandler", "Test end impersonation")

	suite.setupTestData(ctx)
}

// createImpersonationOperation creates a reusable OpenAPI operation for impersonation tests
func (suite *HandlerTestSuite) createImpersonationOperation(operationID, description string) *openapi3.Operation {
	operation := openapi3.NewOperation()
	operation.Description = description
	operation.Tags = []string{"impersonation"}
	operation.OperationID = operationID
	operation.Security = handlers.BearerSecurity()
	return operation
}

// registerTestHandler is a helper to register test handlers with OpenAPI context
func (suite *HandlerTestSuite) registerTestHandler(method, path string, operation *openapi3.Operation, handlerFunc func(echo.Context, *handlers.OpenAPIContext) error) {
	suite.e.Add(method, path, func(c echo.Context) error {
		return handlerFunc(c, &handlers.OpenAPIContext{
			Operation: operation,
			Registry:  suite.router.SchemaRegistry,
		})
	})
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

func setupRouter() (*route.Router, error) {
	// Create a test router with proper schema registry setup
	return server.NewRouter(server.LogConfig{
		PrettyLog: true,
		LogLevel:  1, // INFO level
	})
}

func setupEcho(dbClient *ent.Client) *echo.Echo {
	// create echo context with middleware
	e := echo.New()

	transactionConfig := transaction.Client{
		EntDBClient: dbClient,
	}

	e.Use(transactionConfig.Middleware)

	coreMW := cors.MustNew([]string{"*"})
	e.Use(coreMW)

	return e
}

// handlerSetup to be used for required references in the handler tests
func handlerSetup(db *ent.Client) *handlers.Handler {
	as := authmanager.New(db)

	h := &handlers.Handler{
		IsTest:        true,
		TokenManager:  db.TokenManager,
		DBClient:      db,
		RedisClient:   db.SessionConfig.RedisClient,
		SessionConfig: db.SessionConfig,
		AuthManager:   as,
		Entitlements:  db.EntitlementManager,
		OauthProvider: handlers.OauthProviderConfig{
			RedirectURL: "http://localhost",
		},
		DefaultTrustCenterDomain: "trust.openlane.com",
	}

	return h
}
