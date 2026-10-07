//go:build test

package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/entity"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/integrationwebhook"
	"github.com/theopenlane/core/v2/internal/httpserve/handlers"
	definitionscim "github.com/theopenlane/core/v2/internal/integrations/definitions/scim"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderRequiresOrgEdit() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, false)})
	t.Cleanup(restore)

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	owner := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})
	admin := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})
	member := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	memberships := []struct {
		userID string
		role   enums.Role
	}{
		{userID: admin.ID, role: enums.RoleAdmin},
		{userID: member.ID, role: enums.RoleMember},
	}

	for _, m := range memberships {
		err := suite.db.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
			OrganizationID: owner.OrganizationID,
			UserID:         m.userID,
			Role:           &m.role,
		}).Exec(owner.UserCtx)
		require.NoError(t, err)
	}

	testCases := []struct {
		name           string
		userID         string
		expectedStatus int
		expectStored   bool
	}{
		{
			name:           "member cannot configure an integration",
			userID:         member.ID,
			expectedStatus: http.StatusBadRequest,
			expectStored:   false,
		},
		{
			name:           "admin can configure an integration",
			userID:         admin.ID,
			expectedStatus: http.StatusOK,
			expectStored:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := sendIntegrationConfigRequest(t, suite, auth.NewTestContextWithOrgID(tc.userID, owner.OrganizationID), configTestProviderID, handlers.ConfigureIntegrationRequest{
				DefinitionID:  configTestProviderID,
				CredentialRef: configTestCredentialRef.String(),
				Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "sample-project", "serviceAccountEmail": "svc@example.iam.gserviceaccount.com"})),
			})

			assert.Equal(t, tc.expectedStatus, rec.Code)

			stored, err := suite.db.Integration.Query().
				Where(
					integration.OwnerIDEQ(owner.OrganizationID),
					integration.DefinitionIDEQ(configTestProviderID),
				).
				Exist(owner.UserCtx)
			require.NoError(t, err)
			assert.Equal(t, tc.expectStored, stored)
		})
	}

	t.Run("member cannot replace credentials on an existing integration", func(t *testing.T) {
		installation, err := suite.db.Integration.Query().
			Where(
				integration.OwnerIDEQ(owner.OrganizationID),
				integration.DefinitionIDEQ(configTestProviderID),
			).
			Only(owner.UserCtx)
		require.NoError(t, err)

		rec := sendIntegrationConfigRequest(t, suite, auth.NewTestContextWithOrgID(member.ID, owner.OrganizationID), configTestProviderID, handlers.ConfigureIntegrationRequest{
			DefinitionID:  configTestProviderID,
			IntegrationID: installation.ID,
			CredentialRef: configTestCredentialRef.String(),
			Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "member-project", "serviceAccountEmail": "member@example.iam.gserviceaccount.com"})),
		})

		assert.NotEqual(t, http.StatusOK, rec.Code)

		credential, ok, err := suite.h.IntegrationsRuntime.LoadCredential(owner.UserCtx, installation, configTestCredentialRef)
		require.NoError(t, err)
		require.True(t, ok)
		assert.Contains(t, string(credential.Data), "sample-project")
		assert.NotContains(t, string(credential.Data), "member-project")
	})
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderLinksExistingVendor() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	owner := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	vendor, err := suite.db.Entity.Create().
		SetName(configTestVendorFamily).
		SetOwnerID(owner.OrganizationID).
		Save(owner.UserCtx)
	require.NoError(t, err)

	rec := sendIntegrationConfigRequest(t, suite, auth.NewTestContextWithOrgID(owner.ID, owner.OrganizationID), configTestVendorProviderID, handlers.ConfigureIntegrationRequest{
		DefinitionID:  configTestVendorProviderID,
		CredentialRef: configTestCredentialRef.String(),
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "vendor-project", "serviceAccountEmail": "vendor@example.iam.gserviceaccount.com"})),
	})

	require.Equal(t, http.StatusOK, rec.Code)

	linked, err := suite.db.Entity.Query().
		Where(
			entity.IDEQ(vendor.ID),
			entity.HasIntegrationsWith(integration.DefinitionIDEQ(configTestVendorProviderID)),
		).
		Exist(owner.UserCtx)
	require.NoError(t, err)
	assert.True(t, linked)
}

func (suite *HandlerTestSuite) TestCheckIntegrationHealth() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:integrationID/health", suite.h.CheckIntegrationHealth)

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	owner := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	testCases := []struct {
		name               string
		definitionID       string
		startErrored       bool
		expectedStatus     enums.IntegrationStatus
		expectProbe        bool
		expectProbeHealthy bool
		expectProbeReason  string
	}{
		{
			name:           "errored installation recovers to connected",
			definitionID:   configTestProviderID,
			startErrored:   true,
			expectedStatus: enums.IntegrationStatusConnected,
		},
		{
			name:               "passing operation probe keeps the installation connected",
			definitionID:       configTestProbeProviderID,
			expectedStatus:     enums.IntegrationStatusConnected,
			expectProbe:        true,
			expectProbeHealthy: true,
		},
		{
			name:              "failing operation probe degrades the installation",
			definitionID:      configTestProbeFailProviderID,
			expectedStatus:    enums.IntegrationStatusDegraded,
			expectProbe:       true,
			expectProbeReason: errConfigTestProbeFailed.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			installation := suite.configureIntegrationAsUser(t, owner, tc.definitionID)

			if tc.startErrored {
				require.NoError(t, suite.db.Integration.UpdateOneID(installation.ID).SetStatus(enums.IntegrationStatusErrored).Exec(owner.UserCtx))
			}

			resp := suite.checkIntegrationHealthAsUser(t, owner, installation.ID)

			updated, err := suite.db.Integration.Get(owner.UserCtx, installation.ID)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, updated.Status)

			if !tc.expectProbe {
				assert.NotNil(t, updated.Health.LastSuccessfulHealthCheck)

				return
			}

			probe, ok := lo.Find(resp.Operations, func(op handlers.IntegrationOperationHealth) bool { return op.Name == configTestProbeOperation })
			require.True(t, ok)
			assert.Equal(t, tc.expectProbeHealthy, probe.Healthy)
			assert.Contains(t, probe.Reason, tc.expectProbeReason)
		})
	}
}

func (suite *HandlerTestSuite) configureIntegrationAsUser(t *testing.T, user testUserDetails, definitionID string) *generated.Integration {
	t.Helper()

	rec := sendIntegrationConfigRequest(t, suite, auth.NewTestContextWithOrgID(user.ID, user.OrganizationID), definitionID, handlers.ConfigureIntegrationRequest{
		DefinitionID:  definitionID,
		CredentialRef: configTestCredentialRef.String(),
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "health-project", "serviceAccountEmail": "health@example.iam.gserviceaccount.com"})),
	})
	require.Equal(t, http.StatusOK, rec.Code)

	installation, err := suite.db.Integration.Query().
		Where(
			integration.OwnerIDEQ(user.OrganizationID),
			integration.DefinitionIDEQ(definitionID),
		).
		Only(user.UserCtx)
	require.NoError(t, err)

	return installation
}

func (suite *HandlerTestSuite) checkIntegrationHealthAsUser(t *testing.T, user testUserDetails, integrationID string) handlers.IntegrationHealthResponse {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+integrationID+"/health", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(auth.NewTestContextWithOrgID(user.ID, user.OrganizationID)))
	require.Equal(t, http.StatusOK, rec.Code)

	var resp handlers.IntegrationHealthResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	return resp
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderSuccess() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	rec := sendIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestProviderID, handlers.ConfigureIntegrationRequest{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "sample-project", "serviceAccountEmail": "svc@example.iam.gserviceaccount.com"})),
		UserInput: json.RawMessage(
			mustMarshalJSON(t, map[string]any{"filterExpr": "payload.severity == \"HIGH\""}),
		),
	})

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp handlers.ConfigureIntegrationResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, configTestProviderID, resp.Provider)

	stored := suite.db.Integration.Query().
		Where(
			integration.OwnerIDEQ(testUser.OrganizationID),
			integration.DefinitionIDEQ(configTestProviderID),
		).
		OnlyX(testUser.UserCtx)

	credential, ok, err := suite.keystore.LoadCredential(testUser.UserCtx, stored, configTestCredentialRef)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Contains(t, string(credential.Data), "projectId")
	assert.Equal(t, configTestUserInputRef.Name(), stored.UserInput.Layout)
	assert.Equal(t, `payload.severity == "HIGH"`, decodeUserInputField(t, stored.UserInput.Data, "filterExpr"))
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderStoresOperationConfig() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{operationTestDefinitionBuilder(operationTestDefinitionID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	resp := performIntegrationConfigRequest(t, suite, testUser.UserCtx, operationTestDefinitionID, handlers.ConfigureIntegrationRequest{
		DefinitionID:  operationTestDefinitionID,
		CredentialRef: operationTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"token": "test-token"})),
		OperationConfig: map[string]json.RawMessage{
			opTestValidatedOperation.Name(): json.RawMessage(`{"target":"repo","disable":true}`),
		},
	})

	stored := suite.db.Integration.GetX(testUser.UserCtx, resp.IntegrationID)
	assert.Equal(t, enums.IntegrationStatusConnected, stored.Status)
	assert.JSONEq(t, `{"target":"repo","disable":true}`, string(stored.OperationConfig.For(opTestValidatedOperation.Name())))
	assert.Nil(t, stored.OperationConfig.For(opTestRepoSyncOperation.Name()))
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderRejectsInvalidOperationConfig() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{operationTestDefinitionBuilder(operationTestDefinitionID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body := mustMarshalConfigPayload(t, handlers.ConfigureIntegrationRequest{
		DefinitionID:  operationTestDefinitionID,
		CredentialRef: operationTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"token": "test-token"})),
		OperationConfig: map[string]json.RawMessage{
			opTestValidatedOperation.Name(): json.RawMessage(`{"target":1}`),
		},
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestDefinitionID+"/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderRejectsInvalidFilterExpr() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{operationTestDefinitionBuilder(operationTestDefinitionID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body := mustMarshalConfigPayload(t, handlers.ConfigureIntegrationRequest{
		DefinitionID:  operationTestDefinitionID,
		CredentialRef: operationTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"token": "test-token"})),
		OperationConfig: map[string]json.RawMessage{
			opTestValidatedOperation.Name(): json.RawMessage(`{"target":"repo","filterExpr":"payload."}`),
		},
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestDefinitionID+"/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "filter expression invalid")
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderReturnsSCIMEndpointDetails() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{definitionscim.Builder()})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	rec := sendIntegrationConfigRequest(t, suite, testUser.UserCtx, definitionscim.DefinitionID.ID(), handlers.ConfigureIntegrationRequest{
		DefinitionID: definitionscim.DefinitionID.ID(),
		UserInput:    json.RawMessage(mustMarshalJSON(t, map[string]any{"name": "Okta Production"})),
	})

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp handlers.ConfigureIntegrationResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	assert.True(t, resp.Success)
	assert.Equal(t, definitionscim.DefinitionID.ID(), resp.Provider)
	assert.NotEmpty(t, resp.IntegrationID)
	assert.NotEmpty(t, resp.WebhookSecret)
	assert.True(t, strings.HasPrefix(resp.WebhookEndpointURL, "http://example.com/v1/integrations/scim/"))
	assert.True(t, strings.HasSuffix(resp.WebhookEndpointURL, "/v2"))

	stored := suite.db.Integration.GetX(testUser.UserCtx, resp.IntegrationID)
	assert.Equal(t, enums.IntegrationStatusConnected, stored.Status)
	assert.Equal(t, "Okta Production", stored.Name)

	webhook := suite.db.IntegrationWebhook.Query().
		Where(
			integrationwebhook.IntegrationIDEQ(stored.ID),
			integrationwebhook.NameEQ(definitionscim.SCIMAuthWebhook.Name()),
			integrationwebhook.ExternalEventIDIsNil(),
		).
		OnlyX(testUser.UserCtx)

	assert.NotNil(t, webhook.EndpointID)
	assert.NotNil(t, webhook.EndpointURL)
	assert.Equal(t, "/v1/integrations/scim/"+*webhook.EndpointID+"/v2", *webhook.EndpointURL)
	assert.Equal(t, "http://example.com"+*webhook.EndpointURL, resp.WebhookEndpointURL)
	assert.Equal(t, webhook.SecretToken, resp.WebhookSecret)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderAcceptsDefinitionID() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	rec := sendIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestProviderID, handlers.ConfigureIntegrationRequest{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "sample-project", "serviceAccountEmail": "svc@example.iam.gserviceaccount.com"})),
	})

	assert.Equal(t, http.StatusOK, rec.Code)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderInvalidPayload() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	rec := sendIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestProviderID, handlers.ConfigureIntegrationRequest{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"serviceAccountEmail": "svc@example.iam.gserviceaccount.com"})),
	})

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderRejectsNonObjectPayload() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	rec := sendIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestProviderID, handlers.ConfigureIntegrationRequest{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(`["not","an","object"]`),
	})

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderUnauthorized() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, false)})
	defer restore()

	rec := sendIntegrationConfigRequest(t, suite, context.Background(), configTestProviderID, handlers.ConfigureIntegrationRequest{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "sample-project", "serviceAccountEmail": "svc@example.iam.gserviceaccount.com"})),
	})

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderUpdateExisting() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	first := performIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestProviderID, handlers.ConfigureIntegrationRequest{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "initial-project", "serviceAccountEmail": "initial@example.iam.gserviceaccount.com"})),
	})

	second := performIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestProviderID, handlers.ConfigureIntegrationRequest{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		IntegrationID: first.IntegrationID,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "updated-project", "serviceAccountEmail": "updated@example.iam.gserviceaccount.com"})),
	})

	assert.Equal(t, first.IntegrationID, second.IntegrationID)

	stored := suite.db.Integration.GetX(testUser.UserCtx, first.IntegrationID)
	credential, ok, err := suite.keystore.LoadCredential(testUser.UserCtx, stored, configTestCredentialRef)
	assert.NoError(t, err)
	assert.True(t, ok)

	providerData, err := jsonx.ToMap(credential.Data)
	assert.NoError(t, err)
	assert.Equal(t, "updated-project", providerData["projectId"])
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderUpdateExistingUserInputOnly() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	first := performIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestProviderID, handlers.ConfigureIntegrationRequest{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "initial-project", "serviceAccountEmail": "initial@example.iam.gserviceaccount.com"})),
	})

	second := performIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestProviderID, handlers.ConfigureIntegrationRequest{
		DefinitionID:  configTestProviderID,
		IntegrationID: first.IntegrationID,
		UserInput:     json.RawMessage(mustMarshalJSON(t, map[string]any{"filterExpr": "payload.category == \"critical\""})),
	})

	assert.Equal(t, first.IntegrationID, second.IntegrationID)

	stored := suite.db.Integration.GetX(testUser.UserCtx, first.IntegrationID)
	assert.Equal(t, enums.IntegrationStatusConnected, stored.Status)
	assert.Equal(t, `payload.category == "critical"`, decodeUserInputField(t, stored.UserInput.Data, "filterExpr"))

	credential, ok, err := suite.keystore.LoadCredential(testUser.UserCtx, stored, configTestCredentialRef)
	assert.NoError(t, err)
	assert.True(t, ok)

	providerData, err := jsonx.ToMap(credential.Data)
	assert.NoError(t, err)
	assert.Equal(t, "initial-project", providerData["projectId"])
	assert.Equal(t, "initial@example.iam.gserviceaccount.com", providerData["serviceAccountEmail"])
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderUpdateExistingUserInputOnlyWithEmptyObjectBody() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	first := performIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestProviderID, handlers.ConfigureIntegrationRequest{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "initial-project", "serviceAccountEmail": "initial@example.iam.gserviceaccount.com"})),
	})

	second := performIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestProviderID, handlers.ConfigureIntegrationRequest{
		DefinitionID:  configTestProviderID,
		IntegrationID: first.IntegrationID,
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(`{}`),
		UserInput:     json.RawMessage(mustMarshalJSON(t, map[string]any{"filterExpr": "payload.category == \"critical\""})),
	})

	assert.Equal(t, first.IntegrationID, second.IntegrationID)

	stored := suite.db.Integration.GetX(testUser.UserCtx, first.IntegrationID)
	assert.Equal(t, enums.IntegrationStatusConnected, stored.Status)
	assert.Equal(t, `payload.category == "critical"`, decodeUserInputField(t, stored.UserInput.Data, "filterExpr"))

	credential, ok, err := suite.keystore.LoadCredential(testUser.UserCtx, stored, configTestCredentialRef)
	assert.NoError(t, err)
	assert.True(t, ok)

	providerData, err := jsonx.ToMap(credential.Data)
	assert.NoError(t, err)
	assert.Equal(t, "initial-project", providerData["projectId"])
	assert.Equal(t, "initial@example.iam.gserviceaccount.com", providerData["serviceAccountEmail"])
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderAllowsUserInputOnlyUpdateWithoutCredentialSchema() {
	t := suite.T()

	const definitionID = "def_01K0TESTUIONLY000000000001"

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{userInputOnlyTestDefinitionBuilder(definitionID)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	rec := suite.db.Integration.Create().
		SetOwnerID(testUser.OrganizationID).
		SetName("OIDC Generic").
		SetDefinitionID(definitionID).
		SetStatus(enums.IntegrationStatusConnected).
		SaveX(testUser.UserCtx)

	httpRec := sendIntegrationConfigRequest(t, suite, testUser.UserCtx, definitionID, handlers.ConfigureIntegrationRequest{
		DefinitionID:  definitionID,
		IntegrationID: rec.ID,
		UserInput:     json.RawMessage(mustMarshalJSON(t, map[string]any{"filterExpr": "payload.actor == \"service-account\""})),
	})

	assert.Equal(t, http.StatusOK, httpRec.Code)

	stored := suite.db.Integration.GetX(testUser.UserCtx, rec.ID)
	assert.Equal(t, configTestUserInputRef.Name(), stored.UserInput.Layout)
	assert.Equal(t, `payload.actor == "service-account"`, decodeUserInputField(t, stored.UserInput.Data, "filterExpr"))
	assert.Equal(t, enums.IntegrationStatusConnected, stored.Status)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderRejectsInstallationDefinitionMismatch() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{
		configTestDefinitionBuilder(configTestProviderID, false),
		configTestDefinitionBuilder("def_01K0TESTOTH00000000000001", false),
	})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	other := performIntegrationConfigRequest(t, suite, testUser.UserCtx, "def_01K0TESTOTH00000000000001", handlers.ConfigureIntegrationRequest{
		DefinitionID:  "def_01K0TESTOTH00000000000001",
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "other-project", "serviceAccountEmail": "other@example.iam.gserviceaccount.com"})),
	})

	rec := sendIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestProviderID, handlers.ConfigureIntegrationRequest{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		IntegrationID: other.IntegrationID,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "sample-project", "serviceAccountEmail": "svc@example.iam.gserviceaccount.com"})),
	})

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderHealthFailureDoesNotPersistCredential() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestFailHealthProviderID, true)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	rec := sendIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestFailHealthProviderID, handlers.ConfigureIntegrationRequest{
		DefinitionID:  configTestFailHealthProviderID,
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "sample-project", "serviceAccountEmail": "svc@example.iam.gserviceaccount.com"})),
	})

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	records, err := suite.db.Integration.Query().
		Where(
			integration.OwnerIDEQ(testUser.OrganizationID),
			integration.DefinitionIDEQ(configTestFailHealthProviderID),
		).
		All(testUser.UserCtx)
	assert.NoError(t, err)
	assert.Len(t, records, 1, "expected one PENDING installation row after failed setup")
	assert.Equal(t, enums.IntegrationStatusPending, records[0].Status)

	_, credOk, credErr := suite.keystore.LoadCredential(testUser.UserCtx, records[0], configTestCredentialRef)
	assert.NoError(t, credErr)
	assert.False(t, credOk, "credential must not be stored after a failed health check")
}

func sendIntegrationConfigRequest(t *testing.T, suite *HandlerTestSuite, ctx context.Context, provider string, payload handlers.ConfigureIntegrationRequest) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+provider+"/config", bytes.NewReader(mustMarshalConfigPayload(t, payload)))
	req.Header.Set("Content-Type", "application/json")
	suite.e.ServeHTTP(rec, req.WithContext(ctx))

	return rec
}

func performIntegrationConfigRequest(t *testing.T, suite *HandlerTestSuite, ctx context.Context, provider string, payload handlers.ConfigureIntegrationRequest) handlers.ConfigureIntegrationResponse {
	t.Helper()

	rec := sendIntegrationConfigRequest(t, suite, ctx, provider, payload)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp handlers.ConfigureIntegrationResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	return resp
}

func mustMarshalConfigPayload(t *testing.T, payload handlers.ConfigureIntegrationRequest) []byte {
	t.Helper()

	body, err := json.Marshal(payload)
	assert.NoError(t, err)

	return body
}

func mustMarshalJSON(t *testing.T, value any) []byte {
	t.Helper()

	body, err := json.Marshal(value)
	assert.NoError(t, err)

	return body
}

func decodeUserInputField(t *testing.T, raw json.RawMessage, key string) string {
	t.Helper()

	document, err := jsonx.ToMap(raw)
	assert.NoError(t, err)

	value, _ := document[key].(string)

	return value
}
