package handlers_test

import (
	"context"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/stripe/stripe-go/v82"

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
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/internal/ent/entconfig"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/entdb"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/internal/httpserve/route"
	"github.com/theopenlane/core/internal/httpserve/server"
	objmw "github.com/theopenlane/core/internal/middleware/objects"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/entitlements/mocks"
	"github.com/theopenlane/core/pkg/events/soiree"
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
	fgaModelFile             = "../../../fga/model/model.fga"
	seedStripeSubscriptionID = "sub_test_subscription"
)

// HandlerTestSuite handles the setup and teardown between tests
type HandlerTestSuite struct {
	suite.Suite
	e                    *echo.Echo
	db                   *ent.Client
	api                  *testclient.TestClient
	h                    *handlers.Handler
	router               *route.Router
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
	suite.ofgaTF = fgatest.NewFGATestcontainer(context.Background(),
		fgatest.WithModelFile(fgaModelFile),
		fgatest.WithEnvVars(map[string]string{
			"OPENFGA_MAX_CHECKS_PER_BATCH_CHECK":          "100",
			"OPENFGA_CHECK_ITERATOR_CACHE_ENABLED":        "false",
			"OPENFGA_LIST_OBJECTS_ITERATOR_CACHE_ENABLED": "false",
		}),
	)

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

	// setup mock entitlements client
	entitlements, err := suite.mockStripeClient()
	require.NoError(t, err)

	opts := []ent.Option{
		ent.Authz(*suite.sharedFGAClient),
		ent.Emailer(&emailtemplates.Config{
			CompanyName: "Meow Inc.",
		}),
		ent.TokenManager(suite.sharedTokenManager),
		ent.SessionConfig(&sessionConfig),
		ent.EntConfig(&entconfig.Config{
			Modules: entconfig.Modules{
				Enabled:    true,
				UseSandbox: true,
			},
		}),
		ent.TOTP(suite.sharedOTPManager),
		ent.PondPool(suite.sharedPondPool),
		ent.EntitlementManager(entitlements),
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

// mockStripeClient creates a new stripe client with mock backend
func (suite *HandlerTestSuite) mockStripeClient() (*entitlements.StripeClient, error) {
	suite.stripeMockBackend = new(mocks.MockStripeBackend)
	stripeTestBackends := &stripe.Backends{
		API:     suite.stripeMockBackend,
		Connect: suite.stripeMockBackend,
		Uploads: suite.stripeMockBackend,
	}

	suite.orgSubscriptionMocks()

	return entitlements.NewStripeClient(entitlements.WithAPIKey("sk_test_testing"),
		entitlements.WithConfig(entitlements.Config{
			Enabled:             true,
			StripeWebhookSecret: webhookSecret,
		},
		),
		entitlements.WithBackends(stripeTestBackends),
	)
}

// mockCustomer for webhook tests
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
	// setup mocks for search
	suite.stripeMockBackend.On("CallRaw", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.Params"), mock.AnythingOfType("*stripe.CustomerSearchResult")).Run(func(args mock.Arguments) {
		mockCustomerSearchResult := args.Get(4).(*stripe.CustomerSearchResult)

		data := []*stripe.Customer{}
		data = append(data, mockCustomer)
		*mockCustomerSearchResult = stripe.CustomerSearchResult{
			Data: data,
		}

	}).Return(nil)

	// setup mocks for get customer by id
	suite.stripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.CustomerRetrieveParams"), mock.AnythingOfType("*stripe.Customer")).Run(func(args mock.Arguments) {
		mockCustomerSearchResult := args.Get(4).(*stripe.Customer)

		*mockCustomerSearchResult = *mockCustomer

	}).Return(nil)

	// setup mocks for creating customer params
	suite.stripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.CustomerCreateParams"), mock.AnythingOfType("*stripe.Customer")).Run(func(args mock.Arguments) {
		mockCustomerSearchResult := args.Get(4).(*stripe.Customer)

		*mockCustomerSearchResult = *mockCustomer

	}).Return(nil)

	// mock for subscription create params
	suite.stripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.SubscriptionCreateParams"), mock.AnythingOfType("*stripe.Subscription")).Run(func(args mock.Arguments) {
		mockSubscriptionSearchResult := args.Get(4).(*stripe.Subscription)

		*mockSubscriptionSearchResult = *mockSubscription

	}).Return(nil)

	// mock for product retrieve params
	suite.stripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.ProductRetrieveParams"), mock.AnythingOfType("*stripe.Product")).Run(func(args mock.Arguments) {
		mockProductRetrieveResult := args.Get(4).(*stripe.Product)

		*mockProductRetrieveResult = *mockProduct

	}).Return(nil)

	// mock for product params
	suite.stripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.SubscriptionRetrieveParams"), mock.AnythingOfType("*stripe.Product")).Run(func(args mock.Arguments) {
		mockSubscriptionRetrieveResult := args.Get(4).(*stripe.Subscription)

		*mockSubscriptionRetrieveResult = *mockSubscription

	}).Return(nil)

	// setup mocks for org subscription schedule
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

	// setup mocks for getting entitlements
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
