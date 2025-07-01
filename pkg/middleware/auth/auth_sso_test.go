package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/echox"
	iamauth "github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"

	"entgo.io/ent/dialect"
	"github.com/theopenlane/core/internal/ent/generated/enttest"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/entdb"
	"github.com/theopenlane/utils/testutils"
)

func TestUnauthorizedRedirectToSSO(t *testing.T) {
	tf := entdb.NewTestFixture()
	defer testutils.TeardownFixture(tf)

	db, err := sql.Open("postgres", tf.URI)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	client := enttest.Open(t, dialect.Postgres, tf.URI,
		enttest.WithMigrateOptions(entdb.EnablePostgresOption(db)))
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	ctx := privacy.DecisionContext(context.Background(), privacy.Allow)
	_, err = client.OrganizationSetting.Create().
		SetOrganizationID("org1").
		SetIdentityProviderLoginEnforced(true).
		Save(ctx)
	require.NoError(t, err)

	validator := &tokens.MockValidator{
		OnParse: func(string) (*tokens.Claims, error) {
			return &tokens.Claims{OrgID: "org1"}, nil
		},
	}

	conf := NewAuthOptions(
		WithDBClient(client),
		WithValidator(validator),
	)

	e := echox.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(iamauth.Authorization, "Bearer token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = unauthorized(c, errors.New("invalid"), &conf, validator)
	require.NoError(t, err)
	require.Equal(t, http.StatusFound, rec.Code)
	require.Equal(t, "/v1/sso/login?organization_id=org1", rec.Header().Get("Location"))
}

func TestUnauthorizedNoSSORedirect(t *testing.T) {
	tf := entdb.NewTestFixture()
	defer testutils.TeardownFixture(tf)

	db, err := sql.Open("postgres", tf.URI)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	client := enttest.Open(t, dialect.Postgres, tf.URI,
		enttest.WithMigrateOptions(entdb.EnablePostgresOption(db)))
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	ctx := privacy.DecisionContext(context.Background(), privacy.Allow)
	_, err = client.OrganizationSetting.Create().
		SetOrganizationID("org2").
		SetIdentityProviderLoginEnforced(false).
		Save(ctx)
	require.NoError(t, err)

	validator := &tokens.MockValidator{
		OnParse: func(string) (*tokens.Claims, error) {
			return &tokens.Claims{OrgID: "org2"}, nil
		},
	}

	conf := NewAuthOptions(
		WithDBClient(client),
		WithValidator(validator),
	)

	e := echox.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(iamauth.Authorization, "Bearer token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = unauthorized(c, errors.New("invalid"), &conf, validator)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
