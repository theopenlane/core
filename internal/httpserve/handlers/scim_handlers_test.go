package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/theopenlane/core/common/enums"
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/httpserve/route"
	definitionscim "github.com/theopenlane/core/v2/internal/integrations/definitions/scim"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
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
	assert.NilError(suite.T(), err)

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
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-" + orgID}}).
		Save(ctx)
	assert.NilError(suite.T(), err)

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
	assert.NilError(suite.T(), err)

	assert.Assert(suite.T(), webhook.EndpointID != nil)

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

func TestSCIMUserHandlerCreate(t *testing.T) {
	suite.setupTest(t)
	ctx := context.Background()
	scimTestUser := suite.userBuilderWithInput(ctx, &userInput{
		email: ulids.New().String() + "@example.com",
	})

	// Enable SSO enforcement for SCIM (required for SCIM operations)
	internalCtx := auth.WithInternalOperationContext(scimTestUser.UserCtx)
	org, err := suite.db.Organization.Get(internalCtx, scimTestUser.OrganizationID)
	assert.NilError(t, err)
	setting, err := org.Setting(internalCtx)
	assert.NilError(t, err)
	suite.enforceSSOOnSetting(internalCtx, setting.ID)

	restore := suite.registerSCIMRoutesWithAuth()
	defer restore()

	_, endpointID, secret := suite.createSCIMIntegration(internalCtx, scimTestUser.OrganizationID, "SCIM Directory")

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
		t.Run(tc.name, func(t *testing.T) {
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
			assert.NilError(t, err)

			req := suite.newSCIMRequest(http.MethodPost, fmt.Sprintf("/v1/integrations/scim/%s/v2/Users", endpointID), secret, payload)

			rec := httptest.NewRecorder()
			suite.e.ServeHTTP(rec, req)

			t.Logf("Test case: %s, Status: %d, Body: %s", tc.name, rec.Code, rec.Body.String())
			assert.Assert(t, is.Equal(http.StatusCreated, rec.Code), rec.Body.String())

			var resp map[string]any
			assert.NilError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

			userID := suite.getStringField(resp, "id")

			// Verify the response contains correct values
			active, ok := resp["active"].(bool)
			assert.Assert(t, ok, "active should be a boolean")
			assert.Check(t, is.DeepEqual(tc.active, active), "active flag should match")

			nameMap, ok := resp["name"].(map[string]any)
			assert.Assert(t, ok, "name should be a map")
			assert.Check(t, is.DeepEqual(tc.given, nameMap["givenName"]), "givenName should match")
			assert.Check(t, is.DeepEqual(tc.family, nameMap["familyName"]), "familyName should match")

			expectedDisplay := tc.displayName
			if expectedDisplay == "" {
				expectedDisplay = strings.TrimSpace(tc.given + " " + tc.family)
				if expectedDisplay == "" {
					expectedDisplay = strings.ToLower(email)
				}
			}
			assert.Check(t, is.DeepEqual(expectedDisplay, resp["displayName"]), "displayName should match")

			// Verify DirectoryAccount was created with correct attributes
			da, err := suite.db.DirectoryAccount.Get(internalCtx, userID)
			assert.NilError(t, err)

			expectedEmail := strings.ToLower(email)
			assert.Assert(t, da.CanonicalEmail != nil)
			assert.Check(t, is.DeepEqual(expectedEmail, *da.CanonicalEmail))
			assert.Check(t, is.DeepEqual(expectedDisplay, da.DisplayName))

			assert.Assert(t, da.GivenName != nil)
			assert.Check(t, is.DeepEqual(tc.given, *da.GivenName))
			assert.Assert(t, da.FamilyName != nil)
			assert.Check(t, is.DeepEqual(tc.family, *da.FamilyName))

			if tc.active {
				assert.Check(t, is.DeepEqual(enums.DirectoryAccountStatusActive, da.Status))
			} else {
				assert.Check(t, is.DeepEqual(enums.DirectoryAccountStatusInactive, da.Status))
			}
		})
	}
}

func (suite *HandlerTestSuite) getStringField(data map[string]any, key string) string {
	value, ok := data[key]
	assert.Assert(suite.T(), ok, "missing key %s", key)

	str, ok := value.(string)
	assert.Assert(suite.T(), ok, "key %s is not a string", key)

	return str
}

func TestSCIMUserHandlerPatchActiveToggle(t *testing.T) {
	suite.setupTest(t)
	ctx := context.Background()
	scimTestUser := suite.userBuilderWithInput(ctx, &userInput{
		email: ulids.New().String() + "@example.com",
	})

	// Enable SSO enforcement for SCIM (required for SCIM operations)
	internalCtx := auth.WithInternalOperationContext(scimTestUser.UserCtx)
	org, err := suite.db.Organization.Get(internalCtx, scimTestUser.OrganizationID)
	assert.NilError(t, err)
	setting, err := org.Setting(internalCtx)
	assert.NilError(t, err)
	suite.enforceSSOOnSetting(internalCtx, setting.ID)

	restore := suite.registerSCIMRoutesWithAuth()
	defer restore()

	_, endpointID, secret := suite.createSCIMIntegration(internalCtx, scimTestUser.OrganizationID, "SCIM Directory")

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
	assert.NilError(t, err)

	req := suite.newSCIMRequest(http.MethodPost, fmt.Sprintf("/v1/integrations/scim/%s/v2/Users", endpointID), secret, payload)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)
	assert.Assert(t, is.Equal(http.StatusCreated, rec.Code), rec.Body.String())

	var created map[string]any
	assert.NilError(t, json.Unmarshal(rec.Body.Bytes(), &created))
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
	assert.NilError(t, err)

	req = suite.newSCIMRequest(http.MethodPatch, fmt.Sprintf("/v1/integrations/scim/%s/v2/Users/%s", endpointID, userID), secret, payload)

	rec = httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)
	assert.Assert(t, is.Equal(http.StatusOK, rec.Code), rec.Body.String())

	// Verify user was deactivated by checking the SCIM response
	var deactivatedResp map[string]any
	assert.NilError(t, json.Unmarshal(rec.Body.Bytes(), &deactivatedResp))
	deactivatedActive, ok := deactivatedResp["active"].(bool)
	assert.Assert(t, ok, "active should be a boolean")
	assert.Check(t, !deactivatedActive, "user should be inactive after patch")

	patchBody["Operations"].([]map[string]any)[0]["value"] = map[string]any{"active": true}

	payload, err = json.Marshal(patchBody)
	assert.NilError(t, err)

	req = suite.newSCIMRequest(http.MethodPatch, fmt.Sprintf("/v1/integrations/scim/%s/v2/Users/%s", endpointID, userID), secret, payload)

	rec = httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)
	assert.Assert(t, is.Equal(http.StatusOK, rec.Code), rec.Body.String())

	// Verify user was reactivated by checking the SCIM response
	var reactivatedResp map[string]any
	assert.NilError(t, json.Unmarshal(rec.Body.Bytes(), &reactivatedResp))
	reactivatedActive, ok := reactivatedResp["active"].(bool)
	assert.Assert(t, ok, "active should be a boolean")
	assert.Check(t, reactivatedActive, "user should be active after reactivation patch")
}

func TestSCIMGroupHandlerCreateDeduplicatesMembers(t *testing.T) {
	suite.setupTest(t)
	ctx := context.Background()
	scimTestUser := suite.userBuilderWithInput(ctx, &userInput{
		email: ulids.New().String() + "@example.com",
	})

	// Enable SSO enforcement for SCIM (required for SCIM operations)
	internalCtx := auth.WithInternalOperationContext(scimTestUser.UserCtx)
	org, err := suite.db.Organization.Get(internalCtx, scimTestUser.OrganizationID)
	assert.NilError(t, err)
	setting, err := org.Setting(internalCtx)
	assert.NilError(t, err)
	suite.enforceSSOOnSetting(internalCtx, setting.ID)

	restore := suite.registerSCIMRoutesWithAuth()
	defer restore()

	_, endpointID, secret := suite.createSCIMIntegration(internalCtx, scimTestUser.OrganizationID, "SCIM Directory")

	memberEmail := fmt.Sprintf("scim-member-%s@example.com", strings.ToLower(ulids.New().String()))
	memberPayload, err := json.Marshal(map[string]any{
		"schemas":  []string{scimUserSchema},
		"userName": memberEmail,
		"name": map[string]any{
			"givenName":  "Member",
			"familyName": "Target",
		},
	})
	assert.NilError(t, err)

	memberReq := suite.newSCIMRequest(http.MethodPost, fmt.Sprintf("/v1/integrations/scim/%s/v2/Users", endpointID), secret, memberPayload)
	memberRec := httptest.NewRecorder()
	suite.e.ServeHTTP(memberRec, memberReq)
	assert.Assert(t, is.Equal(http.StatusCreated, memberRec.Code), memberRec.Body.String())

	var createdMember map[string]any
	assert.NilError(t, json.Unmarshal(memberRec.Body.Bytes(), &createdMember))
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
	assert.NilError(t, err)

	req := suite.newSCIMRequest(http.MethodPost, fmt.Sprintf("/v1/integrations/scim/%s/v2/Groups", endpointID), secret, payload)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)
	suite.T().Logf("Group create test - Status: %d, Body: %s", rec.Code, rec.Body.String())
	assert.Assert(t, is.Equal(http.StatusCreated, rec.Code), rec.Body.String())

	var resp map[string]any
	assert.NilError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	groupID := suite.getStringField(resp, "id")
	suite.T().Logf("Created group ID: %s", groupID)

	membersValue, ok := resp["members"].([]any)
	assert.Assert(t, ok, "members should be an array")
	assert.Check(t, is.Len(membersValue, 1), "should have exactly 1 member after deduplication")

	memberMap, ok := membersValue[0].(map[string]any)
	assert.Assert(t, ok, "member should be a map")
	assert.Check(t, is.DeepEqual(memberID, suite.getStringField(memberMap, "value")), "member ID should match")
	assert.Check(t, is.Equal(fmt.Sprintf("/v1/integrations/scim/%s/v2/Users/%s", endpointID, memberID), suite.getStringField(memberMap, "$ref")), "member ref should use the stable SCIM route")

	assert.Check(t, is.DeepEqual(displayName, resp["displayName"]), "displayName should match")
}

func TestSCIMRouteRejectsPendingAndDisabledInstallations(t *testing.T) {
	suite.setupTest(t)
	restore := suite.registerSCIMRoutesWithAuth()
	defer restore()

	ctx := context.Background()
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		email: ulids.New().String() + "@example.com",
	})

	internalCtx := auth.WithInternalOperationContext(testUser.UserCtx)
	integrationID, endpointID, secret := suite.createSCIMIntegration(internalCtx, testUser.OrganizationID, "SCIM Directory")

	testCases := []struct {
		name   string
		status enums.IntegrationStatus
	}{
		{name: "pending", status: enums.IntegrationStatusPending},
		{name: "disabled", status: enums.IntegrationStatusDisabled},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := suite.db.Integration.UpdateOneID(integrationID).
				SetStatus(tc.status).
				Exec(internalCtx)
			assert.NilError(t, err)

			req := suite.newSCIMRequest(http.MethodGet, fmt.Sprintf("/v1/integrations/scim/%s/v2/Users", endpointID), secret, nil)
			rec := httptest.NewRecorder()
			suite.e.ServeHTTP(rec, req)

			assert.Assert(t, is.Equal(http.StatusForbidden, rec.Code), rec.Body.String())

			err = suite.db.Integration.UpdateOneID(integrationID).
				SetStatus(enums.IntegrationStatusConnected).
				Exec(internalCtx)
			assert.NilError(t, err)
		})
	}
}

func TestSCIMRouteRejectsNonSCIMIntegration(t *testing.T) {
	suite.setupTest(t)
	restore := suite.registerSCIMRoutesWithAuth()
	defer restore()

	// Use a non-existent endpoint ID; the SCIM middleware should reject the
	// request because no matching IntegrationWebhook row exists.
	req := suite.newSCIMRequest(http.MethodGet, "/v1/integrations/scim/tolwh_nonexistent/v2/Users", "fake-secret", nil)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	assert.Assert(t, http.StatusOK != rec.Code)
}
