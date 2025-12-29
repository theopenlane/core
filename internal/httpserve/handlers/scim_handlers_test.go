package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
	"github.com/theopenlane/core/internal/httpserve/route"
	"github.com/theopenlane/utils/ulids"
)

const (
	scimUserSchema  = "urn:ietf:params:scim:schemas:core:2.0:User"
	scimGroupSchema = "urn:ietf:params:scim:schemas:core:2.0:Group"
	scimPatchSchema = "urn:ietf:params:scim:api:messages:2.0:PatchOp"
)

func (suite *HandlerTestSuite) TestSCIMUserHandlerCreate() {
	// Create SCIM-specific test user with example.com domain
	// This ensures allowed_email_domains will include example.com
	ctx := context.Background()
	scimTestUser := suite.userBuilderWithInput(ctx, &userInput{
		email: ulids.New().String() + "@example.com",
	})

	// Enable SSO enforcement for SCIM (required for SCIM operations)
	ctx = privacy.DecisionContext(scimTestUser.UserCtx, privacy.Allow)
	org, err := suite.db.Organization.Get(ctx, scimTestUser.OrganizationID)
	suite.Require().NoError(err)
	setting, err := org.Setting(ctx)
	suite.Require().NoError(err)
	err = suite.db.OrganizationSetting.UpdateOneID(setting.ID).SetIdentityProviderLoginEnforced(true).Exec(ctx)
	suite.Require().NoError(err)

	// Register SCIM routes
	suite.router.Handler = suite.h
	err = route.RegisterRoutes(suite.router)
	suite.Require().NoError(err)

	testCases := []struct {
		name        string
		given       string
		family      string
		displayName string
		active      bool
	}{
		{
			name:   "active user with inferred display",
			given:  "Ada",
			family: "Lovelace",
			active: true,
		},
		{
			name:        "inactive user with explicit display",
			given:       "Grace",
			family:      "Hopper",
			displayName: "Captain Hopper",
			active:      false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			email := fmt.Sprintf("scim-user-%s@example.com", strings.ToLower(ulids.New().String()))

			body := map[string]any{
				"schemas":  []string{scimUserSchema},
				"userName": email,
				"name": map[string]any{
					"givenName":  tc.given,
					"familyName": tc.family,
				},
			}

			if tc.displayName != "" {
				body["displayName"] = tc.displayName
			}

			if !tc.active {
				body["active"] = tc.active
			}

			payload, err := json.Marshal(body)
			suite.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/scim/v2/Users", bytes.NewReader(payload))
			req.Header.Set("Content-Type", "application/scim+json")
			req.Header.Set("Accept", "application/scim+json")

			rec := httptest.NewRecorder()
			suite.e.ServeHTTP(rec, req.WithContext(scimTestUser.UserCtx))

			suite.T().Logf("Test case: %s, Status: %d, Body: %s", tc.name, rec.Code, rec.Body.String())
			suite.Require().Equal(http.StatusCreated, rec.Code, rec.Body.String())

			var resp map[string]any
			suite.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &resp))

			userID := suite.getStringField(resp, "id")

			// Verify the response contains correct values
			active, ok := resp["active"].(bool)
			suite.Require().True(ok, "active should be a boolean")
			suite.Equal(tc.active, active, "active flag should match")

			nameMap, ok := resp["name"].(map[string]any)
			suite.Require().True(ok, "name should be a map")
			suite.Equal(tc.given, nameMap["givenName"], "givenName should match")
			suite.Equal(tc.family, nameMap["familyName"], "familyName should match")

			expectedDisplay := tc.displayName
			if expectedDisplay == "" {
				expectedDisplay = strings.TrimSpace(tc.given + " " + tc.family)
				if expectedDisplay == "" {
					expectedDisplay = strings.ToLower(email)
				}
			}
			suite.Equal(expectedDisplay, resp["displayName"], "displayName should match")

			// Only verify database state for active users (soft-deleted users are filtered by default)
			if tc.active {
				verifyCtx := privacy.DecisionContext(context.Background(), privacy.Allow)
				createdUser, err := suite.db.User.Query().Where(user.ID(userID)).Only(verifyCtx)
				suite.Require().NoError(err)

				expectedEmail := strings.ToLower(email)
				suite.Equal(expectedEmail, createdUser.Email)
				suite.Equal(tc.given, createdUser.FirstName)
				suite.Equal(tc.family, createdUser.LastName)
				suite.Equal(expectedDisplay, createdUser.DisplayName)

				suite.Require().NotNil(createdUser.ScimUsername)
				suite.True(strings.EqualFold(expectedEmail, *createdUser.ScimUsername))
				suite.Equal(tc.active, createdUser.ScimActive)
				suite.Equal(enums.AuthProviderOIDC, createdUser.LastLoginProvider)
				suite.Equal(enums.AuthProviderOIDC, createdUser.AuthProvider)
				suite.Equal(enums.RoleUser, createdUser.Role)

				userSettings, err := suite.db.UserSetting.Query().Where(usersetting.UserID(userID)).Only(verifyCtx)
				suite.Require().NoError(err)
				suite.True(userSettings.EmailConfirmed)

				orgMembership, err := suite.db.OrgMembership.Query().
					Where(
						orgmembership.UserID(userID),
						orgmembership.OrganizationID(scimTestUser.OrganizationID),
					).
					Only(verifyCtx)
				suite.Require().NoError(err)
				suite.Equal(enums.RoleMember, orgMembership.Role)
			}
		})
	}
}

func (suite *HandlerTestSuite) getStringField(data map[string]any, key string) string {
	value, ok := data[key]
	suite.Require().True(ok, "missing key %s", key)

	str, ok := value.(string)
	suite.Require().True(ok, "key %s is not a string", key)

	return str
}

func (suite *HandlerTestSuite) TestSCIMUserHandlerPatchActiveToggle() {
	// Create SCIM-specific test user with example.com domain
	ctx := context.Background()
	scimTestUser := suite.userBuilderWithInput(ctx, &userInput{
		email: ulids.New().String() + "@example.com",
	})

	// Enable SSO enforcement for SCIM (required for SCIM operations)
	ctx = privacy.DecisionContext(scimTestUser.UserCtx, privacy.Allow)
	org, err := suite.db.Organization.Get(ctx, scimTestUser.OrganizationID)
	suite.Require().NoError(err)
	setting, err := org.Setting(ctx)
	suite.Require().NoError(err)
	err = suite.db.OrganizationSetting.UpdateOneID(setting.ID).SetIdentityProviderLoginEnforced(true).Exec(ctx)
	suite.Require().NoError(err)

	// Register SCIM routes
	suite.router.Handler = suite.h
	err = route.RegisterRoutes(suite.router)
	suite.Require().NoError(err)

	email := fmt.Sprintf("scim-user-%s@example.com", strings.ToLower(ulids.New().String()))
	createBody := map[string]any{
		"schemas":  []string{scimUserSchema},
		"userName": email,
		"name": map[string]any{
			"givenName":  "Toggle",
			"familyName": "Target",
		},
	}

	payload, err := json.Marshal(createBody)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/scim/v2/Users", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/scim+json")
	req.Header.Set("Accept", "application/scim+json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(scimTestUser.UserCtx))
	suite.Require().Equal(http.StatusCreated, rec.Code, rec.Body.String())

	var created map[string]any
	suite.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &created))
	userID := suite.getStringField(created, "id")

	patchBody := map[string]any{
		"schemas": []string{scimPatchSchema},
		"Operations": []map[string]any{
			{
				"op":    "replace",
				"value": map[string]any{"active": false},
			},
		},
	}

	payload, err = json.Marshal(patchBody)
	suite.Require().NoError(err)

	req = httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/scim/v2/Users/%s", userID), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/scim+json")
	req.Header.Set("Accept", "application/scim+json")

	rec = httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(scimTestUser.UserCtx))
	suite.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())

	// Verify user was deactivated by checking the SCIM response
	var deactivatedResp map[string]any
	suite.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &deactivatedResp))
	deactivatedActive, ok := deactivatedResp["active"].(bool)
	suite.Require().True(ok, "active should be a boolean")
	suite.False(deactivatedActive, "user should be inactive after patch")

	patchBody["Operations"].([]map[string]any)[0]["value"] = map[string]any{"active": true}

	payload, err = json.Marshal(patchBody)
	suite.Require().NoError(err)

	req = httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/scim/v2/Users/%s", userID), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/scim+json")
	req.Header.Set("Accept", "application/scim+json")

	rec = httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(scimTestUser.UserCtx))
	suite.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())

	// Verify user was reactivated by checking the SCIM response
	var reactivatedResp map[string]any
	suite.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &reactivatedResp))
	reactivatedActive, ok := reactivatedResp["active"].(bool)
	suite.Require().True(ok, "active should be a boolean")
	suite.True(reactivatedActive, "user should be active after reactivation patch")
}

func (suite *HandlerTestSuite) TestSCIMGroupHandlerCreateDeduplicatesMembers() {
	// Create SCIM-specific test user with example.com domain
	ctx := context.Background()
	scimTestUser := suite.userBuilderWithInput(ctx, &userInput{
		email: ulids.New().String() + "@example.com",
	})

	// Enable SSO enforcement for SCIM (required for SCIM operations)
	ctx = privacy.DecisionContext(scimTestUser.UserCtx, privacy.Allow)
	org, err := suite.db.Organization.Get(ctx, scimTestUser.OrganizationID)
	suite.Require().NoError(err)
	setting, err := org.Setting(ctx)
	suite.Require().NoError(err)
	err = suite.db.OrganizationSetting.UpdateOneID(setting.ID).SetIdentityProviderLoginEnforced(true).Exec(ctx)
	suite.Require().NoError(err)

	// Register SCIM routes
	suite.router.Handler = suite.h
	err = route.RegisterRoutes(suite.router)
	suite.Require().NoError(err)

	ctx = privacy.DecisionContext(scimTestUser.UserCtx, privacy.Allow)

	member := suite.userBuilderWithInput(ctx, &userInput{
		email: ulids.New().String() + "@example.com",
	})

	_, err = suite.db.OrgMembership.Create().
		SetOrganizationID(scimTestUser.OrganizationID).
		SetUserID(member.ID).
		Save(ctx)
	suite.Require().NoError(err)

	displayName := "SCIM Engineering Team"
	body := map[string]any{
		"schemas":     []string{scimGroupSchema},
		"displayName": displayName,
		"members": []map[string]any{
			{"value": member.ID},
			{"value": member.ID},
		},
	}

	payload, err := json.Marshal(body)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/scim/v2/Groups", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/scim+json")
	req.Header.Set("Accept", "application/scim+json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(scimTestUser.UserCtx))
	suite.T().Logf("Group create test - Status: %d, Body: %s", rec.Code, rec.Body.String())
	suite.Require().Equal(http.StatusCreated, rec.Code, rec.Body.String())

	var resp map[string]any
	suite.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &resp))

	groupID := suite.getStringField(resp, "id")
	suite.T().Logf("Created group ID: %s", groupID)

	membersValue, ok := resp["members"].([]any)
	suite.Require().True(ok, "members should be an array")
	suite.Len(membersValue, 1, "should have exactly 1 member after deduplication")

	memberMap, ok := membersValue[0].(map[string]any)
	suite.Require().True(ok, "member should be a map")
	suite.Equal(member.ID, suite.getStringField(memberMap, "value"), "member ID should match")

	suite.Equal(displayName, resp["displayName"], "displayName should match")
}
