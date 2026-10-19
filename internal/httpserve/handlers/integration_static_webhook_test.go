//go:build test

package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/v2/internal/ent/generated/integrationwebhook"
	"github.com/theopenlane/core/v2/internal/httpserve/handlers"
	"github.com/theopenlane/echox/middleware/echocontext"
)

func (suite *HandlerTestSuite) registerStaticWebhookTestRoutes() {
	suite.registerRouteOnce(http.MethodPost, staticWebhookTestPath, suite.h.IntegrationStaticWebhookHandler(staticWebhookTestDefinitionID, staticWebhookTestVerifiedName))
	suite.registerRouteOnce(http.MethodPost, staticWebhookUnverifiedTestPath, suite.h.IntegrationStaticWebhookHandler(staticWebhookTestDefinitionID, staticWebhookTestUnverifiedName))
}

func postStaticWebhook(suite *HandlerTestSuite, path, signature string, payload []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")

	if signature != "" {
		req.Header.Set(staticWebhookTestSignatureHeader, signature)
	}

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	return rec
}

func (suite *HandlerTestSuite) TestIntegrationStaticWebhookHandlerRejectsMissingVerifier() {
	t := suite.T()

	suite.registerStaticWebhookTestRoutes()

	rec := postStaticWebhook(suite, staticWebhookUnverifiedTestPath, webhookTestSecret, []byte(`{"event":"alert.created","delivery_id":"static-001"}`))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "verification failed")
}

func (suite *HandlerTestSuite) TestIntegrationStaticWebhookHandlerRejectsInvalidSignature() {
	t := suite.T()

	suite.registerStaticWebhookTestRoutes()

	payload := []byte(`{"event":"alert.created","delivery_id":"static-002"}`)

	for _, signature := range []string{"", "wrong-secret"} {
		rec := postStaticWebhook(suite, staticWebhookTestPath, signature, payload)

		require.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "verification failed")
		assert.NotContains(t, rec.Body.String(), "signature mismatch")
	}
}

func (suite *HandlerTestSuite) TestIntegrationStaticWebhookHandlerSkipsUnregisteredEvent() {
	t := suite.T()

	suite.registerStaticWebhookTestRoutes()

	for _, event := range []string{"", staticWebhookTestUnregisteredName} {
		payload := []byte(`{"event":"` + event + `","delivery_id":"static-003"}`)

		require.Equal(t, http.StatusOK, postStaticWebhook(suite, staticWebhookTestPath, webhookTestSecret, payload).Code, "event %q", event)
	}
}

func (suite *HandlerTestSuite) TestIntegrationStaticWebhookHandlerDispatchesRegisteredEvent() {
	t := suite.T()

	suite.registerStaticWebhookTestRoutes()

	rec := postStaticWebhook(suite, staticWebhookTestPath, webhookTestSecret, []byte(`{"event":"alert.created","delivery_id":"static-004"}`))

	require.Equal(t, http.StatusOK, rec.Code)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderSkipsRuntimeOwnedWebhooks() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	rec := sendIntegrationConfigRequest(t, suite, testUser.UserCtx, staticWebhookTestDefinitionID, handlers.ConfigureIntegrationRequest{
		DefinitionID:  staticWebhookTestDefinitionID,
		CredentialRef: webhookTestCredentialRef,
		Body:          json.RawMessage(`{"token":"test-token"}`),
	})

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var resp handlers.ConfigureIntegrationResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	assert.NotEmpty(t, resp.IntegrationID)
	assert.Empty(t, resp.WebhookSecret)
	assert.Empty(t, resp.WebhookEndpointURL)

	rows := suite.db.IntegrationWebhook.Query().
		Where(integrationwebhook.IntegrationIDEQ(resp.IntegrationID)).
		CountX(testUser.UserCtx)

	assert.Zero(t, rows)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderRejectsOperatorOnlyDefinition() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	rec := sendIntegrationConfigRequest(t, suite, testUser.UserCtx, operatorOnlyTestDefinitionID, handlers.ConfigureIntegrationRequest{
		DefinitionID: operatorOnlyTestDefinitionID,
	})

	require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	assert.Contains(t, rec.Body.String(), handlers.ErrInvalidProvider.Error())
}

func (suite *HandlerTestSuite) TestRunIntegrationOperationRejectsHiddenOperation() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)
	suite.registerRouteOnce(http.MethodPost, operationTestPath, suite.h.RunIntegrationOperation)

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	configRec := sendIntegrationConfigRequest(t, suite, testUser.UserCtx, staticWebhookTestDefinitionID, handlers.ConfigureIntegrationRequest{
		DefinitionID:  staticWebhookTestDefinitionID,
		CredentialRef: webhookTestCredentialRef,
		Body:          json.RawMessage(`{"token":"test-token"}`),
	})
	require.Equal(t, http.StatusOK, configRec.Code, configRec.Body.String())

	var configResp handlers.ConfigureIntegrationResponse
	require.NoError(t, json.Unmarshal(configRec.Body.Bytes(), &configResp))

	body, err := json.Marshal(handlers.RunIntegrationOperationRequest{
		IntegrationID: configResp.IntegrationID,
		Body:          handlers.RunIntegrationOperationBody{Operation: staticWebhookHiddenOperation.Name()},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+staticWebhookTestDefinitionID+"/operations?integration_id="+configResp.IntegrationID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	assert.Contains(t, rec.Body.String(), "dispatch input invalid")
}
