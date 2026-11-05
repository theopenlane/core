package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"github.com/theopenlane/utils/rout"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/models"
	apimodels "github.com/theopenlane/core/common/openapi"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/v2/internal/httpserve/handlers"
	sso "github.com/theopenlane/core/v2/pkg/ssoutils"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestLoginHandler(t *testing.T) {
	suite := suite.setupParallelTest(t)

	// Register test handler
	suite.registerTestHandler("POST", "login", suite.h.LoginHandler)

	ctx := echocontext.NewTestEchoContext().Request().Context()

	// set privacy allow in order to allow the creation of the users without
	// authentication in the tests
	ctx = auth.WithInternalOperationContext(ctx)

	// create users in the database
	validPassword := "sup3rs3cu7e!"

	validConfirmedUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      validPassword,
		confirmedUser: true,
	})

	validConfirmedUserRestrictedOrg := suite.userBuilderWithInput(ctx, &userInput{
		email:         fmt.Sprintf("meow+%s@example.com", strings.ToLower(ulids.New().String())),
		password:      validPassword,
		confirmedUser: true,
	})

	invalidConfirmedUserRestrictedOrg := suite.userBuilderWithInput(ctx, &userInput{
		email:         fmt.Sprintf("meow+%s@foobar.com", strings.ToLower(ulids.New().String())),
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

	orgSetting, err := suite.db.OrganizationSetting.Create().SetInput(
		generated.CreateOrganizationSettingInput{
			AllowedEmailDomains: []string{"examples.com"}, // intentionally misspelled to ensure owner (validConfirmedUserRestrictedOrg) can still login
		},
	).Save(ctx)
	assert.NilError(t, err)

	ssoorgSetting, err := suite.db.OrganizationSetting.Create().SetInput(
		generated.CreateOrganizationSettingInput{
			AllowedEmailDomains:           []string{"examples.com"}, // intentionally misspelled to ensure owner (validConfirmedUserRestrictedOrg) can still login
			IdentityProviderLoginEnforced: func(b bool) *bool { return &b }(true),
		},
	).Save(ctx)
	assert.NilError(t, err)

	input := generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &orgSetting.ID,
	}

	ssoOrg := generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &ssoorgSetting.ID,
	}

	ssoMember := suite.userBuilderWithInput(ctx, &userInput{
		email:         "ssomember+" + strings.ToLower(ulids.New().String()) + "@examples.com", // ensure the email is allowed by the org setting
		password:      validPassword,
		confirmedUser: true,
	})

	auditorUser := suite.userBuilderWithInput(ctx, &userInput{
		email:         "auditor+" + strings.ToLower(ulids.New().String()) + "@examples.com",
		password:      validPassword,
		confirmedUser: true,
	})

	// setup allow context with the client in the context which is required for hooks that run
	internalCtx := auth.WithInternalOperationContext(validConfirmedUserRestrictedOrg.UserCtx)
	internalCtx = ent.NewContext(internalCtx, suite.db)

	org, err := suite.db.Organization.Create().SetInput(input).Save(internalCtx)
	assert.NilError(t, err)
	createdssoOrg, err := suite.db.Organization.Create().SetInput(ssoOrg).Save(internalCtx)
	assert.NilError(t, err)

	ctxTargetOrg := auth.NewTestContextWithOrgID(validConfirmedUserRestrictedOrg.ID, createdssoOrg.ID)
	ctxTargetOrg = auth.WithInternalOperationContext(ctxTargetOrg)
	testUserCtx := ent.NewContext(ctxTargetOrg, suite.db)

	auditorOrgCtx := auth.NewTestContextWithOrgID(validConfirmedUserRestrictedOrg.ID, org.ID)
	auditorOrgCtx = auth.WithInternalOperationContext(auditorOrgCtx)
	auditorOrgCtx = ent.NewContext(auditorOrgCtx, suite.db)

	err = suite.db.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
		OrganizationID: createdssoOrg.ID,
		UserID:         ssoMember.UserInfo.ID,
		Role:           &enums.RoleMember,
	}).Exec(testUserCtx)
	assert.NilError(t, err)

	err = suite.db.UserSetting.UpdateOneID(ssoMember.UserInfo.Edges.Setting.ID).SetDefaultOrgID(createdssoOrg.ID).Exec(internalCtx)
	assert.NilError(t, err)

	err = suite.db.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
		OrganizationID: org.ID,
		UserID:         auditorUser.UserInfo.ID,
		Role:           &enums.RoleAuditor,
	}).Exec(auditorOrgCtx)
	assert.NilError(t, err)

	err = suite.db.UserSetting.UpdateOneID(auditorUser.UserInfo.Edges.Setting.ID).SetDefaultOrgID(org.ID).Exec(internalCtx)
	assert.NilError(t, err)

	// update the user settings to have the default org set that is the domain restricted org
	err = suite.db.UserSetting.UpdateOneID(validConfirmedUserRestrictedOrg.UserInfo.Edges.Setting.ID).
		SetDefaultOrgID(org.ID).Exec(internalCtx)
	assert.NilError(t, err)

	err = suite.db.UserSetting.UpdateOneID(invalidConfirmedUserRestrictedOrg.UserInfo.Edges.Setting.ID).
		SetDefaultOrgID(org.ID).Exec(internalCtx)
	assert.NilError(t, err)

	// update the user settings to have the default org set to an inactive subscription
	err = suite.db.OrgSubscription.Update().Where(orgsubscription.OwnerID(userWithInactiveDefaultOrg.OrganizationID)).
		SetActive(false).Exec(internalCtx)
	assert.NilError(t, err)

	err = suite.db.UserSetting.UpdateOneID(userWithInactiveDefaultOrg.UserInfo.Edges.Setting.ID).
		SetDefaultOrgID(userWithInactiveDefaultOrg.OrganizationID).Exec(internalCtx)
	assert.NilError(t, err)

	allModulesWithSignup := []any{
		models.CatalogBaseModule.String(),
		models.CatalogComplianceModule.String(),
		models.CatalogTrustCenterModule.String(),
	}

	testCases := []struct {
		name            string
		username        string
		password        string
		expectedOrgID   string
		expectedErr     error
		expectedStatus  int
		expectedModules []any
	}{
		{
			name:            "happy path, valid credentials",
			username:        validConfirmedUser.UserInfo.Email,
			password:        validPassword,
			expectedStatus:  http.StatusOK,
			expectedOrgID:   validConfirmedUser.OrganizationID,
			expectedModules: allModulesWithSignup,
		},
		{
			name:            "happy path, domain restricted org, but owner so domains can be mismatched",
			username:        validConfirmedUserRestrictedOrg.UserInfo.Email,
			password:        validPassword,
			expectedStatus:  http.StatusOK,
			expectedOrgID:   org.ID,
			expectedModules: allModulesWithSignup,
		},
		{
			name:            "happy path, auditor default org",
			username:        auditorUser.UserInfo.Email,
			password:        validPassword,
			expectedStatus:  http.StatusOK,
			expectedOrgID:   org.ID,
			expectedModules: allModulesWithSignup,
		},
		{
			name:            "domain restricted org, email not allowed, switch to personal org",
			username:        invalidConfirmedUserRestrictedOrg.UserInfo.Email,
			password:        validPassword,
			expectedStatus:  http.StatusOK,
			expectedOrgID:   invalidConfirmedUserRestrictedOrg.PersonalOrgID,
			expectedModules: nil, // personal orgs have no modules enabled, we just assume `base` is always available
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
			expectedErr:    handlers.ErrLoginFailed,
		},
		{
			name:           "user not found",
			username:       fmt.Sprintf("rick.sanchez%s@theopenlane.io", strings.ToLower(ulids.New().String())),
			password:       validPassword,
			expectedStatus: http.StatusBadRequest,
			expectedErr:    handlers.ErrLoginFailed,
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
			loginJSON := apimodels.LoginRequest{
				Username: tc.username,
				Password: tc.password,
			}

			body, err := json.Marshal(loginJSON)
			if err != nil {
				assert.NilError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			var out *apimodels.LoginResponse

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			assert.Check(t, is.Equal(tc.expectedStatus, recorder.Code))

			if tc.expectedStatus == http.StatusOK {
				assert.Check(t, out.Success)
				assert.Check(t, !out.TFAEnabled)
				assert.Assert(t, out.AccessToken != "")
				assert.Check(t, out.RefreshToken != "")
				assert.Check(t, out.Session != "")

				// check the claims to ensure the user is in the correct org
				token, _, err := new(jwt.Parser).ParseUnverified(out.AccessToken, jwt.MapClaims{})
				assert.NilError(t, err)

				claims, ok := token.Claims.(jwt.MapClaims)
				assert.Assert(t, ok)

				assert.Check(t, is.DeepEqual(tc.expectedOrgID, claims["org"]))

				// Module ordering is not guaranteed; assert as a set to avoid flakiness.
				if tc.expectedModules != nil {
					modules, ok := claims["modules"].([]any)
					assert.Assert(t, ok)
					assert.Check(t, is.Len(modules, len(tc.expectedModules)))

					for _, module := range tc.expectedModules {
						assert.Check(t, is.Contains(modules, module))
					}
				} else {
					modules, _ := claims["modules"].([]any)
					assert.Check(t, is.Len(modules, 0))
				}
			} else {
				assert.Check(t, is.Contains(out.Error, tc.expectedErr.Error()))
			}
		})
	}
}

func TestLoginHandlerSSOEnforced(t *testing.T) {
	suite := suite.setupParallelTest(t)

	// Register test handler
	suite.registerTestHandler("POST", "login", suite.h.LoginHandler)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = auth.WithInternalOperationContext(ctx)

	// Create an owner user and org (owner not used for login, but for org setup)
	ownerUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0wn3rP@ssw0rd",
		confirmedUser: true,
	})
	ownerCtx := auth.WithInternalOperationContext(ownerUser.UserCtx)
	ownerCtx = ent.NewContext(ownerCtx, suite.db)

	setting, err := suite.db.OrganizationSetting.Create().Save(ownerCtx)
	assert.NilError(t, err)

	org, err := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).Save(ownerCtx)
	assert.NilError(t, err)

	err = suite.db.OrganizationSetting.UpdateOneID(setting.ID).
		SetOrganizationID(org.ID).
		Exec(ownerCtx)
	assert.NilError(t, err)

	suite.enforceSSOOnSetting(ownerCtx, setting.ID)

	// Create a non-owner user
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "$uper$ecretP@ssword",
		confirmedUser: true,
	})

	ctxTargetOrg := auth.NewTestContextWithOrgID(testUser.ID, org.ID)
	ctxTargetOrg = auth.WithInternalOperationContext(ctxTargetOrg)
	testUserCtx := ent.NewContext(ctxTargetOrg, suite.db)

	err = suite.db.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
		OrganizationID: org.ID,
		UserID:         testUser.UserInfo.ID,
		Role:           &enums.RoleMember, // default to member
	}).Exec(testUserCtx)
	assert.NilError(t, err)

	err = suite.db.UserSetting.UpdateOneID(testUser.UserInfo.Edges.Setting.ID).SetDefaultOrgID(org.ID).Exec(testUserCtx)
	assert.NilError(t, err)

	// Attempt login for the non-owner user (should fail due to SSO enforcement)
	body, _ := json.Marshal(apimodels.LoginRequest{Username: testUser.UserInfo.Email, Password: "$uper$ecretP@ssword"})
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(string(body)))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	assert.Assert(t, is.Equal(http.StatusFound, rec.Code))
	assert.Check(t, is.DeepEqual(sso.SSOLogin(suite.e, org.ID), rec.Header().Get("Location")))
}

func TestLoginHandlerSSOEnforcedOwnerBypass(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerTestHandler("POST", "login", suite.h.LoginHandler)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = auth.WithInternalOperationContext(ctx)

	ownerUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "0wn3rP@ssw0rd",
		confirmedUser: true,
	})
	ownerCtx := auth.WithInternalOperationContext(ownerUser.UserCtx)
	ownerCtx = ent.NewContext(ownerCtx, suite.db)

	setting, err := suite.db.OrganizationSetting.Create().Save(ownerCtx)
	assert.NilError(t, err)

	org, err := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).Save(ownerCtx)
	assert.NilError(t, err)

	err = suite.db.OrganizationSetting.UpdateOneID(setting.ID).
		SetOrganizationID(org.ID).
		Exec(ownerCtx)
	assert.NilError(t, err)

	suite.enforceSSOOnSetting(ownerCtx, setting.ID)
	// OrganizationCreate hook automatically adds the creating user as the organization owner

	err = suite.db.UserSetting.UpdateOneID(ownerUser.UserInfo.Edges.Setting.ID).SetDefaultOrgID(org.ID).Exec(ownerCtx)
	assert.NilError(t, err)

	body, _ := json.Marshal(apimodels.LoginRequest{Username: ownerUser.UserInfo.Email, Password: "0wn3rP@ssw0rd"})
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(string(body)))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	assert.Assert(t, is.Equal(http.StatusOK, rec.Code))
	var out apimodels.LoginResponse
	assert.NilError(t, json.NewDecoder(rec.Body).Decode(&out))
	assert.Check(t, out.Success)
}

func TestLoginHandlerTFAEnforced(t *testing.T) {
	suite := suite.setupParallelTest(t)

	// Register test handler
	suite.registerTestHandler("POST", "login", suite.h.LoginHandler)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = auth.WithInternalOperationContext(ctx)

	// Create user without TFA enabled
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "T3stP@ssw0rd",
		confirmedUser: true,
		tfaEnabled:    false, // User does not have TFA enabled
	})
	testCtx := auth.WithInternalOperationContext(testUser.UserCtx)
	testCtx = ent.NewContext(testCtx, suite.db)

	// Create organization setting with TFA enforced
	setting, err := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		MultifactorAuthEnforced: func(b bool) *bool { return &b }(true),
	}).Save(testCtx)
	assert.NilError(t, err)

	org, err := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).Save(testCtx)
	assert.NilError(t, err)

	err = suite.db.OrganizationSetting.UpdateOneID(setting.ID).
		SetOrganizationID(org.ID).
		Exec(testCtx)
	assert.NilError(t, err)

	// Set user's default org to the TFA-enforced org
	err = suite.db.UserSetting.UpdateOneID(testUser.UserInfo.Edges.Setting.ID).SetDefaultOrgID(org.ID).Exec(testCtx)
	assert.NilError(t, err)

	// Attempt login
	body, _ := json.Marshal(apimodels.LoginRequest{Username: testUser.UserInfo.Email, Password: "T3stP@ssw0rd"})
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(string(body)))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	// Should succeed with an enrollment-only session so the user can set up TFA
	assert.Assert(t, is.Equal(http.StatusOK, rec.Code))
	var out apimodels.LoginResponse
	assert.NilError(t, json.NewDecoder(rec.Body).Decode(&out))
	assert.Check(t, out.Success)
	assert.Check(t, !out.TFAEnabled)
	assert.Check(t, out.TFASetupRequired)
	assert.Check(t, is.Equal("2fa enrollment required", out.Message))

	assert.Assert(t, out.AccessToken != "")
	assert.Check(t, out.RefreshToken != "")
	assert.Check(t, out.Session != "")

	token, _, err := new(jwt.Parser).ParseUnverified(out.AccessToken, jwt.MapClaims{})
	assert.NilError(t, err)

	claims, ok := token.Claims.(jwt.MapClaims)
	assert.Assert(t, ok)

	assert.Check(t, is.Equal(org.ID, claims["org"]))
	assert.Check(t, is.Equal(true, claims["tfa_enrollment_required"]))
}

func TestLoginHandlerTFAEnforcedUserHasTFA(t *testing.T) {
	suite := suite.setupParallelTest(t)

	// Register test handler
	suite.registerTestHandler("POST", "login", suite.h.LoginHandler)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = auth.WithInternalOperationContext(ctx)

	// Create user with TFA enabled
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		password:      "T3stP@ssw0rd",
		confirmedUser: true,
		tfaEnabled:    true, // User has TFA enabled
	})
	testCtx := auth.WithInternalOperationContext(testUser.UserCtx)
	testCtx = ent.NewContext(testCtx, suite.db)

	// Create organization setting with TFA enforced
	setting, err := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		MultifactorAuthEnforced: func(b bool) *bool { return &b }(true),
	}).Save(testCtx)
	assert.NilError(t, err)

	org, err := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).Save(testCtx)
	assert.NilError(t, err)

	err = suite.db.OrganizationSetting.UpdateOneID(setting.ID).
		SetOrganizationID(org.ID).
		Exec(testCtx)
	assert.NilError(t, err)

	// Set user's default org to the TFA-enforced org
	err = suite.db.UserSetting.UpdateOneID(testUser.UserInfo.Edges.Setting.ID).SetDefaultOrgID(org.ID).Exec(testCtx)
	assert.NilError(t, err)

	// Attempt login
	body, _ := json.Marshal(apimodels.LoginRequest{Username: testUser.UserInfo.Email, Password: "T3stP@ssw0rd"})
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(string(body)))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	// Should succeed without requiring TFA setup since user already has it
	assert.Assert(t, is.Equal(http.StatusOK, rec.Code))
	var out apimodels.LoginResponse
	assert.NilError(t, json.NewDecoder(rec.Body).Decode(&out))
	assert.Check(t, out.Success)
	assert.Check(t, out.TFAEnabled)        // User has TFA enabled
	assert.Check(t, !out.TFASetupRequired) // No additional TFA setup needed
}
