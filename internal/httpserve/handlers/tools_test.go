//go:build test

package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/stripe/stripe-go/v86"

	"github.com/redis/go-redis/v9"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/fgax"
	fgatest "github.com/theopenlane/iam/fgax/testutils"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/iam/totp"
	"github.com/theopenlane/riverboat/pkg/riverqueue"
	"github.com/theopenlane/utils/testutils"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/v2/fga/fgaversion"
	"github.com/theopenlane/core/v2/internal/ent/entconfig"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/entdb"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/core/v2/internal/httpserve/authmanager"
	"github.com/theopenlane/core/v2/internal/httpserve/handlers"
	"github.com/theopenlane/core/v2/internal/httpserve/route"
	"github.com/theopenlane/core/v2/internal/httpserve/server"
	mockprovider "github.com/theopenlane/newman/providers/mock"

	"github.com/theopenlane/core/v2/internal/integrations/definitions/catalog"
	emaildef "github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/githubapp"
	definitionscim "github.com/theopenlane/core/v2/internal/integrations/definitions/scim"
	slackdef "github.com/theopenlane/core/v2/internal/integrations/definitions/slack"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/keystore"
	"github.com/theopenlane/core/v2/internal/objects"
	coreutils "github.com/theopenlane/core/v2/internal/testutils"
	"github.com/theopenlane/core/v2/pkg/entitlements"
	"github.com/theopenlane/core/v2/pkg/entitlements/mocks"
	"github.com/theopenlane/core/v2/pkg/gala"
	authmiddleware "github.com/theopenlane/core/v2/pkg/middleware/auth"
	"github.com/theopenlane/core/v2/pkg/middleware/transaction"

	_ "github.com/theopenlane/core/v2/internal/ent/generated/runtime"
	_ "github.com/theopenlane/core/v2/internal/ent/historygenerated/runtime"
)

var (
	emptyResponse    = "null\n"
	validPassword    = "sup3rs3cu7e!"
	otpManagerSecret = totp.Secret{
		Version: 0,
		Key:     "9f0c6da662f018b58b04a093e2dbb2e1",
	}
	webhookSecret = "whsec_test_secret"
)

const (
	fgaModuleFile                   = "../../../fga/model/fga.mod"
	seedStripeSubscriptionID        = "sub_test_subscription"
	cnameTargetTest                 = "cname.test.net"
	previewMappableDomainZoneIDTest = "preview-zone-id"
)

// HandlerTestSuite handles the setup and teardown between tests
type HandlerTestSuite struct {
	suite.Suite
	e                    *echo.Echo
	db                   *ent.Client
	galaDB               *ent.Client
	sharedIntegrationsRT *runtime.Runtime
	api                  *testclient.TestClient
	h                    *handlers.Handler
	router               *route.Router
	tf                   *testutils.TestFixture
	ofgaTF               *fgatest.OpenFGATestFixture
	stripeMockBackend    *mocks.MockStripeBackend
	objectStore          *objects.Service
	sharedTokenManager   *tokens.TokenManager
	sharedRedisClient    *redis.Client
	sharedSessionManager sessions.Store[map[string]any]
	sharedFGAClient      *fgax.Client
	sharedOTPManager     *totp.Client
	sharedPool           *gala.Pool
	galaRuntime          *gala.Gala
	sharedSlackRecorder  *slackWebhookRecorder
	registeredRoutes     map[string]struct{}
	sharedAuthMiddleware echo.MiddlewareFunc
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

	suite.tf = entdb.NewTestFixture()

	version, err := fgaversion.GetVersion()
	require.NoError(suite.T(), err)

	suite.ofgaTF = fgatest.NewFGATestcontainer(context.Background(),
		fgatest.WithModuleFile(fgaModuleFile),
		fgatest.WithEnvVars(coreutils.GetDefaultFGAEnvs()),
		fgatest.WithVersion(version),
	)

	suite.sharedTokenManager, err = coreutils.CreateTokenManager(-15 * time.Minute) //nolint:mnd
	require.NoError(suite.T(), err)

	suite.sharedRedisClient = coreutils.NewRedisClient()

	suite.sharedSessionManager = coreutils.CreateSessionManager()

	suite.sharedFGAClient, err = suite.ofgaTF.NewFgaClient(context.Background())
	require.NoError(suite.T(), err)

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

	suite.sharedPool = gala.NewPool(
		gala.WithWorkers(100), //nolint:mnd
		gala.WithPoolName("ent_client_pool"),
	)

	galaInstance, err := gala.NewGala(context.Background(), gala.Config{
		DispatchMode:      gala.DispatchModeDurable,
		ConnectionURI:     suite.tf.URI,
		QueueName:         "handler_integration_test",
		WorkerCount:       5, //nolint:mnd
		RunMigrations:     true,
		FetchCooldown:     time.Millisecond,
		FetchPollInterval: 10 * time.Millisecond, //nolint:mnd
	})
	require.NoError(suite.T(), err)

	require.NoError(suite.T(), galaInstance.StartWorkers(context.Background()))

	suite.galaRuntime = galaInstance

	hc, err := entdb.NewTestHistoryClient(context.Background(), suite.tf)
	require.NoError(suite.T(), err)

	galaSessionConfig := sessions.NewSessionConfig(
		suite.sharedSessionManager,
		sessions.WithPersistence(suite.sharedRedisClient),
	)

	galaSessionConfig.CookieConfig = sessions.DebugOnlyCookieConfig

	galaMockBackend := new(mocks.MockStripeBackend)
	galaMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	galaMockBackend.On("CallRaw", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	galaEntitlements, err := entitlements.NewStripeClient(
		entitlements.WithAPIKey("not_a_stripe_key"),
		entitlements.WithConfig(entitlements.Config{Enabled: true}),
		entitlements.WithBackends(&stripe.Backends{
			API:     galaMockBackend,
			Connect: galaMockBackend,
			Uploads: galaMockBackend,
		}),
	)
	require.NoError(suite.T(), err)

	galaOpts := []ent.Option{
		ent.Authz(*suite.sharedFGAClient),
		ent.TokenManager(suite.sharedTokenManager),
		ent.SessionConfig(&galaSessionConfig),
		ent.EntConfig(&entconfig.Config{Modules: entconfig.Modules{Enabled: true, UseSandbox: true}}),
		ent.TOTP(suite.sharedOTPManager),
		ent.Pool(suite.sharedPool),
		ent.EntitlementManager(galaEntitlements),
		ent.HistoryClient(hc),
	}

	galaJobOpts := []riverqueue.Option{riverqueue.WithConnectionURI(suite.tf.URI)}

	galaDB, err := entdb.NewTestClient(context.Background(), suite.tf, galaJobOpts, nil, galaOpts)
	require.NoError(suite.T(), err)

	suite.galaDB = galaDB

	hooks.SetTrustCenterConfig(hooks.TrustCenterConfig{
		CnameTarget:   cnameTargetTest,
		PreviewZoneID: previewMappableDomainZoneIDTest,
	})

	previewDomainCtx := privacy.DecisionContext(context.Background(), privacy.Allow)
	_, err = suite.galaDB.MappableDomain.Create().
		SetName(cnameTargetTest).
		SetZoneID(previewMappableDomainZoneIDTest).
		Save(previewDomainCtx)
	require.NoError(suite.T(), err)

	credStore, err := keystore.NewStore(suite.galaDB)
	require.NoError(suite.T(), err)

	suite.sharedSlackRecorder = newSlackWebhookRecorder(suite.T())

	rt, err := runtime.New(runtime.Config{
		DB:       suite.galaDB,
		Gala:     suite.galaRuntime,
		Keystore: credStore,
		DefinitionBuilders: []registry.Builder{
			emaildef.Builder(emaildef.MockRuntimeConfig(), false),
			slackdef.Builder(slackdef.Config{}, &slackdef.RuntimeSlackConfig{WebhookURL: suite.sharedSlackRecorder.URL()}, false),
			registry.Builder(buildTestOAuthDefinition),
			githubapp.Builder(defaultGitHubAppSpec()),
			configTestDefinitionBuilder(configTestProviderID, false),
			configTestDefinitionBuilder(configTestFailHealthProviderID, true),
			configTestDefinitionBuilder("def_01K0TESTOTH00000000000001", false),
			definitionscim.Builder(),
			webhookTestDefinitionBuilder(webhookTestDefinitionID),
			githubTestDefinitionBuilder(disconnectTestDefinitionID),
			operationTestDefinitionBuilder(operationTestDefinitionID, false),
			operationTestDefinitionBuilder(operationTestInlineDefinitionID, true),
			userInputOnlyTestDefinitionBuilder("def_01K0TESTUIONLY000000000001"),
		},
	})
	require.NoError(suite.T(), err)

	suite.sharedIntegrationsRT = rt

	require.NoError(suite.T(), suite.galaRuntime.Attach(
		gala.WithValue(suite.galaDB),
		gala.WithRestoredValue("ent_client", ent.NewContext),
	))
}

func (suite *HandlerTestSuite) SetupTest() {
	t := suite.T()

	suite.registeredRoutes = make(map[string]struct{})

	ctx := context.Background()

	sessionConfig := sessions.NewSessionConfig(
		suite.sharedSessionManager,
		sessions.WithPersistence(suite.sharedRedisClient),
	)

	sessionConfig.CookieConfig = sessions.DebugOnlyCookieConfig

	hc, err := entdb.NewTestHistoryClient(ctx, suite.tf)
	require.NoError(t, err)

	entitlements, err := suite.mockStripeClient()
	require.NoError(t, err)

	opts := []ent.Option{
		ent.Authz(*suite.sharedFGAClient),
		ent.TokenManager(suite.sharedTokenManager),
		ent.SessionConfig(&sessionConfig),
		ent.EntConfig(&entconfig.Config{
			QuestionnaireProductURL: "https://console.example.com",
			Modules: entconfig.Modules{
				Enabled:    true,
				UseSandbox: true,
			},
		}),
		ent.TOTP(suite.sharedOTPManager),
		ent.Pool(suite.sharedPool),
		ent.EntitlementManager(entitlements),
		ent.HistoryClient(hc),
	}

	jobOpts := []riverqueue.Option{riverqueue.WithConnectionURI(suite.tf.URI)}

	db, err := entdb.NewTestClient(ctx, suite.tf, jobOpts, nil, opts)
	require.NoError(t, err, "failed opening connection to database")

	suite.objectStore, _, err = coreutils.MockStorageServiceWithValidationAndProvider(t, nil, nil)
	require.NoError(t, err)

	err = db.Job.TruncateRiverTables(ctx)
	require.NoError(t, err)

	suite.mockEmailSender().Reset()

	runtime.SetDefault(suite.sharedIntegrationsRT)
	suite.db = db

	suite.api, err = coreutils.TestClient(suite.db, suite.objectStore)
	require.NoError(t, err)

	suite.router = setupRouter()

	suite.h = handlerSetup(suite.db)
	suite.configureIntegrationOAuthRuntime()
	if suite.h.Entitlements.Config.StripeWebhookSecrets == nil {
		suite.h.Entitlements.Config.StripeWebhookSecrets = map[string]string{}
	}
	suite.h.Entitlements.Config.StripeWebhookSecrets[stripe.APIVersion] = webhookSecret

	suite.h.OTPManager = suite.sharedOTPManager

	suite.e = suite.router.Echo

	transactionConfig := transaction.Client{
		EntDBClient: suite.db,
	}
	suite.e.Use(transactionConfig.Middleware)

	suite.sharedAuthMiddleware = suite.createAuthMiddleware()

	suite.setupTestData(ctx)
}

// registerAuthenticatedTestHandler registers a handler with authentication middleware for testing authenticated endpoints
func (suite *HandlerTestSuite) registerAuthenticatedTestHandler(method, path string, handlerFunc func(echo.Context) error) {
	suite.e.Add(method, path, handlerFunc, suite.sharedAuthMiddleware)
}

// createAuthMiddleware creates authentication middleware for tests
func (suite *HandlerTestSuite) createAuthMiddleware() echo.MiddlewareFunc {
	keys, err := suite.db.TokenManager.Keys()
	require.NoError(suite.T(), err)

	validator := tokens.NewJWKSValidator(keys, "http://localhost:17608", "http://localhost:17608")

	opts := []authmiddleware.Option{
		authmiddleware.WithDBClient(suite.db),
		authmiddleware.WithAllowAnonymous(true),
		authmiddleware.WithValidator(validator),
	}

	conf := authmiddleware.NewAuthOptions(opts...)

	return authmiddleware.Authenticate(&conf)
}

// registerTestHandler is a helper to register test handlers
func (suite *HandlerTestSuite) registerTestHandler(method, path string, handlerFunc func(echo.Context) error) {
	suite.e.Add(method, path, handlerFunc)
}

func (suite *HandlerTestSuite) registerRouteOnce(method, path string, handlerFunc func(echo.Context) error) {
	key := method + " " + path
	if _, exists := suite.registeredRoutes[key]; exists {
		return
	}
	suite.registeredRoutes[key] = struct{}{}
	suite.registerTestHandler(method, path, handlerFunc)
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

	suite.mockEmailSender().Reset()
}

func (suite *HandlerTestSuite) TearDownSuite() {
	if suite.galaRuntime != nil {
		_ = suite.galaRuntime.StopWorkers(context.Background())
		_ = suite.galaRuntime.Close()
	}

	if suite.galaDB != nil {
		_ = suite.galaDB.CloseAll()
	}

	suite.sharedSlackRecorder.Close()

	testutils.TeardownFixture(suite.tf)

	err := suite.ofgaTF.TeardownFixture()
	require.NoError(suite.T(), err)
}

// WaitForEvents blocks until runnable and in-flight Gala jobs complete
func (suite *HandlerTestSuite) WaitForEvents() {
	require.NoError(suite.T(), suite.galaRuntime.WaitIdle(suite.T().Context()))
}

func (suite *HandlerTestSuite) waitForGala(runtime *gala.Gala) {
	suite.Require().NoError(runtime.WaitIdle(suite.T().Context()))
}

func setupRouter() *route.Router {
	return server.NewRouter(server.LogConfig{
		PrettyLog: true,
		LogLevel:  1,
	})
}

// handlerSetup is used for required references in the handler tests
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
		ConsoleURL:               "http://console.example",
		DefaultTrustCenterDomain: "trust.openlane.com",
		IntegrationsConfig: catalog.Config{
			ConsoleIntegrationPath: "/organization-settings/integrations",
		},
	}

	return h
}

// testAuthDefinitionID is the canonical ID for the test OAuth definition
const testAuthDefinitionID = "def_01TEST0AUTH0000000000000001"

// testOAuthCredential is the credential type the test OAuth flow stores
type testOAuthCredential struct {
	// AccessToken is the issued access token
	AccessToken string `json:"access_token"`
	// RefreshToken is the issued refresh token
	RefreshToken string `json:"refresh_token"`
}

var testAuthCredentialRef = types.CredentialRefOf[testOAuthCredential]()

// configureIntegrationOAuthRuntime sets up the integrations runtime with a test OAuth definition
func (suite *HandlerTestSuite) configureIntegrationOAuthRuntime() {
	suite.h.IntegrationsRuntime = suite.sharedIntegrationsRT
}

// mockEmailSender returns the mock email sender from the shared integration runtime
func (suite *HandlerTestSuite) mockEmailSender() *mockprovider.EmailSender {
	rc, ok := suite.sharedIntegrationsRT.Registry().RuntimeClient(emaildef.DefinitionID.ID())
	suite.Require().True(ok, "email runtime client not found")

	ms := emaildef.MockSenderFromClient(rc)
	suite.Require().NotNil(ms, "mock sender not found")

	return ms
}

func buildTestOAuthDefinition() (types.Definition, error) {
	return types.Definition{
		DefinitionSpec: types.DefinitionSpec{
			ID:          testAuthDefinitionID,
			DisplayName: "Test OAuth",
			Active:      true,
		},
		CredentialRegistrations: []types.CredentialRegistration{
			testAuthCredentialRef.Registration(types.CredentialRegistration{
				Name:        "Test OAuth Credential",
				Description: "Auth-managed credential slot used by the test OAuth definition.",
			}),
		},
		HealthCheck: types.CredentialHealthCheck(func(context.Context, types.OperationRequest) (json.RawMessage, error) {
			return json.RawMessage(`{"ok":true}`), nil
		}),
		Connections: []types.ConnectionRegistration{
			{
				CredentialRef:  testAuthCredentialRef.ID(),
				Name:           "Test OAuth",
				Description:    "Authenticate the test definition using the OAuth callback fixture.",
				CredentialRefs: []types.CredentialSlotID{testAuthCredentialRef.ID()},
				Auth: &types.AuthRegistration{
					CredentialRef: testAuthCredentialRef.ID(),
					Start:         testAuthStart,
					Complete:      testAuthComplete,
				},
				Disconnect: &types.DisconnectRegistration{
					CredentialRef: testAuthCredentialRef.ID(),
					Description:   "Remove the persisted test OAuth credential and disconnect this installation.",
				},
			},
		},
	}, nil
}

type testCallbackPayload struct {
	State string `json:"state"`
	Code  string `json:"code,omitempty"`
}

func testAuthStart(_ context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
	oauthState := ulids.New().String()
	stateBytes, _ := json.Marshal(map[string]string{"state": oauthState})

	return types.AuthStartResult{
		URL:   fmt.Sprintf("https://example.com/oauth/authorize?state=%s", oauthState),
		State: stateBytes,
	}, nil
}

func testAuthComplete(_ context.Context, callbackState json.RawMessage, input types.AuthCallbackInput) (types.AuthCompleteResult, error) {
	var cs, inp testCallbackPayload
	json.Unmarshal(callbackState, &cs) //nolint:errcheck
	inp = testCallbackPayload{
		State: input.First("state"),
		Code:  input.First("code"),
	}

	if inp.Code == "" {
		return types.AuthCompleteResult{}, errors.New("missing oauth code")
	}

	if cs.State != inp.State {
		return types.AuthCompleteResult{}, errors.New("oauth state mismatch")
	}

	tokenData, _ := json.Marshal(map[string]string{
		"access_token":  "test-access-token",
		"refresh_token": "test-refresh-token",
	})

	return types.AuthCompleteResult{
		Credential: types.CredentialSet{
			Data: tokenData,
		},
	}, nil
}

// mockStripeClient creates a new stripe client with mock backend
func (suite *HandlerTestSuite) mockStripeClient() (*entitlements.StripeClient, error) {
	suite.stripeMockBackend = new(mocks.MockStripeBackend)
	stripeTestBackends := &stripe.Backends{
		API:     suite.stripeMockBackend,
		Connect: suite.stripeMockBackend,
		Uploads: suite.stripeMockBackend,
	}

	suite.orgSubscriptionMocks()

	return entitlements.NewStripeClient(entitlements.WithAPIKey("not_a_stripe_key"),
		entitlements.WithConfig(entitlements.Config{
			Enabled:             true,
			StripeWebhookSecret: webhookSecret,
		},
		),
		entitlements.WithBackends(stripeTestBackends),
	)
}

// mockCustomer is the stripe customer fixture used in webhook tests
var mockCustomer = &stripe.Customer{
	ID: "cus_test_customer",
	Subscriptions: &stripe.SubscriptionList{
		Data: []*stripe.Subscription{
			{
				Customer: &stripe.Customer{
					ID: "cus_test_customer",
				},
				ID: seedStripeSubscriptionID,
				Items: &stripe.SubscriptionItemList{
					Data: []*stripe.SubscriptionItem{
						{
							Price: &stripe.Price{
								UnitAmount: 1000,
								ID:         "price_test_price",
								Currency:   "usd",
								Recurring: &stripe.PriceRecurring{
									Interval: "month",
								},
							},
						},
					},
				},
			},
		},
	},
}

var mockSubscription = &stripe.Subscription{
	ID: "sub_test_subscription",
	Items: &stripe.SubscriptionItemList{
		Data: []*stripe.SubscriptionItem{
			{
				Price: &stripe.Price{
					Product: &stripe.Product{
						ID: "prod_test_product",
					},
					ID: "price_test_price",
					Recurring: &stripe.PriceRecurring{
						Interval: "month",
					},
					Currency: "usd",
				},
			},
		},
	},
	Metadata: map[string]string{
		"organization_id": ulids.New().String(),
	},
}

var mockProduct = &stripe.Product{
	ID:   "prod_test_product",
	Name: "Test Product",
}

// orgSubscriptionMocks mocks the stripe calls for org subscription during the webhook tests
func (suite *HandlerTestSuite) orgSubscriptionMocks() {
	suite.stripeMockBackend.On("CallRaw", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.Params"), mock.AnythingOfType("*stripe.v1SearchPage[*github.com/stripe/stripe-go/v86.Customer]")).Run(func(args mock.Arguments) {
		out := args.Get(4)

		payload := map[string]any{
			"object":   "search_result",
			"data":     []*stripe.Customer{mockCustomer},
			"has_more": false,
		}

		b, _ := json.Marshal(payload)
		_ = json.Unmarshal(b, out)
	}).Return(nil)

	suite.stripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.CustomerRetrieveParams"), mock.AnythingOfType("*stripe.Customer")).Run(func(args mock.Arguments) {
		mockCustomerSearchResult := args.Get(4).(*stripe.Customer)

		*mockCustomerSearchResult = *mockCustomer

	}).Return(nil)

	suite.stripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.CustomerCreateParams"), mock.AnythingOfType("*stripe.Customer")).Run(func(args mock.Arguments) {
		mockCustomerSearchResult := args.Get(4).(*stripe.Customer)

		*mockCustomerSearchResult = *mockCustomer

	}).Return(nil)

	suite.stripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.SubscriptionCreateParams"), mock.AnythingOfType("*stripe.Subscription")).Run(func(args mock.Arguments) {
		mockSubscriptionSearchResult := args.Get(4).(*stripe.Subscription)

		*mockSubscriptionSearchResult = *mockSubscription

	}).Return(nil)

	suite.stripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.ProductRetrieveParams"), mock.AnythingOfType("*stripe.Product")).Run(func(args mock.Arguments) {
		mockProductRetrieveResult := args.Get(4).(*stripe.Product)

		*mockProductRetrieveResult = *mockProduct

	}).Return(nil)

	suite.stripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.SubscriptionRetrieveParams"), mock.AnythingOfType("*stripe.Product")).Run(func(args mock.Arguments) {
		mockSubscriptionRetrieveResult := args.Get(4).(*stripe.Subscription)

		*mockSubscriptionRetrieveResult = *mockSubscription

	}).Return(nil)

	suite.stripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.SubscriptionScheduleCreateParams"), mock.AnythingOfType("*stripe.SubscriptionSchedule")).Run(func(args mock.Arguments) {
		mockSubscriptionScheduleResult := args.Get(4).(*stripe.SubscriptionSchedule)

		*mockSubscriptionScheduleResult = stripe.SubscriptionSchedule{
			ID: "sub_sched_test_schedule",
			Phases: []*stripe.SubscriptionSchedulePhase{
				{
					Items: []*stripe.SubscriptionSchedulePhaseItem{
						{
							Price:    mockProduct.DefaultPrice,
							Quantity: 1,
						},
					},
				},
			},
			Object: "subscription_schedule",
		}

	}).Return(nil)

	suite.stripeMockBackend.On("CallRaw", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.Params"), mock.AnythingOfType("*stripe.EntitlementsActiveEntitlementList")).Run(func(args mock.Arguments) {
		mockCustomerSearchResult := args.Get(4).(*stripe.EntitlementsActiveEntitlementList)

		*mockCustomerSearchResult = stripe.EntitlementsActiveEntitlementList{
			Data: []*stripe.EntitlementsActiveEntitlement{
				{
					Feature: &stripe.EntitlementsFeature{
						ID:        "feat_test_feature",
						LookupKey: "test_feature",
					},
				},
			},
		}

	}).Return(nil)
}
