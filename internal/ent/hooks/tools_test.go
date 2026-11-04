//go:build test

package hooks_test

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/rs/zerolog"
	"github.com/theopenlane/core/v2/fga/fgaversion"
	"github.com/theopenlane/core/v2/internal/ent/entconfig"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/migrate"
	_ "github.com/theopenlane/core/v2/internal/ent/generated/runtime"
	"github.com/theopenlane/core/v2/internal/ent/generated/user"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/entdb"
	coreutils "github.com/theopenlane/core/v2/internal/testutils"
	"github.com/theopenlane/core/v2/pkg/entitlements"
	"github.com/theopenlane/core/v2/pkg/gala"
	authmw "github.com/theopenlane/core/v2/pkg/middleware/auth"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	fgatest "github.com/theopenlane/iam/fgax/testutils"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/utils/testutils"
	"gotest.tools/v3/assert"
)

const (
	fgaModuleFile             = "../../../fga/model/fga.mod"
	cnameTargetTest           = "cname.test.net"
	previewMappableZoneIDTest = "preview-zone-id"
)

// HookTestSuite handles the setup and teardown between tests
type HookTestSuite struct {
	t           *testing.T
	client      *generated.Client
	tf          *testutils.TestFixture
	ofgaTF      *fgatest.OpenFGATestFixture
	galaRuntime *gala.Gala
}

// suite is the shared hook test state set up once in TestMain, each test works on its own copy from setupTest
var suite = &HookTestSuite{}

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
func (s *HookTestSuite) T() *testing.T {
	return s.t
}

// setupTest returns a per-test copy of the suite bound to the given test
func (s *HookTestSuite) setupTest(t *testing.T) *HookTestSuite {
	return s.withT(t)
}

// withT returns a copy of the suite bound to the given test or subtest
func (s *HookTestSuite) withT(t *testing.T) *HookTestSuite {
	ts := *s
	ts.t = t

	return &ts
}

func (s *HookTestSuite) waitForEvents() {
	assert.NilError(s.T(), s.galaRuntime.WaitIdle(s.T().Context()))
}

// setupSuite runs before the test suite
func (s *HookTestSuite) setupSuite() {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	if testing.Verbose() {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}

	s.client = s.setupClient()

	hooks.SetTrustCenterConfig(hooks.TrustCenterConfig{
		CnameTarget:   cnameTargetTest,
		PreviewZoneID: previewMappableZoneIDTest,
	})

	ctx := auth.WithInternalOperationContext(context.Background())
	_, err := s.client.MappableDomain.Create().
		SetName(cnameTargetTest).
		SetZoneID(previewMappableZoneIDTest).
		Save(ctx)
	coreutils.RequireNoError(s.T(), err)
}

// tearDownSuite runs after the test suite
func (s *HookTestSuite) tearDownSuite() {
	t := s.T()

	// close the gala runtime
	err := s.galaRuntime.Close()
	coreutils.RequireNoError(t, err)

	// close the database connection
	err = s.client.Close()
	coreutils.RequireNoError(t, err)

	// close the database container
	testutils.TeardownFixture(s.tf)
}

// setupClient sets up the client for the test suite
// using enttest
func (s *HookTestSuite) setupClient() *generated.Client {
	t := s.T()

	version, err := fgaversion.GetVersion()
	coreutils.RequireNoError(t, err)

	s.ofgaTF = fgatest.NewFGATestcontainer(context.Background(),
		fgatest.WithModuleFile(fgaModuleFile),
		fgatest.WithEnvVars(coreutils.GetDefaultFGAEnvs()),
		fgatest.WithVersion(version),
	)

	ctx := context.Background()

	fgaClient, err := s.ofgaTF.NewFgaClient(ctx)
	coreutils.RequireNoError(t, err)

	tm, err := coreutils.CreateTokenManager(-15 * time.Minute) //nolint:mnd
	coreutils.RequireNoError(t, err)

	sm := coreutils.CreateSessionManager()
	rc := coreutils.NewRedisClient()

	sessionConfig := sessions.NewSessionConfig(
		sm,
		sessions.WithPersistence(rc),
	)

	sessionConfig.CookieConfig = sessions.DebugOnlyCookieConfig

	entCfg := &entconfig.Config{
		EntityTypes: []string{},
		Modules: entconfig.Modules{
			Enabled: true,
		},
	}

	entitilements := &entitlements.StripeClient{
		Config: &entitlements.Config{
			Enabled: false,
		},
	}

	opts := []generated.Option{
		generated.Authz(*fgaClient),
		generated.TokenManager(tm),
		generated.SessionConfig(&sessionConfig),
		generated.EntConfig(entCfg),
		generated.EntitlementManager(entitilements),
	}

	s.tf = entdb.NewTestFixture()

	db, err := sql.Open("postgres", s.tf.URI)
	coreutils.RequireNoError(t, err)

	defer db.Close()

	client, err := generated.Open(dialect.Postgres, s.tf.URI, opts...)
	coreutils.RequireNoError(t, err)

	tables, err := schema.CopyTables(migrate.Tables)
	coreutils.RequireNoError(t, err)

	coreutils.RequireNoError(t, migrate.Create(ctx, client.Schema, tables, entdb.EnablePostgresOption(db)))

	galaRuntime, err := gala.NewGala(context.Background(), gala.Config{DispatchMode: gala.DispatchModeInMemory, WorkerCount: 3})
	coreutils.RequireNoError(t, err)

	_, err = gala.Register(galaRuntime, hooks.TaskRuleListeners()...)
	coreutils.RequireNoError(t, err)

	_, err = gala.Register(galaRuntime, hooks.OnboardingProgramListeners()...)
	coreutils.RequireNoError(t, err)

	_, err = gala.Register(galaRuntime, hooks.CampaignRecurringListeners()...)
	coreutils.RequireNoError(t, err)

	coreutils.RequireNoError(t, galaRuntime.Attach(
		gala.WithValue(client),
		gala.WithRestoredValue("ent_client", generated.NewContext),
	))

	client.Use(hooks.EmitGalaEventHook(galaRuntime))

	s.galaRuntime = galaRuntime

	return client
}

func (s *HookTestSuite) seedUser() *generated.User {
	t := s.T()

	// no caller exists for the new user or their personal org yet
	ctx := auth.WithInternalCrossOrgContext(context.Background())
	newUser, err := s.client.User.Create().SetEmail(gofakeit.Email()).Save(ctx)
	assert.NilError(t, err)

	// get user and their org memberships
	newUser, err = s.client.User.Query().Where(user.ID(newUser.ID)).
		WithSetting().
		WithOrgMemberships().Only(ctx)
	assert.NilError(t, err)

	return newUser
}

func (s *HookTestSuite) seedSystemAdmin() *generated.User {
	newUser := s.seedUser()

	req := fgax.TupleRequest{
		SubjectID:   newUser.ID,
		SubjectType: auth.UserSubjectType,
		ObjectID:    authmw.SystemObjectID,
		ObjectType:  authmw.SystemObject,
		Relation:    fgax.SystemAdminRelation,
	}

	// add system admin relation for user
	_, err := s.client.Authz.WriteTupleKeys(context.Background(), []fgax.TupleKey{fgax.GetTupleKey(req)}, nil)
	assert.NilError(s.T(), err)

	return newUser
}
