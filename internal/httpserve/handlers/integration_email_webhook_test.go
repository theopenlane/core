//go:build test

package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	emaildef "github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/pkg/gala"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

const (
	emailResendWebhookPath   = "/v1/integrations/email/webhook"
	emailPostmarkWebhookPath = "/v1/integrations/email/webhook/postmark"
	emailResendWebhookName   = "resend.delivery"
	emailPostmarkWebhookName = "postmark.delivery"
)

// withEmailProviderRuntime swaps in an integrations runtime whose email definition is built for provider, so only that provider's inbound webhook is registered, and returns the restore func
//
// the runtime gets its own in-memory gala because listeners registered on the shared durable bus outlive the swap and would fan out every email operation later tests dispatch
func (suite *HandlerTestSuite) withEmailProviderRuntime(t *testing.T, provider string) func() {
	t.Helper()

	galaInstance, err := gala.NewGala(t.Context(), gala.Config{DispatchMode: gala.DispatchModeInMemory, WorkerCount: 1})
	assert.NilError(t, err)

	rt, err := runtime.New(runtime.Config{
		DB:                 suite.galaDB,
		Gala:               galaInstance,
		Keystore:           suite.keystore,
		DefinitionBuilders: []registry.Builder{emaildef.Builder(emaildef.MockProviderRuntimeConfig(provider), false)},
	})
	assert.NilError(t, err)

	original := suite.h.IntegrationsRuntime
	suite.h.IntegrationsRuntime = rt

	return func() {
		suite.h.IntegrationsRuntime = original

		assert.NilError(t, galaInstance.WaitIdle(t.Context()))
		assert.NilError(t, galaInstance.Close())
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

func TestEmailResendWebhookRejectsUnsigned(t *testing.T) {
	suite.setupTest(t)

	restore := suite.withEmailProviderRuntime(t, "resend")
	defer restore()

	suite.registerEmailWebhookRoutes()

	payload := []byte(`{"type":"email.delivered","data":{"email_id":"e_1","created_at":"2025-01-01T00:00:00Z","tags":{}}}`)

	rec := suite.serveEmailWebhook(httptest.NewRequest(http.MethodPost, emailResendWebhookPath, strings.NewReader(string(payload))))

	assert.Assert(t, is.Equal(http.StatusBadRequest, rec.Code))
	assert.Check(t, is.Contains(rec.Body.String(), "verification failed"))
	assert.Check(t, !strings.Contains(rec.Body.String(), "svix"))
}

func TestEmailResendWebhookRejectsWrongSecret(t *testing.T) {
	suite.setupTest(t)

	restore := suite.withEmailProviderRuntime(t, "resend")
	defer restore()

	suite.registerEmailWebhookRoutes()

	payload := []byte(`{"type":"email.delivered","data":{"email_id":"e_2","created_at":"2025-01-01T00:00:00Z","tags":{}}}`)

	rec := suite.serveEmailWebhook(resendWebhookRequest("whsec_d3Jvbmctc2VjcmV0", payload))

	assert.Assert(t, is.Equal(http.StatusBadRequest, rec.Code))
	assert.Check(t, is.Contains(rec.Body.String(), "verification failed"))
}

func TestEmailResendWebhookAcceptsSignedEvents(t *testing.T) {
	suite.setupTest(t)

	restore := suite.withEmailProviderRuntime(t, "resend")
	defer restore()

	suite.registerEmailWebhookRoutes()

	for _, eventType := range []string{"email.delivered", "email.delivery_delayed"} {
		payload := []byte(`{"type":"` + eventType + `","data":{"email_id":"e_3","created_at":"2025-01-01T00:00:00Z","tags":{}}}`)

		rec := suite.serveEmailWebhook(resendWebhookRequest(emaildef.MockResendWebhookSecret, payload))

		assert.Assert(t, is.Equal(http.StatusOK, rec.Code), "%s: %s", eventType, rec.Body.String())
	}
}

func TestEmailPostmarkWebhookRejectsMissingAuth(t *testing.T) {
	suite.setupTest(t)

	restore := suite.withEmailProviderRuntime(t, "postmark")
	defer restore()

	suite.registerEmailWebhookRoutes()

	rec := suite.serveEmailWebhook(postmarkWebhookRequest("", []byte(`{"RecordType":"Delivery","MessageID":"m_1"}`)))

	assert.Assert(t, is.Equal(http.StatusBadRequest, rec.Code))
	assert.Check(t, is.Contains(rec.Body.String(), "verification failed"))
	assert.Check(t, !strings.Contains(rec.Body.String(), "basic auth"))
}

func TestEmailPostmarkWebhookRejectsWrongPassword(t *testing.T) {
	suite.setupTest(t)

	restore := suite.withEmailProviderRuntime(t, "postmark")
	defer restore()

	suite.registerEmailWebhookRoutes()

	rec := suite.serveEmailWebhook(postmarkWebhookRequest("not-the-secret", []byte(`{"RecordType":"Delivery","MessageID":"m_2"}`)))

	assert.Assert(t, is.Equal(http.StatusBadRequest, rec.Code))
	assert.Check(t, is.Contains(rec.Body.String(), "verification failed"))
}

func TestEmailPostmarkWebhookAcceptsAuthenticatedEvents(t *testing.T) {
	suite.setupTest(t)

	restore := suite.withEmailProviderRuntime(t, "postmark")
	defer restore()

	suite.registerEmailWebhookRoutes()

	for _, recordType := range []string{"Delivery", "SubscriptionChange"} {
		payload := []byte(`{"RecordType":"` + recordType + `","MessageID":"m_3","DeliveredAt":"2025-01-01T00:00:00Z","Metadata":{}}`)

		req := postmarkWebhookRequest(emaildef.MockPostmarkWebhookSecret, payload)
		req.Header.Set("X-PM-Webhook-Trace-Id", "trace-"+recordType)

		rec := suite.serveEmailWebhook(req)

		assert.Assert(t, is.Equal(http.StatusOK, rec.Code), "%s: %s", recordType, rec.Body.String())
	}
}

func TestEmailWebhookOnlySelectedProviderRouteIsConfigured(t *testing.T) {
	suite.setupTest(t)

	restore := suite.withEmailProviderRuntime(t, "resend")
	defer restore()

	suite.registerEmailWebhookRoutes()

	req := postmarkWebhookRequest(emaildef.MockPostmarkWebhookSecret, []byte(`{"RecordType":"Delivery","MessageID":"m_4"}`))

	rec := suite.serveEmailWebhook(req)

	assert.Assert(t, is.Equal(http.StatusBadRequest, rec.Code))
	assert.Check(t, is.Contains(rec.Body.String(), "not configured"))
}
