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

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	webhookTestDefinitionID = "def_01K0TESTWBHK0000000000001"
	webhookTestPath         = "/v1/integrations/webhooks/:endpointID"
	webhookTestSecret       = "webhook-test-secret"
)

// WebhookTestHealthCheck is the config type for the webhook test health check operation
type WebhookTestHealthCheck struct{}

// webhookTestAlertEnvelope is the payload type for the test webhook events
type webhookTestAlertEnvelope struct{}

var (
	webhookTestCredentialRef                         = types.NewCredentialSlotID("webhook_test")
	webhookHealthSchema, webhookHealthCheckOperation = providerkit.OperationSchema[WebhookTestHealthCheck]()
	webhookAlertCreatedEvent                         = types.NewWebhookEventRef[webhookTestAlertEnvelope]("alert.created")
)

func webhookTestDefinitionBuilder(definitionID string) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID,
				DisplayName: "Webhook Test",
				Active:      true,
				Visible:     true,
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:    webhookTestCredentialRef,
					Name:   "Webhook Test Credential",
					Schema: json.RawMessage(`{"type":"object","properties":{"token":{"type":"string"}}}`),
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:  webhookTestCredentialRef,
					Name:           "Webhook Test Connection",
					CredentialRefs: []types.CredentialSlotID{webhookTestCredentialRef},
				},
			},
			Webhooks: []types.WebhookRegistration{
				{
					Name: "inbound.events",
					Event: func(req types.WebhookInboundRequest) (types.WebhookReceivedEvent, error) {
						var envelope struct {
							Event      string `json:"event"`
							DeliveryID string `json:"delivery_id"`
						}
						if err := json.Unmarshal(req.Payload, &envelope); err != nil {
							return types.WebhookReceivedEvent{}, err
						}

						if envelope.Event == "" {
							return types.WebhookReceivedEvent{}, nil
						}

						return types.WebhookReceivedEvent{
							Name:       envelope.Event,
							DeliveryID: envelope.DeliveryID,
							Payload:    req.Payload,
						}, nil
					},
					Events: []types.WebhookEventRegistration{
						{
							Name:  webhookAlertCreatedEvent.Name(),
							Topic: types.NewDefinitionRef(definitionID).WebhookEventTopic(webhookAlertCreatedEvent.Name()),
							Handle: func(context.Context, types.WebhookHandleRequest) error {
								return nil
							},
						},
					},
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         webhookHealthCheckOperation.Name(),
					Description:  "Health check",
					Topic:        types.NewDefinitionRef(definitionID).OperationTopic(webhookHealthCheckOperation.Name()),
					Policy:       types.ExecutionPolicy{Inline: true},
					ConfigSchema: webhookHealthSchema,
					Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
						return json.RawMessage(`{"ok":true}`), nil
					},
				},
			},
		}, nil
	})
}

func webhookHMACSHA256(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func (suite *HandlerTestSuite) TestIntegrationWebhookHandlerSuccess() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "IntegrationWebhookHandlerSuccess"
	suite.registerRouteOnce(http.MethodPost, webhookTestPath, op, suite.h.IntegrationWebhookHandler)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{webhookTestDefinitionBuilder(webhookTestDefinitionID)})
	defer restore()

	requestCtx := privacy.DecisionContext(httptest.NewRequest(http.MethodGet, "/", nil).Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

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

	op := openapi3.NewOperation()
	op.OperationID = "IntegrationWebhookHandlerMissingEndpointID"
	// Register with a path that will result in empty endpointID
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/webhooks/", op, suite.h.IntegrationWebhookHandler)

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

	op := openapi3.NewOperation()
	op.OperationID = "IntegrationWebhookHandlerEmptyPayload"
	suite.registerRouteOnce(http.MethodPost, webhookTestPath, op, suite.h.IntegrationWebhookHandler)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/webhooks/some-endpoint", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestIntegrationWebhookHandlerInvalidSignature() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "IntegrationWebhookHandlerInvalidSig"
	suite.registerRouteOnce(http.MethodPost, webhookTestPath, op, suite.h.IntegrationWebhookHandler)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{webhookTestDefinitionBuilder(webhookTestDefinitionID)})
	defer restore()

	requestCtx := privacy.DecisionContext(httptest.NewRequest(http.MethodGet, "/", nil).Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

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

	op := openapi3.NewOperation()
	op.OperationID = "IntegrationWebhookHandlerMissingSig"
	suite.registerRouteOnce(http.MethodPost, webhookTestPath, op, suite.h.IntegrationWebhookHandler)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{webhookTestDefinitionBuilder(webhookTestDefinitionID)})
	defer restore()

	requestCtx := privacy.DecisionContext(httptest.NewRequest(http.MethodGet, "/", nil).Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

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

	op := openapi3.NewOperation()
	op.OperationID = "IntegrationWebhookHandlerEndpointNotFound"
	suite.registerRouteOnce(http.MethodPost, webhookTestPath, op, suite.h.IntegrationWebhookHandler)

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

	op := openapi3.NewOperation()
	op.OperationID = "IntegrationWebhookHandlerEmptyEvent"
	suite.registerRouteOnce(http.MethodPost, webhookTestPath, op, suite.h.IntegrationWebhookHandler)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{webhookTestDefinitionBuilder(webhookTestDefinitionID)})
	defer restore()

	requestCtx := privacy.DecisionContext(httptest.NewRequest(http.MethodGet, "/", nil).Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

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
