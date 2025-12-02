package hooks_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/ent/entconfig"
	"github.com/theopenlane/ent/entdb"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/enttest"
	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/ent/generated/user"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	fgatest "github.com/theopenlane/iam/fgax/testutils"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/shared/entitlements"
	authmw "github.com/theopenlane/shared/middleware/auth"
	coreutils "github.com/theopenlane/shared/testutils"
	"github.com/theopenlane/utils/testutils"
)

const (
	fgaModelFile = "../../../fga/model/model.fga"
)

// TestHookSuite runs all the tests in the TestHookSuite
func TestHookTestSuite(t *testing.T) {
	suite.Run(t, new(HookTestSuite))
}

// HookTestSuite handles the setup and teardown between tests
type HookTestSuite struct {
	suite.Suite
	client *generated.Client
	tf     *testutils.TestFixture
	ofgaTF *fgatest.OpenFGATestFixture
}

// SetupSuite runs before the test suite
func (suite *HookTestSuite) SetupSuite() {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	if testing.Verbose() {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}

	suite.client = suite.setupClient()
}

// TearDownSuite runs after the test suite
func (suite *HookTestSuite) TearDownSuite() {
	t := suite.T()

	// close the database connection
	err := suite.client.Close()
	require.NoError(t, err)

	// close the database container
	testutils.TeardownFixture(suite.tf)
}

// setupClient sets up the client for the test suite
// using enttest
func (suite *HookTestSuite) setupClient() *generated.Client {
	t := suite.T()

	suite.ofgaTF = fgatest.NewFGATestcontainer(context.Background(), fgatest.WithModelFile(fgaModelFile))
	ctx := context.Background()

	fgaClient, err := suite.ofgaTF.NewFgaClient(ctx)
	require.NoError(t, err)

	tm, err := coreutils.CreateTokenManager(-15 * time.Minute) //nolint:mnd
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
		generated.Emailer(&emailtemplates.Config{}),
		generated.EntConfig(entCfg),
		generated.EntitlementManager(entitilements),
	}

	suite.tf = entdb.NewTestFixture()

	db, err := sql.Open("postgres", suite.tf.URI)
	require.NoError(t, err)

	defer db.Close()

	client := enttest.Open(t, dialect.Postgres, suite.tf.URI,
		enttest.WithMigrateOptions(entdb.EnablePostgresOption(db)),
		enttest.WithOptions(opts...))

	return client
}

func (suite *HookTestSuite) seedUser() *generated.User {
	t := suite.T()

	ctx := privacy.DecisionContext(context.Background(), privacy.Allow)
	newUser, err := suite.client.User.Create().SetEmail(gofakeit.Email()).Save(ctx)
	require.NoError(t, err)

	// get user and their org memberships
	newUser, err = suite.client.User.Query().Where(user.ID(newUser.ID)).
		WithSetting().
		WithOrgMemberships().Only(ctx)
	require.NoError(t, err)

	return newUser
}

func (suite *HookTestSuite) seedSystemAdmin() *generated.User {
	newUser := suite.seedUser()

	req := fgax.TupleRequest{
		SubjectID:   newUser.ID,
		SubjectType: auth.UserSubjectType,
		ObjectID:    authmw.SystemObjectID,
		ObjectType:  authmw.SystemObject,
		Relation:    fgax.SystemAdminRelation,
	}

	// add system admin relation for user
	_, err := suite.client.Authz.WriteTupleKeys(context.Background(), []fgax.TupleKey{fgax.GetTupleKey(req)}, nil)
	require.NoError(suite.T(), err)

	return newUser
}
