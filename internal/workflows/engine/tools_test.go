//go:build test

package engine_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/stripe/stripe-go/v84"
	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/iam/auth"
	fgatest "github.com/theopenlane/iam/fgax/testutils"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/riverboat/pkg/riverqueue"
	dbtestutils "github.com/theopenlane/utils/testutils"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/entconfig"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/internal/entdb"
	coreutils "github.com/theopenlane/core/internal/testutils"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/entitlements/mocks"
	"github.com/theopenlane/core/pkg/events/soiree"

	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
	_ "github.com/theopenlane/core/internal/ent/historygenerated/runtime"
)

const (
	fgaModelFile             = "../../../fga/model/model.fga"
	seedStripeSubscriptionID = "sub_test_subscription"
	webhookSecret            = "whsec_test_secret"
)

// boolPtr returns a pointer to the provided bool
func boolPtr(v bool) *bool {
	return &v
}

// WorkflowEngineTestSuite handles the setup and teardown between tests
type WorkflowEngineTestSuite struct {
	suite.Suite
	ctx               context.Context
	client            *generated.Client
	tf                *dbtestutils.TestFixture
	stripeMockBackend *mocks.MockStripeBackend
	ofgaTF            *fgatest.OpenFGATestFixture
}

// TestWorkflowEngineTestSuite runs all the tests in the WorkflowEngineTestSuite
func TestWorkflowEngineTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowEngineTestSuite))
}

// SetupSuite prepares test dependencies
func (s *WorkflowEngineTestSuite) SetupSuite() {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	if testing.Verbose() {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}

	s.ctx = context.Background()

	// setup db container
	s.tf = entdb.NewTestFixture()

	// setup openFGA container
	s.ofgaTF = fgatest.NewFGATestcontainer(s.ctx,
		fgatest.WithModelFile(fgaModelFile),
		fgatest.WithEnvVars(map[string]string{
			"OPENFGA_MAX_CHECKS_PER_BATCH_CHECK":          "100",
			"OPENFGA_CHECK_ITERATOR_CACHE_ENABLED":        "false",
			"OPENFGA_LIST_OBJECTS_ITERATOR_CACHE_ENABLED": "false",
		}),
	)

	fgaClient, err := s.ofgaTF.NewFgaClient(s.ctx)
	s.Require().NoError(err)

	tm, err := coreutils.CreateTokenManager(-15 * time.Minute) //nolint:mnd
	s.Require().NoError(err)

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
		Modules: entconfig.Modules{
			Enabled:    true,
			UseSandbox: true,
		},
	}

	historyClient, err := entdb.NewTestHistoryClient(s.ctx, s.tf)
	s.Require().NoError(err)

	entitlements, err := s.mockStripeClient()

	pool := soiree.NewPool(
		soiree.WithWorkers(100), //nolint:mnd
		soiree.WithPoolName("ent_client_pool"),
	)

	opts := []generated.Option{
		generated.EntConfig(entCfg),
		generated.Authz(*fgaClient),
		generated.TokenManager(tm),
		generated.SessionConfig(&sessionConfig),
		generated.Emailer(&emailtemplates.Config{}),
		generated.HistoryClient(historyClient),
		generated.EntitlementManager(entitlements),
		generated.Pool(pool),
	}

	jobOpts := []riverqueue.Option{riverqueue.WithConnectionURI(s.tf.URI)}

	db, err := entdb.NewTestClient(s.ctx, s.tf, jobOpts, opts)
	s.Require().NoError(err)

	s.client = db
}

// TearDownSuite cleans up test dependencies
func (s *WorkflowEngineTestSuite) TearDownSuite() {
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

// NewTestEngine creates a new workflow engine for testing
func (s *WorkflowEngineTestSuite) NewTestEngine(emitter soiree.Emitter) *engine.WorkflowEngine {
	wfEngine, err := engine.NewWorkflowEngine(s.client, emitter)
	s.Require().NoError(err)

	return wfEngine
}

// SetupSystemAdmin creates a system admin user and returns user ID, org ID, and admin context
// Use this for tests that rely on ent hooks which need elevated permissions for internal queries
func (s *WorkflowEngineTestSuite) SetupSystemAdmin() (string, string, context.Context) {
	internalCtx := generated.NewContext(rule.WithInternalContext(s.ctx), s.client)

	user, err := s.client.User.Create().
		SetEmail("admin-" + ulids.New().String() + "@example.com").
		SetLastLoginProvider(enums.AuthProviderCredentials).
		Save(internalCtx)
	s.Require().NoError(err)

	personalOrg, err := user.Edges.Setting.DefaultOrg(internalCtx)
	s.Require().NoError(err)

	userCtx := auth.NewTestContextWithOrgID(user.ID, personalOrg.ID)
	setCtx := generated.NewContext(rule.WithInternalContext(userCtx), s.client)

	testOrg, err := s.client.Organization.Create().
		SetName("Admin Organization " + ulids.New().String()).
		Save(setCtx)
	s.Require().NoError(err)

	// Enable all modules for the org (required for workflow operations)
	s.enableModules(user.ID, testOrg.ID)

	// Create system admin context
	adminCtx := auth.NewTestContextForSystemAdmin(user.ID, testOrg.ID)
	adminCtx = generated.NewContext(adminCtx, s.client)

	return user.ID, testOrg.ID, adminCtx
}

// SetupTestUser creates a test user and organization, returning user ID, org ID, and context
func (s *WorkflowEngineTestSuite) SetupTestUser() (string, string, context.Context) {
	internalCtx := generated.NewContext(rule.WithInternalContext(s.ctx), s.client)

	user, err := s.client.User.Create().
		SetEmail("test-" + ulids.New().String() + "@example.com").
		SetLastLoginProvider(enums.AuthProviderCredentials).
		Save(internalCtx)
	s.Require().NoError(err)

	personalOrg, err := user.Edges.Setting.DefaultOrg(internalCtx)
	s.Require().NoError(err)

	userCtx := auth.NewTestContextWithOrgID(user.ID, personalOrg.ID)
	setCtx := generated.NewContext(rule.WithInternalContext(userCtx), s.client)

	testOrg, err := s.client.Organization.Create().
		SetName("Test Organization " + ulids.New().String()).
		Save(setCtx)
	s.Require().NoError(err)

	// Enable all modules for the org (required for workflow operations)
	s.enableModules(user.ID, testOrg.ID)

	userCtx = auth.NewTestContextWithOrgID(user.ID, testOrg.ID)
	userCtx = generated.NewContext(userCtx, s.client)

	return user.ID, testOrg.ID, userCtx
}

// enableModules enables all org modules for the given organization (following httpserve test pattern)
func (s *WorkflowEngineTestSuite) enableModules(userID, orgID string) {
	features := models.AllOrgModules

	// Create authenticated context with user and org IDs
	userCtx := auth.NewTestContextWithOrgID(userID, orgID)
	// Set privacy allow for seeding operations
	userCtx = privacy.DecisionContext(userCtx, privacy.Allow)
	// Add client to context
	userCtx = generated.NewContext(userCtx, s.client)

	// Create org modules for each feature
	for _, feature := range features {
		_, err := s.client.OrgModule.Create().
			SetOwnerID(orgID).
			SetModule(feature).
			SetActive(true).
			SetPrice(models.Price{Amount: 0, Interval: "month"}).
			Save(userCtx)
		s.Require().NoError(err)
	}

	// Create FGA tuples for the features
	err := entitlements.CreateFeatureTuples(userCtx, &s.client.Authz, orgID, features)
	s.Require().NoError(err)
}

// CreateTestUserID creates an additional test user and returns the ID.
func (s *WorkflowEngineTestSuite) CreateTestUserID() string {
	internalCtx := generated.NewContext(rule.WithInternalContext(s.ctx), s.client)

	user, err := s.client.User.Create().
		SetEmail("test-" + ulids.New().String() + "@example.com").
		SetLastLoginProvider(enums.AuthProviderCredentials).
		Save(internalCtx)
	s.Require().NoError(err)

	return user.ID
}

// CreateTestUserInOrg creates a user and attaches them to the provided organization.
func (s *WorkflowEngineTestSuite) CreateTestUserInOrg(orgID string, role enums.Role) (string, context.Context) {
	internalCtx := generated.NewContext(rule.WithInternalContext(s.ctx), s.client)

	user, err := s.client.User.Create().
		SetEmail("test-" + ulids.New().String() + "@example.com").
		SetLastLoginProvider(enums.AuthProviderCredentials).
		Save(internalCtx)
	s.Require().NoError(err)

	// OrgMembership hooks require authenticated user context for creating managed groups
	authCtx := auth.NewTestContextWithOrgID(user.ID, orgID)
	membershipCtx := generated.NewContext(rule.WithInternalContext(authCtx), s.client)

	_, err = s.client.OrgMembership.Create().
		SetOrganizationID(orgID).
		SetUserID(user.ID).
		SetRole(role).
		Save(membershipCtx)
	s.Require().NoError(err)

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
	s.Require().NoError(err)

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
	s.Require().NoError(err)

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
	internalCtx := generated.NewContext(rule.WithInternalContext(s.ctx), s.client)

	updated, err := def.Update().
		SetDefinitionJSON(doc).
		Save(internalCtx)
	s.Require().NoError(err)

	return updated
}

// UpdateWorkflowDefinitionWithPrefilter updates a workflow definition with prefilter and privacy bypass
func (s *WorkflowEngineTestSuite) UpdateWorkflowDefinitionWithPrefilter(def *generated.WorkflowDefinition, doc models.WorkflowDefinitionDocument) *generated.WorkflowDefinition {
	internalCtx := generated.NewContext(rule.WithInternalContext(s.ctx), s.client)

	update := def.Update().SetDefinitionJSON(doc)
	update = s.ApplyTriggerPrefilter(update, doc)

	updated, err := update.Save(internalCtx)
	s.Require().NoError(err)

	return updated
}

// UpdateWorkflowDefinitionInactive updates a workflow definition to inactive with prefilter and privacy bypass
func (s *WorkflowEngineTestSuite) UpdateWorkflowDefinitionInactive(def *generated.WorkflowDefinition, doc models.WorkflowDefinitionDocument) *generated.WorkflowDefinition {
	internalCtx := generated.NewContext(rule.WithInternalContext(s.ctx), s.client)

	update := def.Update().SetDefinitionJSON(doc).SetActive(false)
	update = s.ApplyTriggerPrefilter(update, doc)

	updated, err := update.Save(internalCtx)
	s.Require().NoError(err)

	return updated
}

// ClearWorkflowDefinitions removes all workflow definitions with privacy bypass for testing
func (s *WorkflowEngineTestSuite) ClearWorkflowDefinitions() {
	internalCtx := generated.NewContext(rule.WithInternalContext(s.ctx), s.client)
	_, err := s.client.WorkflowDefinition.Delete().Exec(internalCtx)
	s.Require().NoError(err)
}

// InternalContext returns an internal context with privacy bypass for test setup operations
func (s *WorkflowEngineTestSuite) InternalContext() context.Context {
	return generated.NewContext(rule.WithInternalContext(s.ctx), s.client)
}

// SeedContext creates a context with auth info and privacy bypass for seeding test data.
// Use this for creating test entities (Groups, Controls, etc.) that require privacy bypass.
// The returned context has:
// - User authentication info (userID, orgID)
// - Privacy.Allow decision
// - Ent client
func (s *WorkflowEngineTestSuite) SeedContext(userID, orgID string) context.Context {
	ctx := auth.NewTestContextWithOrgID(userID, orgID)
	ctx = privacy.DecisionContext(ctx, privacy.Allow)
	ctx = generated.NewContext(ctx, s.client)
	return ctx
}

// syncBusEmitter is a synchronous event emitter for testing
type syncBusEmitter struct {
	bus *soiree.EventBus
}

// Emit publishes an event synchronously
func (s *syncBusEmitter) Emit(topic string, event any) <-chan error {
	if s == nil || s.bus == nil {
		ch := make(chan error, 1)
		close(ch)
		return ch
	}

	errCh := s.bus.Emit(topic, event)
	errs := make([]error, 0)
	for err := range errCh {
		if err != nil {
			errs = append(errs, err)
		}
	}

	ch := make(chan error, len(errs))
	for _, err := range errs {
		ch <- err
	}
	close(ch)
	return ch
}

// mockEventEmitter is a mock event emitter for testing
type mockEventEmitter struct {
	emittedEvents   []string
	emittedPayloads []any
	clientSet       bool
}

// Emit publishes an event for testing
func (m *mockEventEmitter) Emit(topic string, event any) <-chan error {
	m.emittedEvents = append(m.emittedEvents, topic)
	m.emittedPayloads = append(m.emittedPayloads, event)

	// Verify client is set on event
	if baseEvent, ok := event.(*soiree.BaseEvent); ok {
		m.clientSet = baseEvent.Client() != nil
	}

	ch := make(chan error, 1)
	close(ch)

	return ch
}

// SetupWorkflowEngineWithListeners creates a workflow engine with all listeners registered
func (s *WorkflowEngineTestSuite) SetupWorkflowEngineWithListeners() *engine.WorkflowEngine {
	bus := soiree.New()
	emitter := &syncBusEmitter{bus: bus}

	wfEngine := s.NewTestEngine(emitter)

	listeners := engine.NewWorkflowListeners(s.client, wfEngine, emitter)

	bindings := []soiree.ListenerBinding{
		soiree.BindListener(soiree.WorkflowTriggeredTopic, func(ctx *soiree.EventContext, payload soiree.WorkflowTriggeredPayload) error {
			return listeners.HandleWorkflowTriggered(ctx, payload)
		}),
		soiree.BindListener(soiree.WorkflowActionStartedTopic, func(ctx *soiree.EventContext, payload soiree.WorkflowActionStartedPayload) error {
			return listeners.HandleActionStarted(ctx, payload)
		}),
		soiree.BindListener(soiree.WorkflowActionCompletedTopic, func(ctx *soiree.EventContext, payload soiree.WorkflowActionCompletedPayload) error {
			return listeners.HandleActionCompleted(ctx, payload)
		}),
		soiree.BindListener(soiree.WorkflowAssignmentCompletedTopic, func(ctx *soiree.EventContext, payload soiree.WorkflowAssignmentCompletedPayload) error {
			return listeners.HandleAssignmentCompleted(ctx, payload)
		}),
		soiree.BindListener(soiree.WorkflowInstanceCompletedTopic, func(ctx *soiree.EventContext, payload soiree.WorkflowInstanceCompletedPayload) error {
			return listeners.HandleInstanceCompleted(ctx, payload)
		}),
	}

	for _, binding := range bindings {
		_, err := binding.Register(bus)
		require.NoError(s.T(), err)
	}

	return wfEngine
}

// CreateApprovalWorkflowDefinition creates an approval workflow definition for testing
func (s *WorkflowEngineTestSuite) CreateApprovalWorkflowDefinition(ctx context.Context, orgID string, action models.WorkflowAction) *generated.WorkflowDefinition {
	def, err := s.client.WorkflowDefinition.Create().
		SetName("Approval Workflow " + ulids.New().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetOwnerID(orgID).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{
			Triggers: []models.WorkflowTrigger{
				{Operation: "UPDATE", Fields: []string{"status"}},
			},
			Conditions: []models.WorkflowCondition{
				{Expression: "true"},
			},
			Actions: []models.WorkflowAction{action},
		}).
		Save(ctx)
	s.Require().NoError(err)

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
func (suite *WorkflowEngineTestSuite) orgSubscriptionMocks() {
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
