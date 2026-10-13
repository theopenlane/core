//go:build test

package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	emaildef "github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/runtime"
)

const (
	emailResendWebhookPath   = "/v1/integrations/email/webhook"
	emailPostmarkWebhookPath = "/v1/integrations/email/webhook/postmark"
	emailResendWebhookName   = "resend.delivery"
	emailPostmarkWebhookName = "postmark.delivery"
)

// withEmailProviderRuntime swaps in an integrations runtime whose email definition is built for provider, so only that provider's inbound webhook is registered, and returns the restore func
func (suite *HandlerTestSuite) withEmailProviderRuntime(t *testing.T, provider string) func() {
	t.Helper()

	rt, err := runtime.New(runtime.Config{
		DB:                 suite.galaDB,
		Gala:               suite.galaRuntime,
		Keystore:           suite.keystore,
		DefinitionBuilders: []registry.Builder{emaildef.Builder(emaildef.MockProviderRuntimeConfig(provider), false)},
	})
	require.NoError(t, err)

	original := suite.h.IntegrationsRuntime
	suite.h.IntegrationsRuntime = rt

	return func() {
		suite.h.IntegrationsRuntime = original
	}
}

func (suite *HandlerTestSuite) registerEmailWebhookRoutes() {
	suite.registerRouteOnce(http.MethodPost, emailResendWebhookPath, suite.h.IntegrationStaticWebhookHandler(emaildef.DefinitionID.ID(), emailResendWebhookName))
	suite.registerRouteOnce(http.MethodPost, emailPostmarkWebhookPath, suite.h.IntegrationStaticWebhookHandler(emaildef.DefinitionID.ID(), emailPostmarkWebhookName))
}

func (suite *HandlerTestSuite) serveEmailWebhook(req *http.Request) *httptest.ResponseRecorder {
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	return rec
}

func resendWebhookRequest(secret string, payload []byte) *http.Request {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	req := httptest.NewRequest(http.MethodPost, emailResendWebhookPath, strings.NewReader(string(payload)))
	req.Header.Set("svix-id", "msg_handler_test")
	req.Header.Set("svix-timestamp", timestamp)
	req.Header.Set("svix-signature", emaildef.SvixSignature(secret, "msg_handler_test", timestamp, payload))

	return req
}

func postmarkWebhookRequest(password string, payload []byte) *http.Request {
	req := httptest.NewRequest(http.MethodPost, emailPostmarkWebhookPath, strings.NewReader(string(payload)))
	if password != "" {
		req.SetBasicAuth("postmark", password)
	}

	return req
}

func (suite *HandlerTestSuite) TestEmailResendWebhookRejectsUnsigned() {
	t := suite.T()

	restore := suite.withEmailProviderRuntime(t, "resend")
	defer restore()

	suite.registerEmailWebhookRoutes()

	payload := []byte(`{"type":"email.delivered","data":{"email_id":"e_1","created_at":"2025-01-01T00:00:00Z","tags":{}}}`)

	rec := suite.serveEmailWebhook(httptest.NewRequest(http.MethodPost, emailResendWebhookPath, strings.NewReader(string(payload))))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "verification failed")
	assert.NotContains(t, rec.Body.String(), "svix")
}

func (suite *HandlerTestSuite) TestEmailResendWebhookRejectsWrongSecret() {
	t := suite.T()

	restore := suite.withEmailProviderRuntime(t, "resend")
	defer restore()

	suite.registerEmailWebhookRoutes()

	payload := []byte(`{"type":"email.delivered","data":{"email_id":"e_2","created_at":"2025-01-01T00:00:00Z","tags":{}}}`)

	rec := suite.serveEmailWebhook(resendWebhookRequest("whsec_d3Jvbmctc2VjcmV0", payload))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "verification failed")
}

func (suite *HandlerTestSuite) TestEmailResendWebhookAcceptsSignedEvents() {
	t := suite.T()

	restore := suite.withEmailProviderRuntime(t, "resend")
	defer restore()

	suite.registerEmailWebhookRoutes()

	for _, eventType := range []string{"email.delivered", "email.delivery_delayed"} {
		payload := []byte(`{"type":"` + eventType + `","data":{"email_id":"e_3","created_at":"2025-01-01T00:00:00Z","tags":{}}}`)

		rec := suite.serveEmailWebhook(resendWebhookRequest(emaildef.MockResendWebhookSecret, payload))

		require.Equal(t, http.StatusOK, rec.Code, "%s: %s", eventType, rec.Body.String())
	}
}

func (suite *HandlerTestSuite) TestEmailPostmarkWebhookRejectsMissingAuth() {
	t := suite.T()

	restore := suite.withEmailProviderRuntime(t, "postmark")
	defer restore()

	suite.registerEmailWebhookRoutes()

	rec := suite.serveEmailWebhook(postmarkWebhookRequest("", []byte(`{"RecordType":"Delivery","MessageID":"m_1"}`)))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "verification failed")
	assert.NotContains(t, rec.Body.String(), "basic auth")
}

func (suite *HandlerTestSuite) TestEmailPostmarkWebhookRejectsWrongPassword() {
	t := suite.T()

	restore := suite.withEmailProviderRuntime(t, "postmark")
	defer restore()

	suite.registerEmailWebhookRoutes()

	rec := suite.serveEmailWebhook(postmarkWebhookRequest("not-the-secret", []byte(`{"RecordType":"Delivery","MessageID":"m_2"}`)))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "verification failed")
}

func (suite *HandlerTestSuite) TestEmailPostmarkWebhookAcceptsAuthenticatedEvents() {
	t := suite.T()

	restore := suite.withEmailProviderRuntime(t, "postmark")
	defer restore()

	suite.registerEmailWebhookRoutes()

	for _, recordType := range []string{"Delivery", "SubscriptionChange"} {
		payload := []byte(`{"RecordType":"` + recordType + `","MessageID":"m_3","DeliveredAt":"2025-01-01T00:00:00Z","Metadata":{}}`)

		req := postmarkWebhookRequest(emaildef.MockPostmarkWebhookSecret, payload)
		req.Header.Set("X-PM-Webhook-Trace-Id", "trace-"+recordType)

		rec := suite.serveEmailWebhook(req)

		require.Equal(t, http.StatusOK, rec.Code, "%s: %s", recordType, rec.Body.String())
	}
}

func (suite *HandlerTestSuite) TestEmailWebhookOnlySelectedProviderRouteIsConfigured() {
	t := suite.T()

	restore := suite.withEmailProviderRuntime(t, "resend")
	defer restore()

	suite.registerEmailWebhookRoutes()

	req := postmarkWebhookRequest(emaildef.MockPostmarkWebhookSecret, []byte(`{"RecordType":"Delivery","MessageID":"m_4"}`))

	rec := suite.serveEmailWebhook(req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "not configured")
}
