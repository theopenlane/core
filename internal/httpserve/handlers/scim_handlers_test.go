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
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/httpserve/route"
	definitionscim "github.com/theopenlane/core/internal/integrations/definitions/scim"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/utils/ulids"
)

const (
	scimUserSchema  = "urn:ietf:params:scim:schemas:core:2.0:User"
	scimGroupSchema = "urn:ietf:params:scim:schemas:core:2.0:Group"
	scimPatchSchema = "urn:ietf:params:scim:api:messages:2.0:PatchOp"

	// scimProviderID is the provider identifier used when creating SCIM webhook rows
	scimProviderID = "scim"
)

// registerSCIMRoutesWithAuth registers SCIM routes and returns a cleanup function
// that restores the original echo handler. The SCIM route is public (unauthenticated)
// so no auth middleware is wired for the SCIM path itself; Bearer secret validation
// happens inside the SCIM middleware registered by Stream 3.
func (suite *HandlerTestSuite) registerSCIMRoutesWithAuth() func() {
	suite.T().Helper()

	restore := suite.withDefinitionRuntime(suite.T(), []registry.Builder{definitionscim.Builder()})
	suite.router.Handler = suite.h

	err := route.RegisterRoutes(suite.router)
	suite.Require().NoError(err)

	return restore
}

// createSCIMIntegration provisions an Integration row and an IntegrationWebhook
// (the "scim.auth" row) that supplies the SCIM Bearer secret. Returns the
// integration ID (for DB assertions), the webhook ID used as the endpoint path
// segment, and the auto-generated secret token.
func (suite *HandlerTestSuite) createSCIMIntegration(ctx context.Context, orgID, name string) (integrationID, endpointID, secret string) {
	suite.T().Helper()

	scimInteg, err := suite.db.Integration.Create().
		SetOwnerID(orgID).
		SetName(name).
		SetKind(scimProviderID).
		SetDefinitionID(definitionscim.DefinitionID.ID()).
		SetStatus(enums.IntegrationStatusConnected).
		Save(ctx)
	suite.Require().NoError(err)

	// Provision the SCIM webhook/auth row; SecretToken is populated by the
	// schema's DefaultFunc (prefix "tola_wsec").
	eid := "tolwh_test_" + scimInteg.ID
	webhook, err := suite.db.IntegrationWebhook.Create().
		SetOwnerID(orgID).
		SetIntegrationID(scimInteg.ID).
		SetProvider(scimProviderID).
		SetName("scim.auth").
		SetEndpointID(eid).
		SetEndpointURL("/v1/integrations/scim/" + eid + "/v2").
		Save(ctx)
	suite.Require().NoError(err)

	suite.Require().NotNil(webhook.EndpointID)

	return scimInteg.ID, *webhook.EndpointID, webhook.SecretToken
}

// newSCIMRequest builds an *http.Request targeting the new SCIM route with
// Bearer secret authentication.
func (suite *HandlerTestSuite) newSCIMRequest(method, path, secret string, payload []byte) *http.Request {
	suite.T().Helper()

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+secret)
	req.Header.Set("Content-Type", "application/scim+json")
	req.Header.Set("Accept", "application/scim+json")

	return req
}

func (suite *HandlerTestSuite) TestSCIMUserHandlerCreate() {
	ctx := context.Background()
	scimTestUser := suite.userBuilderWithInput(ctx, &userInput{
		email: ulids.New().String() + "@example.com",
	})

	// Enable SSO enforcement for SCIM (required for SCIM operations)
	allowCtx := privacy.DecisionContext(scimTestUser.UserCtx, privacy.Allow)
	org, err := suite.db.Organization.Get(allowCtx, scimTestUser.OrganizationID)
	suite.Require().NoError(err)
	setting, err := org.Setting(allowCtx)
	suite.Require().NoError(err)
	err = suite.db.OrganizationSetting.UpdateOneID(setting.ID).SetIdentityProviderLoginEnforced(true).Exec(allowCtx)
	suite.Require().NoError(err)

	restore := suite.registerSCIMRoutesWithAuth()
	defer restore()

	_, endpointID, secret := suite.createSCIMIntegration(allowCtx, scimTestUser.OrganizationID, "SCIM Directory")

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

			req := suite.newSCIMRequest(http.MethodPost, fmt.Sprintf("/v1/integrations/scim/%s/v2/Users", endpointID), secret, payload)

			rec := httptest.NewRecorder()
			suite.e.ServeHTTP(rec, req)

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

			// Verify DirectoryAccount was created with correct attributes
			da, err := suite.db.DirectoryAccount.Get(allowCtx, userID)
			suite.Require().NoError(err)

			expectedEmail := strings.ToLower(email)
			suite.Require().NotNil(da.CanonicalEmail)
			suite.Equal(expectedEmail, *da.CanonicalEmail)
			suite.Equal(expectedDisplay, da.DisplayName)

			suite.Require().NotNil(da.GivenName)
			suite.Equal(tc.given, *da.GivenName)
			suite.Require().NotNil(da.FamilyName)
			suite.Equal(tc.family, *da.FamilyName)

			if tc.active {
				suite.Equal(enums.DirectoryAccountStatusActive, da.Status)
			} else {
				suite.Equal(enums.DirectoryAccountStatusInactive, da.Status)
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
	ctx := context.Background()
	scimTestUser := suite.userBuilderWithInput(ctx, &userInput{
		email: ulids.New().String() + "@example.com",
	})

	// Enable SSO enforcement for SCIM (required for SCIM operations)
	allowCtx := privacy.DecisionContext(scimTestUser.UserCtx, privacy.Allow)
	org, err := suite.db.Organization.Get(allowCtx, scimTestUser.OrganizationID)
	suite.Require().NoError(err)
	setting, err := org.Setting(allowCtx)
	suite.Require().NoError(err)
	err = suite.db.OrganizationSetting.UpdateOneID(setting.ID).SetIdentityProviderLoginEnforced(true).Exec(allowCtx)
	suite.Require().NoError(err)

	restore := suite.registerSCIMRoutesWithAuth()
	defer restore()

	_, endpointID, secret := suite.createSCIMIntegration(allowCtx, scimTestUser.OrganizationID, "SCIM Directory")

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

	req := suite.newSCIMRequest(http.MethodPost, fmt.Sprintf("/v1/integrations/scim/%s/v2/Users", endpointID), secret, payload)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)
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

	req = suite.newSCIMRequest(http.MethodPatch, fmt.Sprintf("/v1/integrations/scim/%s/v2/Users/%s", endpointID, userID), secret, payload)

	rec = httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)
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

	req = suite.newSCIMRequest(http.MethodPatch, fmt.Sprintf("/v1/integrations/scim/%s/v2/Users/%s", endpointID, userID), secret, payload)

	rec = httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)
	suite.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())

	// Verify user was reactivated by checking the SCIM response
	var reactivatedResp map[string]any
	suite.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &reactivatedResp))
	reactivatedActive, ok := reactivatedResp["active"].(bool)
	suite.Require().True(ok, "active should be a boolean")
	suite.True(reactivatedActive, "user should be active after reactivation patch")
}

func (suite *HandlerTestSuite) TestSCIMGroupHandlerCreateDeduplicatesMembers() {
	ctx := context.Background()
	scimTestUser := suite.userBuilderWithInput(ctx, &userInput{
		email: ulids.New().String() + "@example.com",
	})

	// Enable SSO enforcement for SCIM (required for SCIM operations)
	allowCtx := privacy.DecisionContext(scimTestUser.UserCtx, privacy.Allow)
	org, err := suite.db.Organization.Get(allowCtx, scimTestUser.OrganizationID)
	suite.Require().NoError(err)
	setting, err := org.Setting(allowCtx)
	suite.Require().NoError(err)
	err = suite.db.OrganizationSetting.UpdateOneID(setting.ID).SetIdentityProviderLoginEnforced(true).Exec(allowCtx)
	suite.Require().NoError(err)

	restore := suite.registerSCIMRoutesWithAuth()
	defer restore()

	_, endpointID, secret := suite.createSCIMIntegration(allowCtx, scimTestUser.OrganizationID, "SCIM Directory")

	memberEmail := fmt.Sprintf("scim-member-%s@example.com", strings.ToLower(ulids.New().String()))
	memberPayload, err := json.Marshal(map[string]any{
		"schemas":  []string{scimUserSchema},
		"userName": memberEmail,
		"name": map[string]any{
			"givenName":  "Member",
			"familyName": "Target",
		},
	})
	suite.Require().NoError(err)

	memberReq := suite.newSCIMRequest(http.MethodPost, fmt.Sprintf("/v1/integrations/scim/%s/v2/Users", endpointID), secret, memberPayload)
	memberRec := httptest.NewRecorder()
	suite.e.ServeHTTP(memberRec, memberReq)
	suite.Require().Equal(http.StatusCreated, memberRec.Code, memberRec.Body.String())

	var createdMember map[string]any
	suite.Require().NoError(json.Unmarshal(memberRec.Body.Bytes(), &createdMember))
	memberID := suite.getStringField(createdMember, "id")

	displayName := "SCIM Engineering Team"
	body := map[string]any{
		"schemas":     []string{scimGroupSchema},
		"displayName": displayName,
		"members": []map[string]any{
			{"value": memberID},
			{"value": memberID},
		},
	}

	payload, err := json.Marshal(body)
	suite.Require().NoError(err)

	req := suite.newSCIMRequest(http.MethodPost, fmt.Sprintf("/v1/integrations/scim/%s/v2/Groups", endpointID), secret, payload)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)
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
	suite.Equal(memberID, suite.getStringField(memberMap, "value"), "member ID should match")
	suite.Equal(fmt.Sprintf("/v1/integrations/scim/%s/v2/Users/%s", endpointID, memberID), suite.getStringField(memberMap, "$ref"), "member ref should use the stable SCIM route")

	suite.Equal(displayName, resp["displayName"], "displayName should match")
}

func (suite *HandlerTestSuite) TestSCIMRouteRejectsPendingAndDisabledInstallations() {
	restore := suite.registerSCIMRoutesWithAuth()
	defer restore()

	ctx := context.Background()
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		email: ulids.New().String() + "@example.com",
	})

	allowCtx := privacy.DecisionContext(testUser.UserCtx, privacy.Allow)
	integrationID, endpointID, secret := suite.createSCIMIntegration(allowCtx, testUser.OrganizationID, "SCIM Directory")

	testCases := []struct {
		name   string
		status enums.IntegrationStatus
	}{
		{name: "pending", status: enums.IntegrationStatusPending},
		{name: "disabled", status: enums.IntegrationStatusDisabled},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.db.Integration.UpdateOneID(integrationID).
				SetStatus(tc.status).
				Exec(allowCtx)
			suite.Require().NoError(err)

			req := suite.newSCIMRequest(http.MethodGet, fmt.Sprintf("/v1/integrations/scim/%s/v2/Users", endpointID), secret, nil)
			rec := httptest.NewRecorder()
			suite.e.ServeHTTP(rec, req)

			suite.Require().Equal(http.StatusForbidden, rec.Code, rec.Body.String())

			err = suite.db.Integration.UpdateOneID(integrationID).
				SetStatus(enums.IntegrationStatusConnected).
				Exec(allowCtx)
			suite.Require().NoError(err)
		})
	}
}

func (suite *HandlerTestSuite) TestSCIMRouteRejectsNonSCIMIntegration() {
	restore := suite.registerSCIMRoutesWithAuth()
	defer restore()

	// Use a non-existent endpoint ID; the SCIM middleware should reject the
	// request because no matching IntegrationWebhook row exists.
	req := suite.newSCIMRequest(http.MethodGet, "/v1/integrations/scim/tolwh_nonexistent/v2/Users", "fake-secret", nil)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	suite.Require().NotEqual(http.StatusOK, rec.Code)
}
