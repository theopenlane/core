//go:build test

package engine_test

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"sort"
	"testing"
	"time"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go/v87"
	"github.com/theopenlane/iam/auth"
	fgatest "github.com/theopenlane/iam/fgax/testutils"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/riverboat/pkg/riverqueue"
	dbtestutils "github.com/theopenlane/utils/testutils"
	"github.com/theopenlane/utils/ulids"

	mockprovider "github.com/theopenlane/newman/providers/mock"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/fga/fgaversion"
	"github.com/theopenlane/core/v2/internal/ent/entconfig"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowassignmenttarget"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowdefinition"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowevent"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowobjectref"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowproposal"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/ent/validator"
	"github.com/theopenlane/core/v2/internal/entdb"
	emaildef "github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	slackdef "github.com/theopenlane/core/v2/internal/integrations/definitions/slack"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	intruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/internal/keystore"
	coreutils "github.com/theopenlane/core/v2/internal/testutils"
	"github.com/theopenlane/core/v2/internal/workflows"
	"github.com/theopenlane/core/v2/internal/workflows/engine"
	"github.com/theopenlane/core/v2/pkg/entitlements"
	"github.com/theopenlane/core/v2/pkg/entitlements/mocks"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/summarizer"

	_ "github.com/theopenlane/core/v2/internal/ent/generated/runtime"
	_ "github.com/theopenlane/core/v2/internal/ent/historygenerated/runtime"
	"gotest.tools/v3/assert"
)

const (
	fgaModuleFile            = "../../../fga/model/fga.mod"
	seedStripeSubscriptionID = "sub_test_subscription"
	webhookSecret            = "whsec_test_secret"
)

// boolPtr returns a pointer to the provided bool
//
//go:fix inline
func boolPtr(v bool) *bool {
	return new(v)
}

// WorkflowEngineTestSuite handles the setup and teardown between tests
type WorkflowEngineTestSuite struct {
	t                 *testing.T
	ctx               context.Context
	client            *generated.Client
	tf                *dbtestutils.TestFixture
	stripeMockBackend *mocks.MockStripeBackend
	ofgaTF            *fgatest.OpenFGATestFixture
	galaRuntime       *gala.Gala
	integrationsRT    *intruntime.Runtime
	slackMock         *slackdef.MockSlackRuntime
}

// suite is the shared engine test state set up once in TestMain, each test works on its own copy from setupTest
var suite = &WorkflowEngineTestSuite{}

func TestMain(m *testing.M) {
	flag.Parse()

	// only used for setup and teardown, failures there exit through coreutils.RequireNoError
	suite.t = &testing.T{}

	suite.setupSuite()

	exitCode := m.Run()

	suite.tearDownSuite()

	os.Exit(exitCode)
}

// T returns the test currently using the suite
func (s *WorkflowEngineTestSuite) T() *testing.T {
	return s.t
}

// setupTest returns a per-test copy of the suite for tests that use shared state (gala queue, mock senders)
func (s *WorkflowEngineTestSuite) setupTest(t *testing.T) *WorkflowEngineTestSuite {
	return s.withT(t)
}

// setupParallelTest marks the test parallel and returns a per-test copy of the suite
func (s *WorkflowEngineTestSuite) setupParallelTest(t *testing.T) *WorkflowEngineTestSuite {
	t.Parallel()

	return s.withT(t)
}

// withT returns a copy of the suite bound to the given test or subtest
func (s *WorkflowEngineTestSuite) withT(t *testing.T) *WorkflowEngineTestSuite {
	ts := *s
	ts.t = t

	return &ts
}

// setupSuite prepares test dependencies
func (s *WorkflowEngineTestSuite) setupSuite() {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	if testing.Verbose() {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}

	s.ctx = context.Background()

	s.tf = entdb.NewTestFixture()

	version, err := fgaversion.GetVersion()
	coreutils.RequireNoError(s.T(), err)

	s.ofgaTF = fgatest.NewFGATestcontainer(s.ctx,
		fgatest.WithModuleFile(fgaModuleFile),
		fgatest.WithEnvVars(coreutils.GetDefaultFGAEnvs()),
		fgatest.WithVersion(version),
	)

	fgaClient, err := s.ofgaTF.NewFgaClient(s.ctx)
	coreutils.RequireNoError(s.T(), err)

	tm, err := coreutils.CreateTokenManager(-15 * time.Minute) //nolint:mnd
	coreutils.RequireNoError(s.T(), err)

	sm := coreutils.CreateSessionManager()
	rc := coreutils.NewRedisClient()

	sessionConfig := sessions.NewSessionConfig(
		sm,
		sessions.WithPersistence(rc),
	)
	sessionConfig.CookieConfig = sessions.DebugOnlyCookieConfig

	entCfg := &entconfig.Config{
		EmailValidation: validator.EmailVerificationConfig{
			Enabled: false,
		},
		Summarizer: summarizer.Config{
			Type:             summarizer.TypeLexrank,
			MaximumSentences: 60,
		},
		MaxPoolSize: 200, //nolint:mnd
		Modules: entconfig.Modules{
			Enabled:    true,
			UseSandbox: true,
		},
	}

	historyClient, err := entdb.NewTestHistoryClient(s.ctx, s.tf)
	coreutils.RequireNoError(s.T(), err)

	entitlements, err := s.mockStripeClient()
	coreutils.RequireNoError(s.T(), err)

	var pool *gala.Pool
	if entCfg.MaxPoolSize > 0 {
		pool = gala.NewPool(
			gala.WithWorkers(entCfg.MaxPoolSize),
			gala.WithPoolName("ent_client_pool"),
		)
	}

	summarizerClient, err := summarizer.NewSummarizer(entCfg.Summarizer)
	coreutils.RequireNoError(s.T(), err)

	opts := []generated.Option{
		generated.EntConfig(entCfg),
		generated.Authz(*fgaClient),
		generated.TokenManager(tm),
		generated.SessionConfig(&sessionConfig),
		generated.HistoryClient(historyClient),
		generated.EntitlementManager(entitlements),
		generated.Pool(pool),
		generated.Summarizer(summarizerClient),
	}

	jobOpts := []riverqueue.Option{riverqueue.WithConnectionURI(s.tf.URI)}

	runtime, err := gala.NewGala(s.ctx, gala.Config{
		DispatchMode:         gala.DispatchModeDurable,
		ConnectionURI:        s.tf.URI,
		QueueName:            "workflow_engine_test",
		WorkerCount:          5,
		RunMigrations:        true,
		FetchCooldown:        time.Millisecond,
		FetchPollInterval:    10 * time.Millisecond,
		WaitIdlePollInterval: 10 * time.Millisecond,
		CompleteJobsInline:   true,
	})
	coreutils.RequireNoError(s.T(), err)

	workflowCfg := workflows.NewDefaultConfig(workflows.WithEnabled(true), workflows.WithWebhookAllowPrivateAddresses(true))

	db, err := entdb.NewTestClient(s.ctx, s.tf, jobOpts, nil, opts)
	coreutils.RequireNoError(s.T(), err)

	wfEngine, err := engine.NewWorkflowEngineWithConfig(db, runtime, workflowCfg)
	coreutils.RequireNoError(s.T(), err)
	engine.SetDefault(wfEngine)

	db.Use(hooks.EmitGalaEventHook(runtime))

	_, err = gala.Register(runtime, hooks.WorkflowListeners()...)
	coreutils.RequireNoError(s.T(), err)

	coreutils.RequireNoError(s.T(), runtime.Attach(
		gala.WithValue(runtime),
		gala.WithValue(db),
		gala.WithValue(wfEngine),
		gala.WithRestoredValue("ent_client", generated.NewContext),
	))

	coreutils.RequireNoError(s.T(), runtime.StartWorkers(s.ctx))

	s.client = db
	s.galaRuntime = runtime

	credStore, err := keystore.NewStore(db)
	coreutils.RequireNoError(s.T(), err)

	s.slackMock = slackdef.NewMockSlackRuntime()

	rt, err := intruntime.New(intruntime.Config{
		DB:       db,
		Gala:     runtime,
		Keystore: credStore,
		DefinitionBuilders: []registry.Builder{
			emaildef.Builder(emaildef.MockRuntimeConfig(), false),
			s.slackMock.Builder(),
		},
	})
	coreutils.RequireNoError(s.T(), err)

	intruntime.SetDefault(rt)
	s.integrationsRT = rt

	coreutils.RequireNoError(s.T(), wfEngine.SetIntegrationDeps(engine.IntegrationDeps{
		Runtime: rt,
	}))

	s.requireWorkflowSetup(workflowCfg, runtime)
}

// mockSlackRecorder returns the mock Slack message recorder
func (s *WorkflowEngineTestSuite) mockSlackRecorder() *slackdef.MessageRecorder {
	assert.Assert(s.T(), s.slackMock != nil, "slack mock not initialized")
	return s.slackMock.Recorder
}

// mockEmailSender returns the mock email sender from the integration runtime
func (s *WorkflowEngineTestSuite) mockEmailSender() *mockprovider.EmailSender {
	rc, ok := s.integrationsRT.Registry().RuntimeClient(emaildef.DefinitionID.ID())
	assert.Assert(s.T(), ok, "email runtime client not found")

	ms := emaildef.MockSenderFromClient(rc)
	assert.Assert(s.T(), ms != nil, "mock sender not found")

	return ms
}

// tearDownSuite cleans up test dependencies
func (s *WorkflowEngineTestSuite) tearDownSuite() {
	if s.slackMock != nil {
		s.slackMock.Close()
	}

	if s.galaRuntime != nil {
		_ = s.galaRuntime.StopWorkers(context.Background())
		_ = s.galaRuntime.Close()
	}

	if s.client != nil {
		_ = s.client.Close()
	}

	dbtestutils.TeardownFixture(s.tf)

	if s.ofgaTF != nil {
		_ = s.ofgaTF.TeardownFixture()
	}
}

// Client returns the ent client for tests
func (s *WorkflowEngineTestSuite) Client() *generated.Client {
	return s.client
}

// Context returns the context for tests
func (s *WorkflowEngineTestSuite) Context() context.Context {
	return s.ctx
}

// Engine returns the suite's workflow engine initialized in SetupSuite
func (s *WorkflowEngineTestSuite) Engine() *engine.WorkflowEngine {
	wfEngine := engine.Default()
	assert.Assert(s.T(), wfEngine != nil, "workflow engine not initialized")

	return wfEngine
}

// NewIsolatedEngine creates a new workflow engine with a custom Gala runtime for tests that need isolation (e.g., failure testing)
func (s *WorkflowEngineTestSuite) NewIsolatedEngine(runtime *gala.Gala) *engine.WorkflowEngine {
	wfEngine, err := engine.NewWorkflowEngine(s.client, runtime)
	assert.NilError(s.T(), err)

	return wfEngine
}

func (s *WorkflowEngineTestSuite) requireWorkflowSetup(cfg *workflows.Config, runtime *gala.Gala) {
	if s.client == nil {
		coreutils.FailNow(s.T(), "ent client not initialized")
	}

	if runtime == nil {
		coreutils.FailNow(s.T(), "gala runtime not initialized")
	}

	if engine.Default() == nil {
		coreutils.FailNow(s.T(), "workflow engine not initialized")
	}

	if !runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernWorkflow, generated.TypeControl), ent.OpCreate.String()) {
		coreutils.FailNow(s.T(), "mutation listeners not registered")
	}

	if cfg != nil && cfg.Enabled && !runtime.InterestedIn(engine.WorkflowTriggeredEventTopic.Name, "") {
		coreutils.FailNow(s.T(), "workflow listeners not registered")
	}

	eligible := workflows.EligibleWorkflowFields(enums.WorkflowObjectTypeControl)
	if len(eligible) == 0 {
		coreutils.FailNow(s.T(), "workflow eligible fields not registered")
	}

	if _, ok := eligible["reference_id"]; !ok {
		coreutils.FailNow(s.T(), "workflow eligible fields missing control.reference_id")
	}
}

// WaitForEvents blocks until runnable and in-flight Gala jobs complete
func (s *WorkflowEngineTestSuite) WaitForEvents() {
	assert.NilError(s.T(), s.galaRuntime.WaitIdle(s.T().Context()))
}

// SetupSystemAdmin creates a system admin user and returns user ID, org ID, and admin context
func (s *WorkflowEngineTestSuite) SetupSystemAdmin() (string, string, context.Context) {
	internalCtx := generated.NewContext(auth.WithInternalCrossOrgContext(s.ctx), s.client)

	user, err := s.client.User.Create().
		SetEmail("admin-" + ulids.New().String() + "@example.com").
		SetLastLoginProvider(enums.AuthProviderCredentials).
		Save(internalCtx)
	assert.NilError(s.T(), err)

	personalOrg, err := user.Edges.Setting.DefaultOrg(internalCtx)
	assert.NilError(s.T(), err)

	userCtx := auth.NewTestContextWithOrgID(user.ID, personalOrg.ID)
	setCtx := generated.NewContext(auth.WithInternalCrossOrgContext(userCtx), s.client)

	testOrg, err := s.client.Organization.Create().
		SetName("Admin Organization " + ulids.New().String()).
		Save(setCtx)
	assert.NilError(s.T(), err)

	s.enableModules(user.ID, testOrg.ID)

	adminCtx := auth.NewTestContextForSystemAdmin(user.ID, testOrg.ID)
	adminCtx = generated.NewContext(adminCtx, s.client)

	return user.ID, testOrg.ID, adminCtx
}

// SetupTestUser creates a test user and organization, returning user ID, org ID, and context
func (s *WorkflowEngineTestSuite) SetupTestUser() (string, string, context.Context) {
	internalCtx := generated.NewContext(auth.WithInternalCrossOrgContext(s.ctx), s.client)

	user, err := s.client.User.Create().
		SetEmail("test-" + ulids.New().String() + "@example.com").
		SetLastLoginProvider(enums.AuthProviderCredentials).
		Save(internalCtx)
	assert.NilError(s.T(), err)

	personalOrg, err := user.Edges.Setting.DefaultOrg(internalCtx)
	assert.NilError(s.T(), err)

	userCtx := auth.NewTestContextWithOrgID(user.ID, personalOrg.ID)
	setCtx := generated.NewContext(auth.WithInternalCrossOrgContext(userCtx), s.client)

	testOrg, err := s.client.Organization.Create().
		SetName("Test Organization " + ulids.New().String()).
		Save(setCtx)
	assert.NilError(s.T(), err)

	s.enableModules(user.ID, testOrg.ID)

	userCtx = auth.NewTestContextWithOrgID(user.ID, testOrg.ID)
	userCtx = generated.NewContext(userCtx, s.client)

	return user.ID, testOrg.ID, userCtx
}

// enableModules enables all org modules for the given organization (following httpserve test pattern)
func (s *WorkflowEngineTestSuite) enableModules(userID, orgID string) {
	features := models.AllOrgModules

	userCtx := auth.NewTestContextWithOrgID(userID, orgID)
	// Run seeding operations as an internal operation
	userCtx = auth.WithInternalOperationContext(userCtx)
	// Add client to context
	userCtx = generated.NewContext(userCtx, s.client)

	for _, feature := range features {
		_, err := s.client.OrgModule.Create().
			SetOwnerID(orgID).
			SetModule(feature).
			SetActive(true).
			SetPrice(models.Price{Amount: 0, Interval: "month"}).
			Save(userCtx)
		assert.NilError(s.T(), err)
	}

	err := entitlements.CreateFeatureTuples(userCtx, &s.client.Authz, orgID, features)
	assert.NilError(s.T(), err)
}

// CreateTestUserID creates an additional test user and returns the ID
func (s *WorkflowEngineTestSuite) CreateTestUserID() string {
	internalCtx := generated.NewContext(auth.WithInternalCrossOrgContext(s.ctx), s.client)

	user, err := s.client.User.Create().
		SetEmail("test-" + ulids.New().String() + "@example.com").
		SetLastLoginProvider(enums.AuthProviderCredentials).
		Save(internalCtx)
	assert.NilError(s.T(), err)

	return user.ID
}

// CreateTestUserInOrg creates a user and attaches them to the provided organization
func (s *WorkflowEngineTestSuite) CreateTestUserInOrg(orgID string, role enums.Role) (string, context.Context) {
	internalCtx := generated.NewContext(auth.WithInternalCrossOrgContext(s.ctx), s.client)

	user, err := s.client.User.Create().
		SetEmail("test-" + ulids.New().String() + "@example.com").
		SetLastLoginProvider(enums.AuthProviderCredentials).
		Save(internalCtx)
	assert.NilError(s.T(), err)

	authCtx := auth.NewTestContextWithOrgID(user.ID, orgID)
	membershipCtx := generated.NewContext(auth.WithInternalCrossOrgContext(authCtx), s.client)

	_, err = s.client.OrgMembership.Create().
		SetOrganizationID(orgID).
		SetUserID(user.ID).
		SetRole(role).
		Save(membershipCtx)
	assert.NilError(s.T(), err)

	userCtx := auth.NewTestContextWithOrgID(user.ID, orgID)
	userCtx = generated.NewContext(userCtx, s.client)

	return user.ID, userCtx
}

// CreateTestWorkflowDefinition creates a basic workflow definition for testing
func (s *WorkflowEngineTestSuite) CreateTestWorkflowDefinition(ctx context.Context, orgID string) *generated.WorkflowDefinition {
	doc := models.WorkflowDefinitionDocument{
		Triggers: []models.WorkflowTrigger{
			{Operation: "UPDATE", Fields: []string{}},
		},
		Conditions: []models.WorkflowCondition{},
		Actions:    []models.WorkflowAction{},
	}

	builder := s.client.WorkflowDefinition.Create().
		SetName("Test Workflow " + ulids.New().String()).
		SetWorkflowKind(enums.WorkflowKindNotification).
		SetSchemaType("Control").
		SetActive(true).
		SetDraft(false).
		SetOwnerID(orgID).
		SetDefinitionJSON(doc)

	operations, fields := workflows.DeriveTriggerPrefilter(doc)
	if len(operations) == 0 {
		builder.SetTriggerOperations(nil)
	} else {
		builder.SetTriggerOperations(operations)
	}
	if len(fields) == 0 {
		builder.SetTriggerFields(nil)
	} else {
		builder.SetTriggerFields(fields)
	}

	def, err := builder.Save(ctx)
	assert.NilError(s.T(), err)

	return def
}

// CreateTestWorkflowDefinitionWithPrefilter creates a workflow definition with custom triggers and prefilter
func (s *WorkflowEngineTestSuite) CreateTestWorkflowDefinitionWithPrefilter(ctx context.Context, orgID string, triggers []models.WorkflowTrigger, operations []string, fields []string) *generated.WorkflowDefinition {
	builder := s.client.WorkflowDefinition.Create().
		SetName("Test Workflow - " + ulids.New().String()).
		SetWorkflowKind(enums.WorkflowKindNotification).
		SetSchemaType("Control").
		SetActive(true).
		SetDraft(false).
		SetOwnerID(orgID).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{
			Triggers:   triggers,
			Conditions: []models.WorkflowCondition{},
			Actions: []models.WorkflowAction{
				{Key: "test_action", Type: "notification", Params: []byte(`{"title": "test"}`)},
			},
		})

	if operations != nil {
		builder = builder.SetTriggerOperations(operations)
	}

	if fields != nil {
		builder = builder.SetTriggerFields(fields)
	}

	def, err := builder.Save(ctx)
	assert.NilError(s.T(), err)

	return def
}

// ApplyTriggerPrefilter applies the derived trigger prefilter to a workflow definition update
func (s *WorkflowEngineTestSuite) ApplyTriggerPrefilter(update *generated.WorkflowDefinitionUpdateOne, doc models.WorkflowDefinitionDocument) *generated.WorkflowDefinitionUpdateOne {
	operations, fields := workflows.DeriveTriggerPrefilter(doc)
	if len(operations) == 0 {
		update.SetTriggerOperations(nil)
	} else {
		update.SetTriggerOperations(operations)
	}
	if len(fields) == 0 {
		update.SetTriggerFields(nil)
	} else {
		update.SetTriggerFields(fields)
	}
	return update
}

// UpdateWorkflowDefinition updates a workflow definition with privacy bypass for testing
func (s *WorkflowEngineTestSuite) UpdateWorkflowDefinition(def *generated.WorkflowDefinition, doc models.WorkflowDefinitionDocument) *generated.WorkflowDefinition {
	internalCtx := generated.NewContext(auth.WithInternalCrossOrgContext(s.ctx), s.client)

	updated, err := def.Update().
		SetDefinitionJSON(doc).
		Save(internalCtx)
	assert.NilError(s.T(), err)

	return updated
}

// UpdateWorkflowDefinitionWithPrefilter updates a workflow definition with prefilter and privacy bypass
func (s *WorkflowEngineTestSuite) UpdateWorkflowDefinitionWithPrefilter(def *generated.WorkflowDefinition, doc models.WorkflowDefinitionDocument) *generated.WorkflowDefinition {
	internalCtx := generated.NewContext(auth.WithInternalCrossOrgContext(s.ctx), s.client)

	update := def.Update().SetDefinitionJSON(doc)
	update = s.ApplyTriggerPrefilter(update, doc)

	updated, err := update.Save(internalCtx)
	assert.NilError(s.T(), err)

	return updated
}

// UpdateWorkflowDefinitionInactive updates a workflow definition to inactive with prefilter and privacy bypass
func (s *WorkflowEngineTestSuite) UpdateWorkflowDefinitionInactive(def *generated.WorkflowDefinition, doc models.WorkflowDefinitionDocument) *generated.WorkflowDefinition {
	internalCtx := generated.NewContext(auth.WithInternalCrossOrgContext(s.ctx), s.client)

	update := def.Update().SetDefinitionJSON(doc).SetActive(false)
	update = s.ApplyTriggerPrefilter(update, doc)

	updated, err := update.Save(internalCtx)
	assert.NilError(s.T(), err)

	return updated
}

// ClearWorkflowDefinitionsForOrg removes all workflow definitions and related entities for a specific organization
func (s *WorkflowEngineTestSuite) ClearWorkflowDefinitionsForOrg(orgID string) {
	internalCtx := generated.NewContext(auth.WithInternalCrossOrgContext(s.ctx), s.client)

	_, err := s.client.WorkflowEvent.Delete().Where(workflowevent.OwnerIDEQ(orgID)).Exec(internalCtx)
	assert.NilError(s.T(), err)

	_, err = s.client.WorkflowAssignmentTarget.Delete().Where(workflowassignmenttarget.OwnerIDEQ(orgID)).Exec(internalCtx)
	assert.NilError(s.T(), err)

	_, err = s.client.WorkflowAssignment.Delete().Where(workflowassignment.OwnerIDEQ(orgID)).Exec(internalCtx)
	assert.NilError(s.T(), err)

	err = s.client.WorkflowInstance.Update().Where(workflowinstance.OwnerIDEQ(orgID)).ClearWorkflowProposal().Exec(internalCtx)
	assert.NilError(s.T(), err)

	_, err = s.client.WorkflowProposal.Delete().Where(workflowproposal.OwnerIDEQ(orgID)).Exec(internalCtx)
	assert.NilError(s.T(), err)

	_, err = s.client.WorkflowObjectRef.Delete().Where(workflowobjectref.OwnerIDEQ(orgID)).Exec(internalCtx)
	assert.NilError(s.T(), err)

	_, err = s.client.WorkflowInstance.Delete().Where(workflowinstance.OwnerIDEQ(orgID)).Exec(internalCtx)
	assert.NilError(s.T(), err)

	_, err = s.client.WorkflowDefinition.Delete().Where(workflowdefinition.OwnerIDEQ(orgID)).Exec(internalCtx)
	assert.NilError(s.T(), err)
}

// InternalContext returns an internal context with privacy bypass for test setup operations
func (s *WorkflowEngineTestSuite) InternalContext() context.Context {
	return generated.NewContext(auth.WithInternalCrossOrgContext(s.ctx), s.client)
}

// SeedContext creates a context with auth info and privacy bypass for seeding test data
func (s *WorkflowEngineTestSuite) SeedContext(userID, orgID string) context.Context {
	ctx := auth.NewTestContextWithOrgID(userID, orgID)
	ctx = auth.WithInternalOperationContext(ctx)
	ctx = generated.NewContext(ctx, s.client)
	ctxClient := generated.FromContext(ctx)
	assert.Assert(s.T(), ctxClient != nil, "seed context missing ent client")
	assert.Assert(s.T(), engine.Default() != nil, "workflow engine not initialized")
	seedCaller, seedCallerOk := auth.CallerFromContext(ctx)
	assert.Assert(s.T(), seedCallerOk && seedCaller != nil && seedCaller.OrganizationID != "", "seed context missing org")
	return ctx
}

// TriggerInstance creates a workflow instance via the engine to avoid manual instance construction in tests
func (s *WorkflowEngineTestSuite) TriggerInstance(ctx context.Context, wfEngine *engine.WorkflowEngine, def *generated.WorkflowDefinition, obj *workflows.Object, input engine.TriggerInput) *generated.WorkflowInstance {
	instance, err := wfEngine.TriggerWorkflow(ctx, def, obj, input)
	assert.NilError(s.T(), err)
	assert.Assert(s.T(), instance != nil)
	return instance
}

// CreateApprovalWorkflowDefinition creates an approval workflow definition for testing
func (s *WorkflowEngineTestSuite) CreateApprovalWorkflowDefinition(ctx context.Context, orgID string, action models.WorkflowAction) *generated.WorkflowDefinition {
	triggerFields := []string{"status"}
	if len(action.Params) > 0 {
		var params workflows.ApprovalActionParams
		err := json.Unmarshal(action.Params, &params)
		assert.NilError(s.T(), err, "failed to parse approval action params")
		fields := entityops.NormalizeStrings(params.Fields)
		if len(fields) > 0 {
			triggerFields = fields
		}
	}
	sort.Strings(triggerFields)

	doc := models.WorkflowDefinitionDocument{
		ApprovalSubmissionMode: enums.WorkflowApprovalSubmissionModeAutoSubmit,
		Triggers: []models.WorkflowTrigger{
			{Operation: "UPDATE", Fields: triggerFields},
		},
		Conditions: []models.WorkflowCondition{
			{Expression: "true"},
		},
		Actions: []models.WorkflowAction{action},
	}

	operations, fields := workflows.DeriveTriggerPrefilter(doc)
	fields = triggerFields

	def, err := s.client.WorkflowDefinition.Create().
		SetName("Approval Workflow " + ulids.New().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetDraft(false).
		SetOwnerID(orgID).
		SetTriggerOperations(operations).
		SetTriggerFields(fields).
		SetApprovalSubmissionMode(enums.WorkflowApprovalSubmissionModeAutoSubmit).
		SetDefinitionJSON(doc).
		Save(ctx)
	assert.NilError(s.T(), err)

	return def
}

// mockStripeClient creates a new stripe client with mock backend
func (suite *WorkflowEngineTestSuite) mockStripeClient() (*entitlements.StripeClient, error) {
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

// mockCustomer is the mock customer used for webhook tests
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

var subsID = ulids.New().String()
var custID = ulids.New().String()

var mockSubscription = &stripe.Subscription{
	ID:     subsID,
	Status: "active",
	Items: &stripe.SubscriptionItemList{
		Data: mockItems,
	},
	Metadata: map[string]string{
		"organization_id": ulids.New().String(),
	},
	Customer: &stripe.Customer{
		ID: custID,
	},
	TrialEnd:     time.Now().Add(7 * 24 * time.Hour).Unix(),
	DaysUntilDue: 15,
}

var mockProduct = &stripe.Product{
	ID:   "prod_test_product",
	Name: "Test Product",
}

// orgSubscriptionMocks mocks the stripe calls for org subscription during the webhook tests
func (suite *WorkflowEngineTestSuite) orgSubscriptionMocks() {
	suite.stripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.CustomerRetrieveParams"), mock.AnythingOfType("*stripe.Customer")).Run(func(args mock.Arguments) {
		mockCustomerSearchResult := args.Get(4).(*stripe.Customer)

		*mockCustomerSearchResult = *mockCustomer

	}).Return(nil)

	suite.stripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.CustomerCreateParams"), mock.AnythingOfType("*stripe.Customer")).Run(func(args mock.Arguments) {
		mockCustomerSearchResult := args.Get(4).(*stripe.Customer)

		*mockCustomerSearchResult = *mockCustomer

	}).Return(nil)

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

	suite.stripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.SubscriptionScheduleCreateParams"), mock.AnythingOfType("*stripe.SubscriptionSchedule")).Run(func(args mock.Arguments) {
		mockSubscriptionScheduleResult := args.Get(4).(*stripe.SubscriptionSchedule)

		*mockSubscriptionScheduleResult = stripe.SubscriptionSchedule{
			ID:     "sched_test_schedule",
			Status: "active",
		}

	}).Return(nil)

	suite.stripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.CustomerUpdateParams"), mock.AnythingOfType("*stripe.Customer")).Run(func(args mock.Arguments) {
		mockCustomerUpdateResult := args.Get(4).(*stripe.Customer)

		*mockCustomerUpdateResult = *mockCustomer

	}).Return(nil)
}
