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

const (
	webhookTestDefinitionID = "def_01K0TESTWBHK0000000000001"
	webhookTestPath         = "/v1/integrations/webhooks/:endpointID"
	webhookTestSecret       = "webhook-test-secret"
)

// WebhookTestHealthCheck is the config type for the webhook test health check operation
type WebhookTestHealthCheck struct{}

// webhookTestAlertEnvelope is the payload type for the test webhook events
type webhookTestAlertEnvelope struct{}

// webhookTestCredential is the credential type stored by the webhook test definition
type webhookTestCredential struct {
	// Token is the stored token
	Token string `json:"token"`
}

var (
	webhookTestCredentialRef    = types.CredentialRefOf[webhookTestCredential]()
	webhookHealthCheckOperation = types.OperationRefOf[WebhookTestHealthCheck]()
	webhookAlertCreatedEvent    = types.NewWebhookEventRef[webhookTestAlertEnvelope]("alert.created")
)

func webhookTestDefinitionBuilder(definitionID string) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		definition := types.NewDefinitionRef(definitionID)

		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID,
				DisplayName: "Webhook Test",
				Active:      true,
				Visible:     true,
			},
			CredentialRegistrations: []types.CredentialRegistration{
				webhookTestCredentialRef.Registration(types.CredentialRegistration{
					Name: "Webhook Test Credential",
				}),
			},
			HealthCheck: types.CredentialHealthCheck(func(context.Context, types.OperationRequest) (json.RawMessage, error) {
				return json.RawMessage(`{"ok":true}`), nil
			}),
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:  webhookTestCredentialRef.ID(),
					Name:           "Webhook Test Connection",
					CredentialRefs: []types.CredentialSlotID{webhookTestCredentialRef.ID()},
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
						webhookAlertCreatedEvent.Registration(definition, types.WebhookEventRegistration{
							Handle: func(context.Context, types.WebhookHandleRequest) error {
								return nil
							},
						}),
					},
				},
			},
			Operations: []types.OperationRegistration{
				webhookHealthCheckOperation.Registration(definition, types.OperationRegistration{
					Description: "Health check",
					Policy:      types.ExecutionPolicy{Inline: true},
					Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
						return json.RawMessage(`{"ok":true}`), nil
					},
				}),
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

	suite.registerRouteOnce(http.MethodPost, webhookTestPath, suite.h.IntegrationWebhookHandler)

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

	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/webhooks/", suite.h.IntegrationWebhookHandler)

	payload := []byte(`{"event":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/webhooks/", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

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

	suite.registerRouteOnce(http.MethodPost, webhookTestPath, suite.h.IntegrationWebhookHandler)

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

	requestCtx := privacy.DecisionContext(httptest.NewRequest(http.MethodGet, "/", nil).Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	wh := suite.createWebhookTestIntegration(t, user.UserCtx, user.OrganizationID, webhookTestDefinitionID)

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

// createWebhookTestIntegration creates an integration and webhook record for testing, returning the endpoint ID and secret token
func (suite *HandlerTestSuite) createWebhookTestIntegration(t *testing.T, ctx context.Context, orgID, definitionID string) webhookTestIntegration {
	t.Helper()

	def, ok := suite.h.IntegrationsRuntime.Definition(definitionID)
	require.True(t, ok)

	integrationRec, _, err := suite.h.IntegrationsRuntime.EnsureInstallation(ctx, orgID, "", def)
	require.NoError(t, err)

	credential := types.CredentialSet{
		Data: json.RawMessage(`{"token":"test-token"}`),
	}
	err = suite.h.IntegrationsRuntime.Reconcile(ctx, integrationRec, nil, webhookTestCredentialRef.ID(), &credential, nil)
	require.NoError(t, err)

	webhookRec, err := suite.h.IntegrationsRuntime.EnsureWebhook(ctx, integrationRec, "inbound.events", "")
	require.NoError(t, err)

	require.NotNil(t, webhookRec.EndpointID)

	return webhookTestIntegration{
		endpointID:  *webhookRec.EndpointID,
		secretToken: webhookRec.SecretToken,
	}
}
