//go:build test

package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/theopenlane/core/v2/internal/ent/generated/integrationwebhook"
	"github.com/theopenlane/core/v2/internal/httpserve/handlers"
	"github.com/theopenlane/echox/middleware/echocontext"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
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

func TestIntegrationStaticWebhookHandlerRejectsMissingVerifier(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerStaticWebhookTestRoutes()

	rec := postStaticWebhook(suite, staticWebhookUnverifiedTestPath, webhookTestSecret, []byte(`{"event":"alert.created","delivery_id":"static-001"}`))

	assert.Assert(t, is.Equal(http.StatusBadRequest, rec.Code))
	assert.Check(t, is.Contains(rec.Body.String(), "verification failed"))
}

func TestIntegrationStaticWebhookHandlerRejectsInvalidSignature(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerStaticWebhookTestRoutes()

	payload := []byte(`{"event":"alert.created","delivery_id":"static-002"}`)

	for _, signature := range []string{"", "wrong-secret"} {
		rec := postStaticWebhook(suite, staticWebhookTestPath, signature, payload)

		assert.Assert(t, is.Equal(http.StatusBadRequest, rec.Code))
		assert.Check(t, is.Contains(rec.Body.String(), "verification failed"))
		assert.Check(t, !strings.Contains(rec.Body.String(), "signature mismatch"))
	}
}

func TestIntegrationStaticWebhookHandlerSkipsUnregisteredEvent(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerStaticWebhookTestRoutes()

	for _, event := range []string{"", staticWebhookTestUnregisteredName} {
		payload := []byte(`{"event":"` + event + `","delivery_id":"static-003"}`)

		assert.Assert(t, is.Equal(http.StatusOK, postStaticWebhook(suite, staticWebhookTestPath, webhookTestSecret, payload).Code), "event %q", event)
	}
}

func TestIntegrationStaticWebhookHandlerDispatchesRegisteredEvent(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerStaticWebhookTestRoutes()

	rec := postStaticWebhook(suite, staticWebhookTestPath, webhookTestSecret, []byte(`{"event":"alert.created","delivery_id":"static-004"}`))

	assert.Assert(t, is.Equal(http.StatusOK, rec.Code))
}

func TestConfigureIntegrationProviderSkipsRuntimeOwnedWebhooks(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	rec := sendIntegrationConfigRequest(t, suite, testUser.UserCtx, staticWebhookTestDefinitionID, handlers.ConfigureIntegrationRequest{
		DefinitionID:  staticWebhookTestDefinitionID,
		CredentialRef: webhookTestCredentialRef,
		Body:          json.RawMessage(`{"token":"test-token"}`),
	})

	assert.Assert(t, is.Equal(http.StatusOK, rec.Code), rec.Body.String())

	var resp handlers.ConfigureIntegrationResponse
	assert.NilError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	assert.Check(t, len(resp.IntegrationID) > 0)
	assert.Check(t, is.Len(resp.WebhookSecret, 0))
	assert.Check(t, is.Len(resp.WebhookEndpointURL, 0))

	rows := suite.db.IntegrationWebhook.Query().
		Where(integrationwebhook.IntegrationIDEQ(resp.IntegrationID)).
		CountX(testUser.UserCtx)

	assert.Check(t, is.Equal(0, rows))
}

func TestConfigureIntegrationProviderRejectsOperatorOnlyDefinition(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	rec := sendIntegrationConfigRequest(t, suite, testUser.UserCtx, operatorOnlyTestDefinitionID, handlers.ConfigureIntegrationRequest{
		DefinitionID: operatorOnlyTestDefinitionID,
	})

	assert.Assert(t, is.Equal(http.StatusBadRequest, rec.Code), rec.Body.String())
	assert.Check(t, is.Contains(rec.Body.String(), handlers.ErrInvalidProvider.Error()))
}

func TestRunIntegrationOperationRejectsHiddenOperation(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", suite.h.ConfigureIntegrationProvider)
	suite.registerRouteOnce(http.MethodPost, operationTestPath, suite.h.RunIntegrationOperation)

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	configRec := sendIntegrationConfigRequest(t, suite, testUser.UserCtx, staticWebhookTestDefinitionID, handlers.ConfigureIntegrationRequest{
		DefinitionID:  staticWebhookTestDefinitionID,
		CredentialRef: webhookTestCredentialRef,
		Body:          json.RawMessage(`{"token":"test-token"}`),
	})
	assert.Assert(t, is.Equal(http.StatusOK, configRec.Code), configRec.Body.String())

	var configResp handlers.ConfigureIntegrationResponse
	assert.NilError(t, json.Unmarshal(configRec.Body.Bytes(), &configResp))

	body, err := json.Marshal(handlers.RunIntegrationOperationRequest{
		IntegrationID: configResp.IntegrationID,
		Body:          handlers.RunIntegrationOperationBody{Operation: staticWebhookHiddenOperation.Name()},
	})
	assert.NilError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+staticWebhookTestDefinitionID+"/operations?integration_id="+configResp.IntegrationID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	assert.Assert(t, is.Equal(http.StatusBadRequest, rec.Code), rec.Body.String())
	assert.Check(t, is.Contains(rec.Body.String(), "dispatch input invalid"))
}
