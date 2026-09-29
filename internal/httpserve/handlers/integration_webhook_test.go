//go:build test

package handlers_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

func webhookHMACSHA256(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func (suite *HandlerTestSuite) TestIntegrationWebhookHandlerSuccess() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, webhookTestPath, suite.h.IntegrationWebhookHandler)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{webhookTestDefinitionBuilder(webhookTestDefinitionID)})
	defer restore()

	user := suite.userBuilderWithInput(context.Background(), &userInput{confirmedUser: true})

	wh := suite.createWebhookTestIntegration(t, user.UserCtx, user.OrganizationID, webhookTestDefinitionID)

	payload := []byte(`{"event":"alert.created","delivery_id":"del-001"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/webhooks/"+wh.endpointID, strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature-256", webhookHMACSHA256(wh.secretToken, payload))
	req = req.WithContext(privacy.DecisionContext(req.Context(), privacy.Allow))

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func (suite *HandlerTestSuite) TestIntegrationWebhookHandlerMissingEndpointID() {
	t := suite.T()

	// Register with a path that will result in empty endpointID
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/webhooks/", suite.h.IntegrationWebhookHandler)

	payload := []byte(`{"event":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/webhooks/", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	// Endpoint not matched or returns bad request
	assert.True(t, rec.Code == http.StatusBadRequest || rec.Code == http.StatusNotFound)
}

func (suite *HandlerTestSuite) TestIntegrationWebhookHandlerEmptyPayload() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, webhookTestPath, suite.h.IntegrationWebhookHandler)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/webhooks/some-endpoint", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestIntegrationWebhookHandlerInvalidSignature() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, webhookTestPath, suite.h.IntegrationWebhookHandler)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{webhookTestDefinitionBuilder(webhookTestDefinitionID)})
	defer restore()

	user := suite.userBuilderWithInput(context.Background(), &userInput{confirmedUser: true})

	wh := suite.createWebhookTestIntegration(t, user.UserCtx, user.OrganizationID, webhookTestDefinitionID)

	payload := []byte(`{"event":"alert.created","delivery_id":"del-002"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/webhooks/"+wh.endpointID, strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature-256", webhookHMACSHA256("wrong-secret", payload))
	req = req.WithContext(privacy.DecisionContext(req.Context(), privacy.Allow))

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestIntegrationWebhookHandlerMissingSignature() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, webhookTestPath, suite.h.IntegrationWebhookHandler)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{webhookTestDefinitionBuilder(webhookTestDefinitionID)})
	defer restore()

	user := suite.userBuilderWithInput(context.Background(), &userInput{confirmedUser: true})

	wh := suite.createWebhookTestIntegration(t, user.UserCtx, user.OrganizationID, webhookTestDefinitionID)

	payload := []byte(`{"event":"alert.created","delivery_id":"del-003"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/webhooks/"+wh.endpointID, strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(privacy.DecisionContext(req.Context(), privacy.Allow))

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestIntegrationWebhookHandlerEndpointNotFound() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, webhookTestPath, suite.h.IntegrationWebhookHandler)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{webhookTestDefinitionBuilder(webhookTestDefinitionID)})
	defer restore()

	payload := []byte(`{"event":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/webhooks/nonexistent-endpoint", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature-256", webhookHMACSHA256(webhookTestSecret, payload))
	req = req.WithContext(privacy.DecisionContext(req.Context(), privacy.Allow))

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestIntegrationWebhookHandlerEmptyEventNameReturnsSuccess() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, webhookTestPath, suite.h.IntegrationWebhookHandler)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{webhookTestDefinitionBuilder(webhookTestDefinitionID)})
	defer restore()

	user := suite.userBuilderWithInput(context.Background(), &userInput{confirmedUser: true})

	wh := suite.createWebhookTestIntegration(t, user.UserCtx, user.OrganizationID, webhookTestDefinitionID)

	// Payload with empty event name - the event handler returns empty name which should be a no-op success
	payload := []byte(`{"event":"","delivery_id":"del-004"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/webhooks/"+wh.endpointID, strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature-256", webhookHMACSHA256(wh.secretToken, payload))
	req = req.WithContext(privacy.DecisionContext(req.Context(), privacy.Allow))

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

// webhookTestIntegration holds the endpoint ID and auto-generated secret for one test webhook
type webhookTestIntegration struct {
	endpointID  string
	secretToken string
}

// createWebhookTestIntegration creates an integration and webhook record for testing,
// returning the webhook endpoint ID and auto-generated secret token
func (suite *HandlerTestSuite) createWebhookTestIntegration(t *testing.T, ctx context.Context, orgID, definitionID string) webhookTestIntegration {
	t.Helper()

	integrationRec, err := suite.db.Integration.Create().
		SetOwnerID(orgID).
		SetName(definitionID).
		SetDefinitionID(definitionID).
		Save(ctx)
	require.NoError(t, err)

	credential := types.CredentialSet{
		Data: json.RawMessage(`{"token":"test-token"}`),
	}
	err = suite.h.IntegrationsRuntime.Reconcile(ctx, integrationRec, nil, webhookTestCredentialRef, &credential, nil)
	require.NoError(t, err)

	webhookRec, err := suite.h.IntegrationsRuntime.EnsureWebhook(ctx, integrationRec, "inbound.events", "")
	require.NoError(t, err)

	require.NotNil(t, webhookRec.EndpointID)

	return webhookTestIntegration{
		endpointID:  *webhookRec.EndpointID,
		secretToken: webhookRec.SecretToken,
	}
}
