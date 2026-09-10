//go:build test

package testharness

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/mcuadros/go-defaults"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/fgax"
	fgatest "github.com/theopenlane/iam/fgax/testutils"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/riverboat/pkg/riverqueue"

	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/totp"
	"github.com/theopenlane/utils/testutils"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/fga/fgaversion"
	"github.com/theopenlane/core/v2/internal/ent/entconfig"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/ent/validator"
	"github.com/theopenlane/core/v2/internal/entdb"
	"github.com/theopenlane/core/v2/internal/graphapi"
	"github.com/theopenlane/core/v2/internal/graphapi/common"
	gqlgenerated "github.com/theopenlane/core/v2/internal/graphapi/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/core/v2/internal/httpserve/config"
	emaildef "github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	slackdef "github.com/theopenlane/core/v2/internal/integrations/definitions/slack"
	systemdef "github.com/theopenlane/core/v2/internal/integrations/definitions/system"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	intruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/internal/keystore"
	"github.com/theopenlane/core/v2/internal/objects"
	"github.com/theopenlane/core/v2/internal/objects/validators"
	coreutils "github.com/theopenlane/core/v2/internal/testutils"
	testint "github.com/theopenlane/core/v2/internal/testutils/integrations"
	"github.com/theopenlane/core/v2/internal/workflows/engine"
	"github.com/theopenlane/core/v2/pkg/entitlements/mocks"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
	authmw "github.com/theopenlane/core/v2/pkg/middleware/auth"
	mock_shared "github.com/theopenlane/core/v2/pkg/objects/mocks"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/core/v2/pkg/summarizer"

	// import generated runtime which is required to prevent cyclical dependencies
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	_ "github.com/theopenlane/core/v2/internal/ent/generated/runtime"
	_ "github.com/theopenlane/core/v2/internal/ent/historygenerated/runtime"
)

const (
	Redacted = "*****************************"

	// common error message strings
	NotFoundErrorMsg         = "not found"
	NotAuthorizedErrorMsg    = "you are not authorized to perform this action"
	MissingScopeErrorMsg     = "lacks the required scopes"
	InvalidInputErrorMsg     = "invalid input"
	seedStripeSubscriptionID = "sub_test_subscription"
	webhookSecret            = "whsec_test_secret"

	MappableDomainZoneTestID = "mappable-domain-zone-id"
	CnameTargetTest          = "cname-target.test.com"
	PreviewCnameTargetTest   = "preview-cname-target.test.com"
	DefaultDomainTest        = "test.default.domain"
)

// GraphTestSuite handles the setup and teardown between tests
type GraphTestSuite struct {
	Client             *Client
	TF                 *testutils.TestFixture
	OFGATF             *fgatest.OpenFGATestFixture
	StripeMockBackend  *mocks.MockStripeBackend
	CacheRefreshServer *httptest.Server
	GalaRuntime        *gala.Gala
	IntegrationsRT     *intruntime.Runtime
	WorkflowEngine     *engine.WorkflowEngine
}

// Client contains all the clients the test need to interact with
type Client struct {
	DB                 *ent.Client
	API                *testclient.TestClient
	APIWithPAT         *testclient.TestClient
	APIWithToken       *testclient.TestClient
	APIWithTokenOrg2   *testclient.TestClient
	FGA                *fgax.Client
	ObjectStore        *objects.Service
	MockProvider       *mock_shared.MockProvider
	DeletedStorageKeys *DeletedKeys
}

// repoRoot resolves the repository root from this file's own location so paths below do
// not depend on the working directory of whichever suite is running
func repoRoot() string {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("unable to determine testharness source location")
	}

	return filepath.Join(filepath.Dir(thisFile), "..", "..", "..")
}

// fgaModuleFile is the fga model the test containers are seeded with
var fgaModuleFile = filepath.Join(repoRoot(), "fga", "model", "fga.mod")

// Suite is the shared harness instance used by each graphapi test suite
var Suite = &GraphTestSuite{}

func (suite *GraphTestSuite) SetupSuite(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	if testing.Verbose() {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}

	// setup test server for cache refresh requests
	suite.CacheRefreshServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Extract host from test server URL (e.g., "127.0.0.1:12345" from "http://127.0.0.1:12345")
	testServerHost := suite.CacheRefreshServer.URL[len("http://"):]

	hooks.SetTrustCenterConfig(hooks.TrustCenterConfig{
		CacheRefreshScheme:       "http",
		DefaultTrustCenterDomain: testServerHost,
	})

	// setup db container
	suite.TF = entdb.NewTestFixture()

	version, err := fgaversion.GetVersion()
	RequireNoError(t, err)

	// setup openFGA container
	suite.OFGATF = fgatest.NewFGATestcontainer(context.Background(),
		fgatest.WithModuleFile(fgaModuleFile),
		fgatest.WithEnvVars(coreutils.GetDefaultFGAEnvs()),
		fgatest.WithVersion(version),
	)

	ctx := context.Background()

	// setup fga client
	fgaClient, err := suite.OFGATF.NewFgaClient(ctx)
	RequireNoError(t, err)

	c := &Client{
		FGA: fgaClient,
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

	tm, err := coreutils.CreateTokenManager(-15 * time.Minute) //nolint:mnd
	RequireNoError(t, err)

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
		Summarizer: summarizer.Config{
			Type:             summarizer.TypeLexrank,
			MaximumSentences: 60,
		},
		Modules: entconfig.Modules{
			Enabled: true,
		},
		EmailValidation: validator.EmailVerificationConfig{
			Enabled:           true,
			AllowedEmailTypes: validator.AllowedEmailTypes{Free: false},
		},
	}

	// we want the email validator to error if a free email domain is used
	// in org settings, but we don't want to error all user creations on email validation
	ev := entCfg.EmailValidation.NewVerifier()

	// now disable email validation for tests so that we don't have to make real email addresses
	entCfg.EmailValidation.Enabled = false

	summarizerClient, err := summarizer.NewSummarizer(entCfg.Summarizer)
	RequireNoError(t, err)

	pool := gala.NewPool(
		gala.WithWorkers(200), //nolint:mnd
		gala.WithPoolName("ent_client_pool"),
	)

	// setup history client
	hc, err := entdb.NewTestHistoryClient(ctx, suite.TF)
	RequireNoError(t, err)

	// setup mock entitlements client
	entitlements, err := suite.MockStripeClient()
	RequireNoError(t, err)

	c.ObjectStore, c.MockProvider, err = coreutils.MockStorageServiceWithValidationAndProvider(t, nil, validators.MimeTypeValidator)
	RequireNoError(t, err)

	c.DeletedStorageKeys = NewDeletedKeys()

	c.MockProvider.On("GetPresignedURL", mock.Anything, mock.Anything, mock.Anything).Return("file:///tmp/test-presigned", nil).Maybe()
	c.MockProvider.On("Download", mock.Anything, mock.Anything, mock.Anything).Return(&storage.DownloadedMetadata{
		File: TestPDFBytes(),
		Size: 1024,
	}, nil).Maybe()

	// record the storage keys removed so tests can assert the objects backing deleted files are
	// cleaned up out of object storage and not just orphaned
	c.MockProvider.On("Delete", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		f, ok := args.Get(1).(*storagetypes.File)
		if !ok || f == nil {
			return
		}

		c.DeletedStorageKeys.Add(f.Key)
	}).Return(nil).Maybe()

	opts := []ent.Option{
		ent.Authz(*fgaClient),
		ent.TOTP(&totp.Client{
			Manager: otpMan,
		}),
		ent.TokenManager(tm),
		ent.SessionConfig(&sessionConfig),
		ent.EntConfig(entCfg),
		ent.Summarizer(summarizerClient),
		ent.Pool(pool),
		ent.EntitlementManager(entitlements),
		ent.EmailVerifier(ev),
		ent.HistoryClient(hc),
		ent.ObjectManager(c.ObjectStore),
	}

	// create database connection
	jobOpts := []riverqueue.Option{riverqueue.WithConnectionURI(suite.TF.URI)}

	db, err := entdb.NewTestClient(ctx, suite.TF, jobOpts, nil, opts)
	RequireNoError(t, err)

	// assign values
	c.DB = db
	c.API, err = coreutils.TestClient(c.DB, c.ObjectStore)
	RequireNoError(t, err)

	// durable gala runtime for integration dispatch
	galaInstance, err := gala.NewGala(ctx, gala.Config{
		DispatchMode:      gala.DispatchModeDurable,
		ConnectionURI:     suite.TF.URI,
		QueueName:         "graphapi_integration_test",
		WorkerCount:       5, //nolint:mnd
		RunMigrations:     true,
		FetchCooldown:     time.Millisecond,
		FetchPollInterval: 10 * time.Millisecond, //nolint:mnd
	})
	RequireNoError(t, err)

	db.Use(hooks.EmitGalaEventHook(galaInstance))

	wfEngine, err := engine.NewWorkflowEngine(c.DB, galaInstance)
	RequireNoError(t, err)

	RequireNoError(t, galaInstance.Attach(
		gala.WithValue(galaInstance),
		gala.WithValue(c.DB),
		gala.WithValue(entitlements),
		gala.WithValue(wfEngine),
		// without the restored ent client every durable mutation listener fails
		gala.WithRestoredValue("ent_client", ent.NewContext),
	))

	_, err = gala.Register(galaInstance, hooks.EntitlementListeners()...)
	RequireNoError(t, err)

	_, err = gala.Register(galaInstance, hooks.NDAAttestationListeners()...)
	RequireNoError(t, err)

	_, err = gala.Register(galaInstance, hooks.OrganizationCleanupListeners()...)
	RequireNoError(t, err)

	_, err = gala.Register(galaInstance, hooks.IntegrationCleanupListeners()...)
	RequireNoError(t, err)

	// wire integration runtime with mock email provider
	credStore, err := keystore.NewStore(c.DB)
	RequireNoError(t, err)

	rt, err := intruntime.New(intruntime.Config{
		DB:          c.DB,
		Gala:        galaInstance,
		Keystore:    credStore,
		RedisClient: coreutils.NewRedisClient(),
		DefinitionBuilders: []registry.Builder{
			emaildef.Builder(emaildef.MockRuntimeConfig(), false),
			slackdef.Builder(slackdef.Config{}, &slackdef.RuntimeSlackConfig{WebhookURL: "https://hooks.slack.com/services/test/mock/url"}, false),
			systemdef.Builder(systemdef.PaymentReminderConfig{}, systemdef.OrganizationDeleteConfig{}, systemdef.IntegrationLifecycleConfig{}),
			testint.Builder(),
		},
	})
	RequireNoError(t, err)

	intruntime.SetDefault(rt)
	suite.IntegrationsRT = rt

	// cleanup/reseed listeners resolve the runtime from the gala injector as in production
	RequireNoError(t, galaInstance.Attach(gala.WithValue(rt)))
	RequireNoError(t, wfEngine.SetIntegrationDeps(engine.IntegrationDeps{Runtime: rt}))

	// Start workers after attaching all shared dependencies
	RequireNoError(t, galaInstance.StartWorkers(ctx))

	suite.GalaRuntime = galaInstance
	suite.WorkflowEngine = wfEngine

	// Set trust center config for hooks
	hooks.SetTrustCenterConfig(hooks.TrustCenterConfig{
		CnameTarget:              CnameTargetTest,
		PreviewCnameTarget:       PreviewCnameTargetTest,
		DefaultTrustCenterDomain: DefaultDomainTest,
	})

	_, err = c.DB.MappableDomain.Create().
		SetName(PreviewCnameTargetTest).
		SetZoneID(MappableDomainZoneTestID).
		Save(privacy.DecisionContext(ctx, privacy.Allow))
	RequireNoError(t, err)

	suite.Client = c
}

func (suite *GraphTestSuite) TearDownSuite(t *testing.T) {
	if suite.GalaRuntime != nil {
		err := suite.GalaRuntime.StopWorkers(context.Background())
		RequireNoError(t, err)

		err = suite.GalaRuntime.Close()
		RequireNoError(t, err)
	}

	// close the database connection
	err := suite.Client.DB.Close()
	RequireNoError(t, err)

	// close the database container
	testutils.TeardownFixture(suite.TF)

	// terminate all fga containers
	err = suite.OFGATF.TeardownFixture()
	RequireNoError(t, err)

	// close the cache refresh test server
	if suite.CacheRefreshServer != nil {
		suite.CacheRefreshServer.Close()
	}
}

// NewTestGraphServer creates a new GraphQL server for testing
// this is used when the test client can't be used such as subscriptions
func NewTestGraphServer(t *testing.T) http.Handler {
	cfg := config.Config{}
	defaults.SetDefaults(&cfg)

	// get keys from the token manager
	keys, err := Suite.Client.DB.TokenManager.Keys()
	require.NoError(t, err)

	// local validator to avoid JWK cache issues
	validator := tokens.NewJWKSValidator(keys, "http://localhost:17608", "http://localhost:17608")

	r := graphapi.NewResolver(Suite.Client.DB, nil).
		WithExtensions(true).
		WithDevelopment(true).
		WithSubscriptions(true, nil).
		WithAuthOptions(
			authmw.WithSkipperFunc(
				func(c echo.Context) bool {
					return authmw.AuthenticateSkipperFuncForWebsockets(c)
				},
			),
			authmw.WithDBClient(Suite.Client.DB),
			authmw.WithValidator(validator),
		)

	r.WithPool(10)

	c := &gqlgenerated.Config{Resolvers: r}

	srv := handler.New(gqlgenerated.NewExecutableSchema(
		*c,
	))

	srv.AddTransport(transport.GET{})
	srv.AddTransport(r.CreateWebsocketClient())

	// add test case extension to signal tests on after cancel
	testCaseExtension(srv)

	// add common extensions
	common.AddAllExtensions(srv)

	return srv
}

// TestAfterCancel is used to signal when a response is returned after context cancellation in tests
var TestAfterCancel = make(chan struct{}, 1)

// testCaseExtension is used to signal tests when a response is returned after context cancellation
func testCaseExtension(h *handler.Server) {
	h.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		resp := next(ctx)
		if resp != nil {
			// Signal the test that a response was returned after cancellation
			select {
			case TestAfterCancel <- struct{}{}:
			default:

			}

			logx.FromContext(ctx).Warn().Msg("response returned after context cancelled in test case extension, returning nil response to close connection")

			return nil
		}

		return resp
	})
}
