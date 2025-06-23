package authmanager_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/iam/auth"
	fgatest "github.com/theopenlane/iam/fgax/testutils"
	"github.com/theopenlane/iam/sessions"

	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated/privacy"

	generated "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/enttest"
	"github.com/theopenlane/core/internal/ent/generated/orgmodule"
	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/internal/entdb"
	coretest "github.com/theopenlane/core/pkg/testutils"
	"github.com/theopenlane/utils/testutils"
)

const fgaModelFile = "../../../fga/model/model.fga"

func newPostgresClient(t *testing.T) *generated.Client {
	tf := testutils.GetTestURI(testutils.WithImage("docker://postgres:17-alpine"))
	t.Cleanup(func() { testutils.TeardownFixture(tf) })

	db, err := sql.Open("postgres", tf.URI)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	fgaTF := fgatest.NewFGATestcontainer(context.Background(), fgatest.WithModelFile(fgaModelFile))
	t.Cleanup(func() { _ = fgaTF.TeardownFixture() })

	fgaClient, err := fgaTF.NewFgaClient(context.Background())
	require.NoError(t, err)

	tm, err := coretest.CreateTokenManager(15 * time.Minute)
	require.NoError(t, err)
	sm := coretest.CreateSessionManager()
	rc := coretest.NewRedisClient()

	sessionConfig := sessions.NewSessionConfig(
		sm,
		sessions.WithPersistence(rc),
	)
	sessionConfig.CookieConfig = &sessions.DebugOnlyCookieConfig

	opts := []generated.Option{
		generated.Authz(*fgaClient),
		generated.TokenManager(tm),
		generated.SessionConfig(&sessionConfig),
		generated.Emailer(&emailtemplates.Config{}),
	}

	client := enttest.Open(t, dialect.Postgres, tf.URI,
		enttest.WithMigrateOptions(entdb.EnablePostgresOption(db)),
		enttest.WithOptions(opts...))

	client.WithAuthz()

	return client
}

func TestOrganizationHookCreatesBaseSubscription(t *testing.T) {
	client := newPostgresClient(t)
	defer client.Close()

	ec := echocontext.NewTestEchoContext()
	ctx := privacy.DecisionContext(ec.Request().Context(), privacy.Allow)
	ctx = generated.NewContext(ctx, client)

	// skip owner hooks when creating the initial user and personal org
	createCtx := contextx.With(ctx, auth.OrganizationCreationContextKey{})
	user, err := client.User.Create().
		SetEmail("user@example.com").
		SetDisplayName("Test User").
		Save(createCtx)
	require.NoError(t, err)

	ctx = auth.NewTestContextWithValidUser(user.ID)
	ctx = privacy.DecisionContext(ctx, privacy.Allow)
	ctx = generated.NewContext(ctx, client)

	org, err := client.Organization.Create().
		SetName("exampleorg").
		Save(ctx)
	require.NoError(t, err)

	ctx = auth.NewTestContextWithOrgID(user.ID, org.ID)

	sub, err := client.OrgSubscription.Query().Where(orgsubscription.OwnerID(org.ID)).Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, sub.Active, true)

	module, err := client.OrgModule.Query().Where(orgmodule.OwnerID(org.ID), orgmodule.Module("base")).Only(ctx)
	require.NoError(t, err)
	assert.Contains(t, module.Module, "base")
}
