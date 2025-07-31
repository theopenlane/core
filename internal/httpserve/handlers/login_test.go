package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/utils/rout"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/internal/ent/generated"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	sso "github.com/theopenlane/core/pkg/ssoutils"
)

func (suite *HandlerTestSuite) TestLoginHandler() {
	t := suite.T()

	// Register test handler with OpenAPI context
	suite.registerTestHandler("POST", "login", suite.createImpersonationOperation("LoginHandler", "Test login"), suite.h.LoginHandler)

	ctx := echocontext.NewTestEchoContext().Request().Context()

	// set privacy allow in order to allow the creation of the users without
	// authentication in the tests
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// create users in the database
	validPassword := "sup3rs3cu7e!"

	validConfirmedUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      validPassword,
		confirmedUser: true,
	})

	validConfirmedUserRestrictedOrg := suite.userBuilderWithInput(ctx, &userInput{
		email:         "meow@example.com",
		password:      validPassword,
		confirmedUser: true,
	})

	invalidConfirmedUserRestrictedOrg := suite.userBuilderWithInput(ctx, &userInput{
		email:         "meow@foobar.com",
		password:      validPassword,
		confirmedUser: true,
	})

	validUnconfirmedUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      validPassword,
		confirmedUser: false,
	})

	userWithInactiveDefaultOrg := suite.userBuilderWithInput(ctx, &userInput{
		password:      validPassword,
		confirmedUser: true,
	})

	orgSetting := suite.db.OrganizationSetting.Create().SetInput(
		generated.CreateOrganizationSettingInput{
			AllowedEmailDomains: []string{"examples.com"}, // intentionally misspelled to ensure owner (validConfirmedUserRestrictedOrg) can still login
		},
	).SaveX(ctx)

	ssoorgSetting := suite.db.OrganizationSetting.Create().SetInput(
		generated.CreateOrganizationSettingInput{
			AllowedEmailDomains:           []string{"examples.com"}, // intentionally misspelled to ensure owner (validConfirmedUserRestrictedOrg) can still login
			IdentityProviderLoginEnforced: func(b bool) *bool { return &b }(true),
		},
	).SaveX(ctx)

	input := generated.CreateOrganizationInput{
		Name:      gofakeit.AdjectiveDescriptive() + " " + gofakeit.Noun() + time.Now().Format("20060102150405"),
		SettingID: &orgSetting.ID,
	}

	ssoOrg := generated.CreateOrganizationInput{
		Name:      gofakeit.AdjectiveDescriptive() + " " + gofakeit.Noun() + time.Now().Format("20060102150405"),
		SettingID: &ssoorgSetting.ID,
	}

	ssoMember := suite.userBuilderWithInput(ctx, &userInput{
		email:         gofakeit.Username() + "@examples.com", // ensure the email is allowed by the org setting
		password:      validPassword,
		confirmedUser: true,
	})
	// setup allow context with the client in the context which is required for hooks that run
	allowCtx := privacy.DecisionContext(validConfirmedUserRestrictedOrg.UserCtx, privacy.Allow)
	allowCtx = ent.NewContext(allowCtx, suite.db)

	org := suite.db.Organization.Create().SetInput(input).SaveX(allowCtx)
	createdssoOrg := suite.db.Organization.Create().SetInput(ssoOrg).SaveX(allowCtx)
	suite.db.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
		OrganizationID: createdssoOrg.ID,
		UserID:         ssoMember.UserInfo.ID,
		Role:           &enums.RoleMember,
	}).ExecX(allowCtx)

	suite.db.UserSetting.UpdateOneID(ssoMember.UserInfo.Edges.Setting.ID).SetDefaultOrgID(createdssoOrg.ID).ExecX(allowCtx)

	// update the user settings to have the default org set that is the domain restricted org
	suite.db.UserSetting.UpdateOneID(validConfirmedUserRestrictedOrg.UserInfo.Edges.Setting.ID).
		SetDefaultOrgID(org.ID).ExecX(allowCtx)

	suite.db.UserSetting.UpdateOneID(invalidConfirmedUserRestrictedOrg.UserInfo.Edges.Setting.ID).
		SetDefaultOrgID(org.ID).ExecX(allowCtx)

	// update the user settings to have the default org set to an inactive subscription
	suite.db.OrgSubscription.Update().Where(orgsubscription.OwnerID(userWithInactiveDefaultOrg.OrganizationID)).
		SetActive(false).ExecX(allowCtx)

	suite.db.UserSetting.UpdateOneID(userWithInactiveDefaultOrg.UserInfo.Edges.Setting.ID).
		SetDefaultOrgID(userWithInactiveDefaultOrg.OrganizationID).ExecX(allowCtx)

	// setup mock entitlements client
	entitlements, err := suite.mockStripeClient()
	require.NoError(t, err)

	suite.h.DBClient.EntitlementManager = entitlements
	suite.h.Entitlements = entitlements

	testCases := []struct {
		name           string
		username       string
		password       string
		expectedOrgID  string
		expectedErr    error
		expectedStatus int
	}{
		{
			name:           "happy path, valid credentials",
			username:       validConfirmedUser.UserInfo.Email,
			password:       validPassword,
			expectedStatus: http.StatusOK,
			expectedOrgID:  validConfirmedUser.OrganizationID,
		},
		{
			name:           "happy path, domain restricted org, but owner so domains can be mismatched",
			username:       validConfirmedUserRestrictedOrg.UserInfo.Email,
			password:       validPassword,
			expectedStatus: http.StatusOK,
			expectedOrgID:  org.ID,
		},
		{
			name:           "domain restricted org, email not allowed, switch to personal org",
			username:       invalidConfirmedUserRestrictedOrg.UserInfo.Email,
			password:       validPassword,
			expectedStatus: http.StatusOK,
			expectedOrgID:  invalidConfirmedUserRestrictedOrg.PersonalOrgID,
		},
		{
			name:           "inactive org, switch to personal org",
			username:       userWithInactiveDefaultOrg.UserInfo.Email,
			password:       validPassword,
			expectedStatus: http.StatusOK,
			expectedOrgID:  userWithInactiveDefaultOrg.PersonalOrgID,
		},
		{
			name:           "email unverified",
			username:       validUnconfirmedUser.UserInfo.Email,
			password:       validPassword,
			expectedStatus: http.StatusBadRequest,
			expectedErr:    auth.ErrUnverifiedUser,
		},
		{
			name:           "invalid password",
			username:       validConfirmedUser.UserInfo.Email,
			password:       "thisisnottherightone",
			expectedStatus: http.StatusBadRequest,
			expectedErr:    rout.ErrInvalidCredentials,
		},
		{
			name:           "user not found",
			username:       "rick.sanchez@theopenlane.io",
			password:       validPassword,
			expectedStatus: http.StatusBadRequest,
			expectedErr:    auth.ErrNoAuthUser,
		},
		{
			name:           "empty username",
			username:       "",
			password:       validPassword,
			expectedStatus: http.StatusBadRequest,
			expectedErr:    rout.NewMissingRequiredFieldError("username"),
		},
		{
			name:           "empty password",
			username:       validConfirmedUser.UserInfo.Email,
			password:       "",
			expectedStatus: http.StatusBadRequest,
			expectedErr:    rout.NewMissingRequiredFieldError("password"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			loginJSON := models.LoginRequest{
				Username: tc.username,
				Password: tc.password,
			}

			body, err := json.Marshal(loginJSON)
			if err != nil {
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			var out *models.LoginReply

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.expectedStatus == http.StatusOK {
				assert.True(t, out.Success)
				assert.True(t, out.TFAEnabled) // we set the user to have TFA enabled in the tests
				require.NotNil(t, out.AccessToken)

				// check the claims to ensure the user is in the correct org
				token, _, err := new(jwt.Parser).ParseUnverified(out.AccessToken, jwt.MapClaims{})
				require.NoError(t, err)

				claims, ok := token.Claims.(jwt.MapClaims)
				require.True(t, ok)

				assert.Equal(t, tc.expectedOrgID, claims["org"])
			} else {
				assert.Contains(t, out.Error, tc.expectedErr.Error())
			}
		})
	}
}

func (suite *HandlerTestSuite) TestLoginHandlerSSOEnforced() {
	t := suite.T()

	// Register test handler with OpenAPI context
	suite.registerTestHandler("POST", "login", suite.createImpersonationOperation("LoginHandler", "Test login"), suite.h.LoginHandler)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// Create an owner user and org (owner not used for login, but for org setup)
	ownerUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0wn3rP@ssw0rd",
		confirmedUser: true,
	})
	ownerCtx := privacy.DecisionContext(ownerUser.UserCtx, privacy.Allow)
	ownerCtx = ent.NewContext(ownerCtx, suite.db)

	setting := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		IdentityProviderLoginEnforced: func(b bool) *bool { return &b }(true),
	}).SaveX(ownerCtx)

	org := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      "ssoorg" + time.Now().Format("20060102150405"),
		SettingID: &setting.ID,
	}).SaveX(ownerCtx)

	suite.db.OrganizationSetting.UpdateOneID(setting.ID).
		SetOrganizationID(org.ID).
		ExecX(ownerCtx)

	// Create a non-owner user
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "$uper$ecretP@ssword",
		confirmedUser: true,
	})
	testUserCtx := privacy.DecisionContext(testUser.UserCtx, privacy.Allow)
	testUserCtx = ent.NewContext(testUserCtx, suite.db)

	suite.db.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
		OrganizationID: org.ID,
		UserID:         testUser.UserInfo.ID,
		Role:           &enums.RoleMember, // default to member
	}).ExecX(testUserCtx)

	suite.db.UserSetting.UpdateOneID(testUser.UserInfo.Edges.Setting.ID).SetDefaultOrgID(org.ID).ExecX(testUserCtx)

	// Attempt login for the non-owner user (should fail due to SSO enforcement)
	body, _ := json.Marshal(models.LoginRequest{Username: testUser.UserInfo.Email, Password: "$uper$ecretP@ssword"})
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(string(body)))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, sso.SSOLogin(suite.e, org.ID), rec.Header().Get("Location"))
}

func (suite *HandlerTestSuite) TestLoginHandlerSSOEnforcedOwnerBypass() {
	t := suite.T()

	suite.registerTestHandler("POST", "login", suite.createImpersonationOperation("LoginHandler", "Test login"), suite.h.LoginHandler)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	ownerUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0wn3rP@ssw0rd",
		confirmedUser: true,
	})
	ownerCtx := privacy.DecisionContext(ownerUser.UserCtx, privacy.Allow)
	ownerCtx = ent.NewContext(ownerCtx, suite.db)

	setting := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		IdentityProviderLoginEnforced: func(b bool) *bool { return &b }(true),
	}).SaveX(ownerCtx)

	org := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).SaveX(ownerCtx)

	suite.db.OrganizationSetting.UpdateOneID(setting.ID).
		SetOrganizationID(org.ID).
		ExecX(ownerCtx)
	// OrganizationCreate hook automatically adds the creating user as the organization owner

	suite.db.UserSetting.UpdateOneID(ownerUser.UserInfo.Edges.Setting.ID).SetDefaultOrgID(org.ID).ExecX(ownerCtx)

	body, _ := json.Marshal(models.LoginRequest{Username: ownerUser.UserInfo.Email, Password: "0wn3rP@ssw0rd"})
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(string(body)))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var out models.LoginReply
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&out))
	assert.True(t, out.Success)
}
