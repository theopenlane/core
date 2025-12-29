package graphapi_test

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go/v84"
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
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/internal/entdb"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/internal/objects/validators"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/entitlements/mocks"
	"github.com/theopenlane/core/pkg/events/soiree"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
	mock_shared "github.com/theopenlane/core/pkg/objects/mocks"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/core/pkg/summarizer"
	coreutils "github.com/theopenlane/core/pkg/testutils"

	// import generated runtime which is required to prevent cyclical dependencies
	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
	_ "github.com/theopenlane/core/internal/ent/historygenerated/runtime"
)

const (
	fgaModelFile = "../../fga/model/model.fga"

	redacted = "*****************************"

	// common error message strings
	notFoundErrorMsg         = "not found"
	notAuthorizedErrorMsg    = "you are not authorized to perform this action"
	invalidInputErrorMsg     = "invalid input"
	seedStripeSubscriptionID = "sub_test_subscription"
	webhookSecret            = "whsec_test_secret"
)

// GraphTestSuite handles the setup and teardown between tests
type GraphTestSuite struct {
	client            *client
	tf                *testutils.TestFixture
	ofgaTF            *fgatest.OpenFGATestFixture
	stripeMockBackend *mocks.MockStripeBackend
}

// client contains all the clients the test need to interact with
type client struct {
	db           *ent.Client
	api          *testclient.TestClient
	apiWithPAT   *testclient.TestClient
	apiWithToken *testclient.TestClient
	fga          *fgax.Client
	objectStore  *objects.Service
	mockProvider *mock_shared.MockProvider
}

var suite = &GraphTestSuite{}

func TestMain(m *testing.M) {
	flag.Parse()

	// Create a new testing.T instance
	// Note: this is only to seed data; you should not use this instance for actual tests
	// this also cannot be used with a t.FailNow(), you must os.Exit when using this t
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
	suite.ofgaTF = fgatest.NewFGATestcontainer(context.Background(),
		fgatest.WithModelFile(fgaModelFile),
		fgatest.WithEnvVars(map[string]string{
			"OPENFGA_MAX_CHECKS_PER_BATCH_CHECK":          "100",
			"OPENFGA_CHECK_ITERATOR_CACHE_ENABLED":        "false",
			"OPENFGA_LIST_OBJECTS_ITERATOR_CACHE_ENABLED": "false",
		},
		))

	ctx := context.Background()

	// setup fga client
	fgaClient, err := suite.ofgaTF.NewFgaClient(ctx)
	requireNoError(t, err)

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

	tm, err := coreutils.CreateTokenManager(-15 * time.Minute) //nolint:mnd
	requireNoError(t, err)

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
	requireNoError(t, err)

	pool := soiree.NewPondPool(
		soiree.WithMaxWorkers(100), //nolint:mnd
		soiree.WithName("ent_client_pool"),
	)

	// setup history client
	hc, err := entdb.NewTestHistoryClient(ctx, suite.tf)
	requireNoError(t, err)

	// setup mock entitlements client
	entitlements, err := suite.mockStripeClient()
	requireNoError(t, err)

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
		ent.EntitlementManager(entitlements),
		ent.EmailVerifier(ev),
		ent.HistoryClient(hc),
	}

	// create database connection
	jobOpts := []riverqueue.Option{riverqueue.WithConnectionURI(suite.tf.URI)}

	db, err := entdb.NewTestClient(ctx, suite.tf, jobOpts, opts)
	requireNoError(t, err)

	c.objectStore, c.mockProvider, err = coreutils.MockStorageServiceWithValidationAndProvider(t, nil, validators.MimeTypeValidator)
	requireNoError(t, err)

	// assign values
	c.db = db
	c.api, err = coreutils.TestClient(c.db, c.objectStore)
	requireNoError(t, err)

	suite.client = c
}

func (suite *GraphTestSuite) TearDownSuite(t *testing.T) {
	// close the database connection
	err := suite.client.db.Close()
	requireNoError(t, err)

	// close the database container
	testutils.TeardownFixture(suite.tf)

	// terminate all fga containers
	err = suite.ofgaTF.TeardownFixture()
	requireNoError(t, err)
}

// expectUpload sets up the mock object store to expect an upload and related operations
func expectUpload(t *testing.T, mockProvider *mock_shared.MockProvider, expectedUploads []graphql.Upload) {
	assert.Assert(t, mockProvider != nil)

	mockScheme := "file://"

	for _, upload := range expectedUploads {
		mockProvider.On("GetScheme").Return(&mockScheme).Once()
		mockProvider.On("ProviderType").Return(storage.DiskProvider).Maybe()
		mockProvider.On("Upload", mock.Anything, mock.Anything, mock.Anything).Return(&storage.UploadedMetadata{
			FileMetadata: pkgobjects.FileMetadata{
				Key:          "test-key",
				Size:         upload.Size,
				Folder:       "test-folder",
				Bucket:       "test-bucket",
				ContentType:  upload.ContentType,
				ProviderType: storage.DiskProvider,
				FullURI:      "file:///tmp/test-file",
			},
		}, nil).Once()

		// Allow document hooks to download the just-uploaded content for parsing
		mockProvider.On("Download", mock.Anything, mock.Anything, mock.Anything).Return(&storage.DownloadedMetadata{
			File: []byte("test content"),
			Size: upload.Size,
		}, nil).Maybe()
	}
}

// expectDelete sets up the mock object store to expect a delete and related operations
func expectDelete(t *testing.T, mockProvider *mock_shared.MockProvider, expectedUploads []graphql.Upload) {
	assert.Assert(t, mockProvider != nil)

	mockScheme := "file://"

	for range expectedUploads {
		mockProvider.On("GetScheme").Return(&mockScheme).Once()
		mockProvider.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	}
}

// expectUploadNillable sets up the mock object store to expect an upload and related operations
func expectUploadNillable(t *testing.T, mockProvider *mock_shared.MockProvider, expectedUploads []*graphql.Upload) {
	assert.Check(t, mockProvider != nil)

	mockScheme := "file://"

	for _, upload := range expectedUploads {
		if upload != nil {
			mockProvider.On("GetScheme").Return(&mockScheme).Once()
			mockProvider.On("ProviderType").Return(storage.DiskProvider).Maybe()
			mockProvider.On("Upload", mock.Anything, mock.Anything, mock.Anything).Return(&storage.UploadedMetadata{
				FileMetadata: pkgobjects.FileMetadata{
					Key:          "test-key",
					Size:         upload.Size,
					Folder:       "test-folder",
					Bucket:       "test-bucket",
					ContentType:  upload.ContentType,
					ProviderType: storage.DiskProvider,
					FullURI:      "file:///tmp/test-file",
				},
			}, nil).Once()

			// Allow document hooks to download the just-uploaded content for parsing
			mockProvider.On("Download", mock.Anything, mock.Anything, mock.Anything).Return(&storage.DownloadedMetadata{
				File: []byte("test content"),
				Size: upload.Size,
			}, nil).Maybe()
		}
	}
}

// expectUploadCheckOnly sets up the mock object store to expect an upload check only operation
// but fails before the upload is attempted
func expectUploadCheckOnly(t *testing.T, mockProvider *mock_shared.MockProvider) {
	assert.Assert(t, mockProvider != nil)

	mockScheme := "file://"

	mockProvider.On("GetScheme").Return(&mockScheme).Once()
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

	assert.Equal(t, code, testclient.GetErrorCode(err))
}

// assertErrorMessage checks if the error message matches the expected message
func assertErrorMessage(t *testing.T, err *gqlerror.Error, msg string) {
	t.Helper()

	assert.Equal(t, msg, testclient.GetErrorMessage(err))
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		log.Error().Err(err).Msg("fatal error during test setup or teardown")

		os.Exit(1)
	}
}

// mockStripeClient creates a new stripe client with mock backend
func (suite *GraphTestSuite) mockStripeClient() (*entitlements.StripeClient, error) {
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

var mockItems = []*stripe.SubscriptionItem{
	{
		Price: &stripe.Price{
			Product: &stripe.Product{
				ID: "prod_test_product",
			},
			ID: "price_test_price",
			Recurring: &stripe.PriceRecurring{
				Interval: "month",
			},
			Currency:   "usd",
			UnitAmount: 1000,
		},
	},
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
					Data: mockItems,
				},
			},
		},
	},
}

var mockSubscription = &stripe.Subscription{
	ID:     "sub_test_subscription",
	Status: "active",
	Items: &stripe.SubscriptionItemList{
		Data: mockItems,
	},
	Metadata: map[string]string{
		"organization_id": ulids.New().String(),
	},
	Customer: &stripe.Customer{
		ID: "cus_test_customer",
	},
	TrialEnd:     time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 days from now
	DaysUntilDue: 15,
}

var mockProduct = &stripe.Product{
	ID:   "prod_test_product",
	Name: "Test Product",
}

// orgSubscriptionMocks mocks the stripe calls for org subscription during the webhook tests
func (suite *GraphTestSuite) orgSubscriptionMocks() {
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

	// mock customer search
	suite.stripeMockBackend.On("CallRaw", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.Params"), mock.AnythingOfType("*stripe.v1SearchPage[*github.com/stripe/stripe-go/v84.Customer]")).Run(func(args mock.Arguments) {
		out := args.Get(4) // this is *v1SearchPage[*stripe.Customer] now, but unexported

		// Build a payload that matches Stripe search response shape
		payload := map[string]any{
			"object":   "search_result",
			"data":     []*stripe.Customer{mockCustomer},
			"has_more": false,
		}

		b, _ := json.Marshal(payload)
		_ = json.Unmarshal(b, out)
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

	// setup mocks for getting entitlements
	suite.stripeMockBackend.On("CallRaw", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.Params"), mock.AnythingOfType("*stripe.EntitlementsActiveEntitlementList")).Run(func(args mock.Arguments) {
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

	// setup mocks for subscription schedule creation
	suite.stripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.SubscriptionScheduleCreateParams"), mock.AnythingOfType("*stripe.SubscriptionSchedule")).Run(func(args mock.Arguments) {
		mockSubscriptionScheduleResult := args.Get(4).(*stripe.SubscriptionSchedule)

		*mockSubscriptionScheduleResult = stripe.SubscriptionSchedule{
			ID:     "sched_test_schedule",
			Status: "active",
		}

	}).Return(nil)

	// setup mocks for customer update params
	suite.stripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.CustomerUpdateParams"), mock.AnythingOfType("*stripe.Customer")).Run(func(args mock.Arguments) {
		mockCustomerUpdateResult := args.Get(4).(*stripe.Customer)

		*mockCustomerUpdateResult = *mockCustomer

	}).Return(nil)
}
