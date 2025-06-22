package authmanager_test

import (
	"context"
	"database/sql"
	"testing"

	"entgo.io/ent/dialect"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/auth"
	fgatest "github.com/theopenlane/iam/fgax/testutils"

	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated/privacy"

	generated "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/enttest"
	"github.com/theopenlane/core/internal/ent/generated/orgmodule"
	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/internal/entdb"
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

	client := enttest.Open(t, dialect.Postgres, tf.URI,
		enttest.WithMigrateOptions(entdb.EnablePostgresOption(db)),
		enttest.WithOptions(generated.Authz(*fgaClient)))

	client.WithAuthz()

	return client
}

func TestOrganizationHookCreatesBaseSubscription(t *testing.T) {
	client := newPostgresClient(t)
	defer client.Close()

	ctx := privacy.DecisionContext(context.Background(), privacy.Allow)

	// skip owner hooks when creating the initial user and personal org
	createCtx := contextx.With(ctx, auth.OrganizationCreationContextKey{})
	user, err := client.User.Create().
		SetEmail("user@example.com").
		SetDisplayName("Test User").
		Save(createCtx)
	require.NoError(t, err)

	ctx = auth.WithAuthenticatedUser(ctx, &auth.AuthenticatedUser{
		SubjectID:    user.ID,
		SubjectEmail: user.Email,
	})

	org, err := client.Organization.Create().SetName("MITB").Save(ctx)
	require.NoError(t, err)

	sub, err := client.OrgSubscription.Query().Where(orgsubscription.OwnerID(org.ID)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, []string{"base"}, sub.Features)

	_, err = client.OrgModule.Query().Where(orgmodule.OwnerID(org.ID), orgmodule.Module("base")).Only(ctx)
	require.NoError(t, err)
}
