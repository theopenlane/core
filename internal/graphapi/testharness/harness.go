//go:build test

package testharness

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gqlgo/gqlgenc/clientv2"
	"github.com/mcuadros/go-defaults"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v86"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"gotest.tools/v3/assert"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/fgax"
	fgatest "github.com/theopenlane/iam/fgax/testutils"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/riverboat/pkg/riverqueue"

	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/totp"
	"github.com/theopenlane/utils/testutils"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
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
	"github.com/theopenlane/core/v2/pkg/entitlements"
	"github.com/theopenlane/core/v2/pkg/entitlements/mocks"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
	authmw "github.com/theopenlane/core/v2/pkg/middleware/auth"
	pkgobjects "github.com/theopenlane/core/v2/pkg/objects"
	mock_shared "github.com/theopenlane/core/v2/pkg/objects/mocks"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/core/v2/pkg/summarizer"
	mockprovider "github.com/theopenlane/newman/providers/mock"

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
	Client               *Client
	TF                   *testutils.TestFixture
	OFGATF               *fgatest.OpenFGATestFixture
	StripeMockBackend    *mocks.MockStripeBackend
	CacheRefreshServer   *httptest.Server
	GalaRuntime          *gala.Gala
	IntegrationsRT       *intruntime.Runtime
	WorkflowEngine       *engine.WorkflowEngine
	WorkflowListenersMu  sync.Mutex
	WorkflowListenerRefs int
	WorkflowListenerIDs  []gala.ListenerID
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

// DeletedKeys is a concurrency safe set of the storage keys removed from object storage, the gala
// workers deleting them run on their own goroutines while the test asserts on the main one
type DeletedKeys struct {
	mu   sync.Mutex
	keys map[string]struct{}
}

// NewDeletedKeys returns an empty set
func NewDeletedKeys() *DeletedKeys {
	return &DeletedKeys{keys: map[string]struct{}{}}
}

// Add records a storage key that was deleted
func (d *DeletedKeys) Add(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.keys[key] = struct{}{}
}

// Has reports whether the given storage key was deleted
func (d *DeletedKeys) Has(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, ok := d.keys[key]

	return ok
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
			systemdef.Builder(systemdef.PaymentReminderConfig{}, systemdef.OrganizationDeleteConfig{}),
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

// AcquireWorkflowRuntime shares one workflow listener registration across parallel tests
func (suite *GraphTestSuite) AcquireWorkflowRuntime(t *testing.T) (*engine.WorkflowEngine, *gala.Gala) {
	t.Helper()

	suite.WorkflowListenersMu.Lock()
	if suite.WorkflowListenerRefs == 0 {
		listenerIDs, err := gala.Register(suite.GalaRuntime, hooks.WorkflowListeners()...)
		if err != nil {
			suite.WorkflowListenersMu.Unlock()
			RequireNoError(t, err)

			return nil, nil
		}

		suite.WorkflowListenerIDs = listenerIDs
	}

	suite.WorkflowListenerRefs++
	workflowEngine := suite.WorkflowEngine
	runtime := suite.GalaRuntime
	suite.WorkflowListenersMu.Unlock()

	t.Cleanup(func() {
		suite.ReleaseWorkflowRuntime(t)
	})

	return workflowEngine, runtime
}

// ReleaseWorkflowRuntime removes workflow listeners after the last borrowing test
func (suite *GraphTestSuite) ReleaseWorkflowRuntime(t *testing.T) {
	t.Helper()

	suite.WorkflowListenersMu.Lock()
	defer suite.WorkflowListenersMu.Unlock()

	if suite.WorkflowListenerRefs == 0 {
		return
	}

	suite.WorkflowListenerRefs--
	if suite.WorkflowListenerRefs > 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), gala.DefaultSoftStopTimeout)
	defer cancel()

	err := suite.GalaRuntime.RemoveListeners(ctx, suite.WorkflowListenerIDs...)
	RequireNoError(t, err)
	suite.WorkflowListenerIDs = nil
}

// WaitForEvents blocks until runnable and in-flight Gala jobs complete
func (suite *GraphTestSuite) WaitForEvents() {
	if err := suite.GalaRuntime.WaitIdle(context.Background()); err != nil {
		panic(err)
	}
}

func WaitForGala(t *testing.T, runtime *gala.Gala) {
	t.Helper()
	assert.NilError(t, runtime.WaitIdle(t.Context()))
}

// MockEmailSender extracts the mock email sender from the integration runtime
func (suite *GraphTestSuite) MockEmailSender() *mockprovider.EmailSender {
	rc, ok := suite.IntegrationsRT.Registry().RuntimeClient(emaildef.DefinitionID.ID())
	if !ok {
		panic("email runtime client not found")
	}

	ms := emaildef.MockSenderFromClient(rc)
	if ms == nil {
		panic("mock sender not found")
	}

	return ms
}

func (suite *GraphTestSuite) EnableGalaForTestSuite(t *testing.T) {
	t.Helper()

	if suite.GalaRuntime != nil {
		return
	}

	runtime, err := gala.NewGala(context.Background(), gala.Config{
		ConnectionURI:     suite.TF.URI,
		QueueName:         fmt.Sprintf("graphapi_test_%d", time.Now().UnixNano()),
		WorkerCount:       1,
		RunMigrations:     true,
		FetchCooldown:     time.Millisecond,
		FetchPollInterval: 10 * time.Millisecond,
	})
	require.NoError(t, err)

	require.NoError(t, runtime.Attach(
		gala.WithValue(runtime),
		gala.WithValue(suite.Client.DB),
		gala.WithValue(suite.Client.DB.EntitlementManager),
	))

	_, err = gala.Register(runtime, hooks.EntitlementListeners()...)
	require.NoError(t, err)

	err = runtime.StartWorkers(context.Background())
	require.NoError(t, err)

	suite.GalaRuntime = runtime

	t.Cleanup(func() {
		if suite.GalaRuntime == nil {
			return
		}

		err := suite.GalaRuntime.StopWorkers(context.Background())
		require.NoError(t, err)

		err = suite.GalaRuntime.Close()
		require.NoError(t, err)

		suite.GalaRuntime = nil
	})
}

// ExpectUpload sets up the mock object store to expect an upload and related operations
func ExpectUpload(t *testing.T, mockProvider *mock_shared.MockProvider, expectedUploads []graphql.Upload) {
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
			File: TestPDFBytes(),
			Size: upload.Size,
		}, nil).Maybe()
	}
}

// ExpectAttestedUpload sets up mock expectations for the attested PDF upload triggered by attestNDADocument
func ExpectAttestedUpload(t *testing.T, mockProvider *mock_shared.MockProvider) {
	assert.Assert(t, mockProvider != nil)

	mockScheme := "file://"

	mockProvider.On("GetScheme").Return(&mockScheme).Once()
	mockProvider.On("ProviderType").Return(storage.DiskProvider).Maybe()
	mockProvider.On("Upload", mock.Anything, mock.Anything, mock.Anything).Return(&storage.UploadedMetadata{
		FileMetadata: pkgobjects.FileMetadata{
			Key:          "test-key-attested",
			Folder:       "test-folder",
			Bucket:       "test-bucket",
			ContentType:  "application/pdf",
			ProviderType: storage.DiskProvider,
			FullURI:      "file:///tmp/test-file-attested",
		},
	}, nil).Once()
}

func ExpectUploadWithTemplateKind(t *testing.T, mockProvider *mock_shared.MockProvider, expectedUploads []graphql.Upload, kind enums.TemplateKind) {
	assert.Assert(t, mockProvider != nil)

	mockScheme := "file://"

	for _, upload := range expectedUploads {
		mockProvider.On("GetScheme").Return(&mockScheme).Once()
		mockProvider.On("ProviderType").Return(storage.DiskProvider).Maybe()
		uploadOpts := mock.MatchedBy(func(opts *storage.UploadOptions) bool {
			if opts == nil || opts.ProviderHints == nil || opts.ProviderHints.Metadata == nil {
				return false
			}

			return opts.ProviderHints.Metadata[objects.TemplateKindMetadataKey] == kind.String()
		})
		mockProvider.On("Upload", mock.Anything, mock.Anything, uploadOpts).Return(&storage.UploadedMetadata{
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
			File: TestPDFBytes(),
			Size: upload.Size,
		}, nil).Maybe()
	}
}

// ExpectDelete sets up the mock object store to expect a delete and related operations
func ExpectDelete(t *testing.T, mockProvider *mock_shared.MockProvider, expectedUploads []graphql.Upload) {
	assert.Assert(t, mockProvider != nil)

	mockScheme := "file://"

	for range expectedUploads {
		mockProvider.On("GetScheme").Return(&mockScheme).Once()
		mockProvider.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	}
}

// ExpectUploadNillable sets up the mock object store to expect an upload and related operations
func ExpectUploadNillable(t *testing.T, mockProvider *mock_shared.MockProvider, expectedUploads []*graphql.Upload) {
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

// ExpectUploadCheckOnly sets up the mock object store to expect an upload check only operation
// but fails before the upload is attempted
func ExpectUploadCheckOnly(t *testing.T, mockProvider *mock_shared.MockProvider) {
	assert.Assert(t, mockProvider != nil)

	mockScheme := "file://"

	mockProvider.On("GetScheme").Return(&mockScheme).Once()
}

// ParseClientError parses the error response from the client and returns a slice of gqlerror.Error
func ParseClientError(t *testing.T, err error) []*gqlerror.Error {
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

// AssertErrorCode checks if the error code matches the expected code
func AssertErrorCode(t *testing.T, err *gqlerror.Error, code string) {
	t.Helper()

	assert.Equal(t, code, testclient.GetErrorCode(err))
}

// AssertErrorMessage checks if the error message matches the expected message
func AssertErrorMessage(t *testing.T, err *gqlerror.Error, msg string) {
	t.Helper()

	assert.Equal(t, msg, testclient.GetErrorMessage(err))
}

func RequireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		log.Error().Err(err).Msg("fatal error during test setup or teardown")

		os.Exit(1)
	}
}

func FailNow(t *testing.T, msgs ...string) {
	t.Helper()
	logMsg := log.Error()

	for _, m := range msgs {
		logMsg.Str("msg", m)
	}

	logMsg.Msg("fatal error during test setup or teardown")

	os.Exit(1)
}

func TestPDFBytes() []byte {
	page := map[string]any{
		"paper":  "A4P",
		"origin": "UpperLeft",
		"fonts": map[string]any{
			"f": map[string]any{"name": "Helvetica", "size": 12},
		},
		"pages": map[string]any{
			"1": map[string]any{
				"content": map[string]any{
					"text": []map[string]any{
						{"value": "test", "pos": [2]float64{20, 20}, "font": map[string]any{"name": "$f"}},
					},
				},
			},
		},
	}

	jsonData, _ := json.Marshal(page)

	var buf bytes.Buffer
	_ = api.Create(nil, bytes.NewReader(jsonData), &buf, nil)

	return buf.Bytes()
}

// MockStripeClient creates a new stripe client with mock backend
func (suite *GraphTestSuite) MockStripeClient() (*entitlements.StripeClient, error) {
	suite.StripeMockBackend = new(mocks.MockStripeBackend)
	stripeTestBackends := &stripe.Backends{
		API:     suite.StripeMockBackend,
		Connect: suite.StripeMockBackend,
		Uploads: suite.StripeMockBackend,
	}

	suite.OrgSubscriptionMocks()

	return entitlements.NewStripeClient(entitlements.WithAPIKey("sk_test_testing"),
		entitlements.WithConfig(entitlements.Config{
			Enabled:             true,
			StripeWebhookSecret: webhookSecret,
		},
		),
		entitlements.WithBackends(stripeTestBackends),
	)
}

var MockItems = []*stripe.SubscriptionItem{
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

// MockCustomer for webhook tests
var MockCustomer = &stripe.Customer{
	ID: "cus_test_customer",
	Subscriptions: &stripe.SubscriptionList{
		Data: []*stripe.Subscription{
			{
				Customer: &stripe.Customer{
					ID: "cus_test_customer",
				},
				ID: seedStripeSubscriptionID,
				Items: &stripe.SubscriptionItemList{
					Data: MockItems,
				},
			},
		},
	},
}

var MockSubscription = &stripe.Subscription{
	ID:     "sub_test_subscription",
	Status: "active",
	Items: &stripe.SubscriptionItemList{
		Data: MockItems,
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

var MockProduct = &stripe.Product{
	ID:   "prod_test_product",
	Name: "Test Product",
}

// OrgSubscriptionMocks mocks the stripe calls for org subscription during the webhook tests
func (suite *GraphTestSuite) OrgSubscriptionMocks() {
	// setup mocks for get customer by id
	suite.StripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.CustomerRetrieveParams"), mock.AnythingOfType("*stripe.Customer")).Run(func(args mock.Arguments) {
		mockCustomerSearchResult := args.Get(4).(*stripe.Customer)

		*mockCustomerSearchResult = *MockCustomer

	}).Return(nil)

	// setup mocks for creating customer params
	suite.StripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.CustomerCreateParams"), mock.AnythingOfType("*stripe.Customer")).Run(func(args mock.Arguments) {
		mockCustomerSearchResult := args.Get(4).(*stripe.Customer)

		*mockCustomerSearchResult = *MockCustomer

	}).Return(nil)

	// mock customer search
	suite.StripeMockBackend.On("CallRaw", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.Params"), mock.AnythingOfType("*stripe.v1SearchPage[*github.com/stripe/stripe-go/v86.Customer]")).Run(func(args mock.Arguments) {
		out := args.Get(4) // this is *v1SearchPage[*stripe.Customer] now, but unexported

		// Build a payload that matches Stripe search response shape
		payload := map[string]any{
			"object":   "search_result",
			"data":     []*stripe.Customer{MockCustomer},
			"has_more": false,
		}

		b, _ := json.Marshal(payload)
		_ = json.Unmarshal(b, out)
	}).Return(nil)

	// mock for subscription create params
	suite.StripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.SubscriptionCreateParams"), mock.AnythingOfType("*stripe.Subscription")).Run(func(args mock.Arguments) {
		mockSubscriptionSearchResult := args.Get(4).(*stripe.Subscription)

		*mockSubscriptionSearchResult = *MockSubscription

	}).Return(nil)

	// mock for product retrieve params
	suite.StripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.ProductRetrieveParams"), mock.AnythingOfType("*stripe.Product")).Run(func(args mock.Arguments) {
		mockProductRetrieveResult := args.Get(4).(*stripe.Product)

		*mockProductRetrieveResult = *MockProduct

	}).Return(nil)

	// mock for product params
	suite.StripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.SubscriptionRetrieveParams"), mock.AnythingOfType("*stripe.Product")).Run(func(args mock.Arguments) {
		mockSubscriptionRetrieveResult := args.Get(4).(*stripe.Subscription)

		*mockSubscriptionRetrieveResult = *MockSubscription

	}).Return(nil)

	// setup mocks for getting entitlements
	suite.StripeMockBackend.On("CallRaw", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.Params"), mock.AnythingOfType("*stripe.EntitlementsActiveEntitlementList")).Run(func(args mock.Arguments) {
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
	suite.StripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.SubscriptionScheduleCreateParams"), mock.AnythingOfType("*stripe.SubscriptionSchedule")).Run(func(args mock.Arguments) {
		mockSubscriptionScheduleResult := args.Get(4).(*stripe.SubscriptionSchedule)

		*mockSubscriptionScheduleResult = stripe.SubscriptionSchedule{
			ID:     "sched_test_schedule",
			Status: "active",
		}

	}).Return(nil)

	// setup mocks for customer update params
	suite.StripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.CustomerUpdateParams"), mock.AnythingOfType("*stripe.Customer")).Run(func(args mock.Arguments) {
		mockCustomerUpdateResult := args.Get(4).(*stripe.Customer)

		*mockCustomerUpdateResult = *MockCustomer

	}).Return(nil)
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
