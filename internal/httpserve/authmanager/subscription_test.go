package authmanager_test

import (
	"context"
	"database/sql"
	"testing"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/stretchr/testify/require"

	generated "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/enttest"
	"github.com/theopenlane/core/internal/ent/generated/orgmodule"
	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/utils/testutils"
)

func newPostgresClient(t *testing.T) *generated.Client {
	tf := testutils.GetTestURI(testutils.WithImage("docker://postgres:17-alpine"))
	t.Cleanup(func() { testutils.TeardownFixture(tf) })

	db, err := sql.Open("postgres", tf.URI)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	client := enttest.Open(t, dialect.Postgres, tf.URI,
		enttest.WithMigrateOptions(enablePostgresOption(db)))

	return client
}

func enablePostgresExtensions(db *sql.DB) error {
	_, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;`)
	return err
}

func enablePostgresOption(db *sql.DB) schema.MigrateOption {
	return schema.WithHooks(func(next schema.Creator) schema.Creator {
		return schema.CreateFunc(func(ctx context.Context, tables ...*schema.Table) error {
			if err := enablePostgresExtensions(db); err != nil {
				return err
			}

			return next.Create(ctx, tables...)
		})
	})
}

func TestOrganizationHookCreatesBaseSubscription(t *testing.T) {
	client := newPostgresClient(t)
	defer client.Close()

	ctx := context.Background()
	org, err := client.Organization.Create().Save(ctx)
	require.NoError(t, err)

	sub, err := client.OrgSubscription.Query().Where(orgsubscription.OwnerID(org.ID)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, []string{"base"}, sub.Features)

	_, err = client.OrgModule.Query().Where(orgmodule.OwnerID(org.ID), orgmodule.Module("base")).Only(ctx)
	require.NoError(t, err)
}
