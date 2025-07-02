package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/echox"
	iamauth "github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/ulids"

	generated "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	rule "github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/sso"
	dbtest "github.com/theopenlane/core/pkg/testutils/fga"
)

func TestUnauthorizedRedirectToSSO(t *testing.T) {
	client := dbtest.NewPostgresClient(t)

	allowCtx := privacy.DecisionContext(context.Background(), privacy.Allow)
	allowCtx = generated.NewContext(allowCtx, client)

	ownerID := ulids.New().String()
	ownerCtx := iamauth.NewTestContextWithValidUser(ownerID)
	ownerCtx = privacy.DecisionContext(ownerCtx, privacy.Allow)
	ownerCtx = generated.NewContext(ownerCtx, client)

	ownerSetting, err := client.UserSetting.Create().Save(ownerCtx)
	assert.NoError(t, err)

	_, err = client.User.Create().
		SetID(ownerID).
		SetEmail("owner@example.com").
		SetDisplayName("owner").
		SetPassword("p@$$w0rd$@r3gr3@t!").
		SetSettingID(ownerSetting.ID).
		Save(ownerCtx)
	assert.NoError(t, err)

	orgSetting, err := client.OrganizationSetting.Create().
		SetBillingEmail("org1@example.com").
		Save(ownerCtx)
	assert.NoError(t, err)

	org, err := client.Organization.Create().
		SetName("org1").
		SetSettingID(orgSetting.ID).
		Save(ownerCtx)
	assert.NoError(t, err)

	_, err = orgSetting.Update().
		SetOrganizationID(org.ID).
		SetIdentityProviderLoginEnforced(true).
		Save(ownerCtx)
	assert.NoError(t, err)

	userID := ulids.New().String()
	userSetting, err := client.UserSetting.Create().Save(allowCtx)
	assert.NoError(t, err)

	user, err := client.User.Create().
		SetID(userID).
		SetEmail("user1@example.com").
		SetDisplayName("user1").
		SetPassword("p@$$w0rd$@r3gr3@t!").
		SetSettingID(userSetting.ID).
		Save(allowCtx)
	assert.NoError(t, err)

	roleMember := enums.RoleMember
	err = client.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
		OrganizationID: org.ID,
		UserID:         user.ID,
		Role:           &roleMember,
	}).Exec(ownerCtx)
	assert.NoError(t, err)

	_, err = client.UserSetting.UpdateOneID(userSetting.ID).SetDefaultOrgID(org.ID).Save(ownerCtx)
	assert.NoError(t, err)

	validator := &tokens.MockValidator{
		OnParse: func(string) (*tokens.Claims, error) {
			return &tokens.Claims{OrgID: org.ID, UserID: user.ID}, nil
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
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, sso.SSOLogin(e, org.ID), rec.Header().Get("Location"))
}

func TestUnauthorizedNoSSORedirect(t *testing.T) {
	client := dbtest.NewPostgresClient(t)

	allowCtx := privacy.DecisionContext(context.Background(), privacy.Allow)
	allowCtx = generated.NewContext(allowCtx, client)

	ownerID := ulids.New().String()
	ownerCtx := iamauth.NewTestContextWithValidUser(ownerID)
	ownerCtx = privacy.DecisionContext(ownerCtx, privacy.Allow)
	ownerCtx = generated.NewContext(ownerCtx, client)

	ownerSetting, err := client.UserSetting.Create().Save(ownerCtx)
	assert.NoError(t, err)

	_, err = client.User.Create().
		SetID(ownerID).
		SetEmail("owner2@example.com").
		SetDisplayName("owner2").
		SetPassword("p@$$w0rd$@r3gr3@t!").
		SetSettingID(ownerSetting.ID).
		Save(ownerCtx)
	assert.NoError(t, err)

	orgSetting, err := client.OrganizationSetting.Create().
		SetBillingEmail("org2@example.com").
		Save(ownerCtx)
	assert.NoError(t, err)

	org, err := client.Organization.Create().
		SetName("org2").
		SetSettingID(orgSetting.ID).
		Save(ownerCtx)
	assert.NoError(t, err)

	_, err = orgSetting.Update().
		SetOrganizationID(org.ID).
		SetIdentityProviderLoginEnforced(false).
		Save(ownerCtx)
	assert.NoError(t, err)

	userID := ulids.New().String()
	userSetting, err := client.UserSetting.Create().Save(allowCtx)
	assert.NoError(t, err)

	user, err := client.User.Create().
		SetID(userID).
		SetEmail("user2@example.com").
		SetDisplayName("user2").
		SetPassword("p@$$w0rd$@r3gr3@t!").
		SetSettingID(userSetting.ID).
		Save(allowCtx)
	assert.NoError(t, err)

	roleMember := enums.RoleMember
	err = client.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
		OrganizationID: org.ID,
		UserID:         user.ID,
		Role:           &roleMember,
	}).Exec(ownerCtx)
	assert.NoError(t, err)

	_, err = client.UserSetting.UpdateOneID(userSetting.ID).SetDefaultOrgID(org.ID).Save(ownerCtx)
	assert.NoError(t, err)

	validator := &tokens.MockValidator{
		OnParse: func(string) (*tokens.Claims, error) {
			return &tokens.Claims{OrgID: org.ID, UserID: user.ID}, nil
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
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestUnauthorizedOwnerBypass(t *testing.T) {
	client := dbtest.NewPostgresClient(t)

	ownerID := ulids.New().String()
	ctx := iamauth.NewTestContextWithValidUser(ownerID)
	ctx = rule.WithInternalContext(ctx)
	ctx = privacy.DecisionContext(ctx, privacy.Allow)
	ctx = generated.NewContext(ctx, client)

	ownerSetting, err := client.UserSetting.Create().Save(ctx)
	assert.NoError(t, err)

	owner, err := client.User.Create().
		SetID(ownerID).
		SetEmail("owner3@example.com").
		SetDisplayName("owner3").
		SetPassword("p@$$w0rd$@r3gr3@t!").
		SetSettingID(ownerSetting.ID).
		Save(ctx)
	assert.NoError(t, err)

	orgSetting, err := client.OrganizationSetting.Create().
		SetBillingEmail("org3@example.com").
		Save(ctx)
	assert.NoError(t, err)

	org, err := client.Organization.Create().
		SetName("org3").
		SetSettingID(orgSetting.ID).
		Save(ctx)
	assert.NoError(t, err)

	_, err = orgSetting.Update().
		SetOrganizationID(org.ID).
		SetIdentityProviderLoginEnforced(true).
		Save(ctx)
	assert.NoError(t, err)

	_, err = client.UserSetting.UpdateOneID(ownerSetting.ID).SetDefaultOrgID(org.ID).Save(ctx)
	assert.NoError(t, err)

	validator := &tokens.MockValidator{
		OnParse: func(string) (*tokens.Claims, error) {
			return &tokens.Claims{OrgID: org.ID, UserID: owner.ID}, nil
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
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
