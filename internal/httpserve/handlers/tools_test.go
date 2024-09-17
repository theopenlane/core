package handlers_test

import (
	"context"
	"log"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	mock_fga "github.com/theopenlane/iam/fgax/mockery"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/utils/emails"
	"github.com/theopenlane/utils/marionette"
	"github.com/theopenlane/utils/testutils"

	"github.com/theopenlane/core/internal/ent/entconfig"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/entdb"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/core/pkg/openlaneclient"
	coreutils "github.com/theopenlane/core/pkg/testutils"
)

var (
	// commonly used vars in tests
	emptyResponse = "null\n"
	validPassword = "sup3rs3cu7e!"

	// mock email send settings
	maxWaitInMillis      = 2000
	pollIntervalInMillis = 100
)

// HandlerTestSuite handles the setup and teardown between tests
type HandlerTestSuite struct {
	suite.Suite
	e   *echo.Echo
	db  *ent.Client
	api *openlaneclient.OpenlaneClient
	h   *handlers.Handler
	fga *mock_fga.MockSdkClient
	tf  *testutils.TestFixture
}

// TestHandlerTestSuite runs all the tests in the HandlerTestSuite
func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func (suite *HandlerTestSuite) SetupSuite() {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	suite.tf = entdb.NewTestFixture()
}

func (suite *HandlerTestSuite) SetupTest() {
	t := suite.T()

	ctx := context.Background()

	suite.fga = mock_fga.NewMockSdkClient(t)

	// create mock FGA client
	fc := fgax.NewMockFGAClient(t, suite.fga)

	emConfig := emails.Config{
		Testing:   true,
		Archive:   filepath.Join("fixtures", "emails"),
		FromEmail: "mitb@theopenlane.io",
	}

	em, err := emails.New(emConfig)
	if err != nil {
		t.Fatal("error creating email manager")
	}

	// Start task manager
	tmConfig := marionette.Config{}

	taskMan := marionette.New(tmConfig)

	taskMan.Start()

	tm, err := coreutils.CreateTokenManager(15 * time.Minute) //nolint:mnd
	if err != nil {
		t.Fatal("error creating token manager")
	}

	sm := coreutils.CreateSessionManager()
	rc := coreutils.NewRedisClient()

	sessionConfig := sessions.NewSessionConfig(
		sm,
		sessions.WithPersistence(rc),
	)

	sessionConfig.CookieConfig = &sessions.DebugOnlyCookieConfig

	opts := []ent.Option{
		ent.Authz(*fc),
		ent.JobQueue(river.Client[pgx.Tx]{}), // todo: add a noop job queue
		ent.EmailConfig(&emailtemplates.Config{}),
		ent.TokenManager(tm),
		ent.SessionConfig(&sessionConfig),
		ent.EntConfig(&entconfig.Config{
			Flags: entconfig.Flags{
				UseListUserService:   false,
				UseListObjectService: false,
			},
		}),
	}

	// create database connection
	db, err := entdb.NewTestClient(ctx, suite.tf, opts)
	require.NoError(t, err, "failed opening connection to database")

	// add db to test client
	suite.db = db

	// add the client
	suite.api, err = coreutils.TestClient(t, suite.db)
	require.NoError(t, err)

	// setup handler
	suite.h = handlerSetup(t, suite.db)

	// setup echo router
	suite.e = setupEcho(suite.db)
}

func (suite *HandlerTestSuite) TearDownTest() {
	// clear all fga mocks
	mock_fga.ClearMocks(suite.fga)

	if suite.db != nil {
		if err := suite.db.Close(); err != nil {
			log.Fatalf("failed to close database: %s", err)
		}
	}
}

func (suite *HandlerTestSuite) TearDownSuite() {
	testutils.TeardownFixture(suite.tf)
}

func setupEcho(entClient *ent.Client) *echo.Echo {
	// create echo context with middleware
	e := echo.New()

	transactionConfig := transaction.Client{
		EntDBClient: entClient,
	}

	e.Use(transactionConfig.Middleware)

	return e
}

// handlerSetup to be used for required references in the handler tests
func handlerSetup(t *testing.T, ent *ent.Client) *handlers.Handler {
	as := authmanager.New()
	as.SetTokenManager(ent.TokenManager)
	as.SetSessionConfig(ent.SessionConfig)

	h := &handlers.Handler{
		IsTest:        true,
		TokenManager:  ent.TokenManager,
		DBClient:      ent,
		RedisClient:   ent.SessionConfig.RedisClient,
		SessionConfig: ent.SessionConfig,
		AuthManager:   as,
		EmailConfig:   emailtemplates.Config{},
		JobQueue:      river.Client[pgx.Tx]{}, // todo: add a noop job queue
	}

	return h
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
