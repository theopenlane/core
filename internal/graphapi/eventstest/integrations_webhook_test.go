//go:build test

package eventstest_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	webhookdef "github.com/theopenlane/core/v2/internal/integrations/definitions/webhook"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/keystore"
)

const (
	// webhookIDHeader is the Standard Webhooks header carrying the message id
	webhookIDHeader = "webhook-id"
	// webhookTimestampHeader is the Standard Webhooks header carrying the attempt time
	webhookTimestampHeader = "webhook-timestamp"
	// webhookSignatureHeader is the Standard Webhooks header carrying the signature
	webhookSignatureHeader = "webhook-signature"
	// webhookSecretPrefix marks a Standard Webhooks secret
	webhookSecretPrefix = "whsec_"
	// webhookStaticHeader is a static header the organization configures on the destination
	webhookStaticHeader = "X-Customer-Token"
	// webhookStaticHeaderValue is the configured static header value
	webhookStaticHeaderValue = "customer-token-value"
)

// webhookDelivery is one request the local endpoint received
type webhookDelivery struct {
	// method is the HTTP method used
	method string
	// header is the received request header
	header http.Header
	// body is the received request body
	body string
	// signatureValid reports whether the signature header is the Standard Webhooks signature under the endpoint's secret
	signatureValid bool
}

// webhookEndpoint is a local endpoint that checks deliveries the way a customer's endpoint would
type webhookEndpoint struct {
	server *httptest.Server
	mu     sync.Mutex
	// secret is the signing secret the endpoint verifies against
	secret string
	// statuses are answered in order, one per delivery; 200 once empty
	statuses []int
	// deliveries are the requests received so far
	deliveries []webhookDelivery
}

// newWebhookEndpoint starts a local endpoint verifying deliveries against secret
func newWebhookEndpoint(t *testing.T, secret string) *webhookEndpoint {
	t.Helper()

	endpoint := &webhookEndpoint{secret: secret}

	endpoint.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		endpoint.mu.Lock()
		defer endpoint.mu.Unlock()

		// verified the way a Standard Webhooks consumer does, with the standard library and independently of the code that signs
		key, _ := base64.StdEncoding.DecodeString(strings.TrimPrefix(endpoint.secret, webhookSecretPrefix))

		mac := hmac.New(sha256.New, key)
		mac.Write([]byte(r.Header.Get(webhookIDHeader) + "." + r.Header.Get(webhookTimestampHeader) + "." + string(body)))
		expected := "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))

		endpoint.deliveries = append(endpoint.deliveries, webhookDelivery{
			method:         r.Method,
			header:         r.Header.Clone(),
			body:           string(body),
			signatureValid: hmac.Equal([]byte(expected), []byte(r.Header.Get(webhookSignatureHeader))),
		})

		status := http.StatusOK
		if len(endpoint.statuses) > 0 {
			status, endpoint.statuses = endpoint.statuses[0], endpoint.statuses[1:]
		}

		w.WriteHeader(status)
	}))
	t.Cleanup(endpoint.server.Close)

	return endpoint
}

// answerNext makes the endpoint answer its next deliveries with the given statuses
func (e *webhookEndpoint) answerNext(statuses ...int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.statuses = statuses
}

// setSecret changes the secret the endpoint verifies against
func (e *webhookEndpoint) setSecret(secret string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.secret = secret
}

// received returns the deliveries whose body matches
func (e *webhookEndpoint) received(body string) []webhookDelivery {
	e.mu.Lock()
	defer e.mu.Unlock()

	var matched []webhookDelivery

	for _, delivery := range e.deliveries {
		if delivery.body == body {
			matched = append(matched, delivery)
		}
	}

	return matched
}

// installWebhook installs the webhook integration for the caller's organization the way the
// configuration endpoint does: create the installation, then store and verify the credential
func installWebhook(ctx context.Context, t *testing.T, ownerID string, credential webhookdef.CredentialSchema) (*ent.Integration, error) {
	t.Helper()

	def, ok := suite.IntegrationsRT.Registry().Definition(webhookdef.DefinitionID.ID())
	assert.Assert(t, ok)

	installation, _, err := suite.IntegrationsRT.EnsureInstallation(ctx, ownerID, "", def, nil, nil)
	assert.NilError(t, err)

	return installation, saveWebhookCredential(ctx, t, installation, credential)
}

// saveWebhookCredential stores the destination credential on an installation
func saveWebhookCredential(ctx context.Context, t *testing.T, installation *ent.Integration, credential webhookdef.CredentialSchema) error {
	t.Helper()

	def, ok := suite.IntegrationsRT.Registry().Definition(webhookdef.DefinitionID.ID())
	assert.Assert(t, ok)

	data, err := json.Marshal(credential)
	assert.NilError(t, err)

	return suite.IntegrationsRT.ReconcileCredential(ctx, installation, def.Connections[0].Connection().Credential.Name, integrationtypes.CredentialSet{Data: data})
}

// dispatchWebhook queues one delivery and returns its run id
func dispatchWebhook(ctx context.Context, t *testing.T, installation *ent.Integration, body, messageID string) string {
	t.Helper()

	def, ok := suite.IntegrationsRT.Registry().Definition(webhookdef.DefinitionID.ID())
	assert.Assert(t, ok)

	config, err := json.Marshal(webhookdef.DeliverOperation{ID: messageID, Body: json.RawMessage(body)})
	assert.NilError(t, err)

	result, err := suite.IntegrationsRT.Dispatch(ctx, integrationtypes.DispatchRequest{
		IntegrationID: installation.ID,
		Operation:     def.Operations[0].Name,
		Config:        config,
	})
	assert.NilError(t, err)
	assert.Assert(t, result.RunID != "")

	return result.RunID
}

// webhookRunStatus returns the recorded status of a delivery run
func webhookRunStatus(ctx context.Context, t *testing.T, runID string) enums.IntegrationRunStatus {
	t.Helper()

	run, err := suite.Client.DB.IntegrationRun.Get(ctx, runID)
	assert.NilError(t, err)

	return run.Status
}

// TestWebhookIntegrationDelivery installs the webhook integration against a local endpoint and
// follows deliveries through the job queue to that endpoint; subtests share the installation
// and run in order
func TestWebhookIntegrationDelivery(t *testing.T) {
	const signingSecret = "whsec_Zmlyc3Qtc2lnbmluZy1rZXktMDAwMDAx"

	org := suite.UserBuilder(context.Background(), t)
	ctx := th.SetInternalContext(org.UserCtx, suite.Client.DB)

	endpoint := newWebhookEndpoint(t, signingSecret)

	destination, err := url.Parse(endpoint.server.URL)
	assert.NilError(t, err)

	credential := webhookdef.CredentialSchema{
		URL:           endpoint.server.URL,
		SigningSecret: signingSecret,
		Headers:       map[string]string{webhookStaticHeader: webhookStaticHeaderValue},
	}

	installation, err := installWebhook(ctx, t, org.OrganizationID, credential)
	assert.NilError(t, err)

	t.Run("installing connects without calling the endpoint", func(t *testing.T) {
		reloaded := reloadIntegration(t, ctx, installation.ID)

		assert.Check(t, is.Equal(org.OrganizationID, reloaded.OwnerID))
		assert.Check(t, is.Equal(enums.IntegrationStatusConnected, reloaded.Status))

		var metadata webhookdef.InstallationMetadata

		assert.NilError(t, json.Unmarshal(reloaded.InstallationMetadata.Attributes, &metadata))
		assert.Check(t, is.Equal(destination.Host, metadata.Host))

		// any request to the endpoint is a delivery, so verification must not send one
		assert.Check(t, is.Len(endpoint.received(""), 0))
	})

	t.Run("the secret is kept in the credential store, not on the installation", func(t *testing.T) {
		reloaded := reloadIntegration(t, ctx, installation.ID)

		row, err := json.Marshal(reloaded)
		assert.NilError(t, err)
		assert.Check(t, !strings.Contains(string(row), signingSecret))
		assert.Check(t, !strings.Contains(string(row), webhookStaticHeaderValue))

		def, ok := suite.IntegrationsRT.Registry().Definition(webhookdef.DefinitionID.ID())
		assert.Assert(t, ok)

		store, err := keystore.NewStore(suite.Client.DB)
		assert.NilError(t, err)

		stored, found, err := store.LoadCredential(ctx, reloaded, def.Connections[0].Connection().Credential.Name)
		assert.NilError(t, err)
		assert.Assert(t, found)

		var saved webhookdef.CredentialSchema

		assert.NilError(t, json.Unmarshal(stored.Data, &saved))
		assert.Check(t, is.DeepEqual(credential, saved))
	})

	t.Run("a delivery reaches the endpoint with the standard headers, signed, and with the configured headers", func(t *testing.T) {
		const body = `{"event":"workflow.step","object":{"type":"Control","id":"ctl_1"}}`

		runID := dispatchWebhook(ctx, t, installation, body, "delivery-1")

		waitForEvents()

		deliveries := endpoint.received(body)
		assert.Assert(t, is.Len(deliveries, 1))

		delivery := deliveries[0]
		assert.Check(t, is.Equal(http.MethodPost, delivery.method))
		assert.Check(t, delivery.signatureValid, "signature header %q did not match the body", delivery.header.Get(webhookSignatureHeader))
		assert.Check(t, is.Equal(webhookStaticHeaderValue, delivery.header.Get(webhookStaticHeader)))
		assert.Check(t, is.Equal("delivery-1", delivery.header.Get(webhookIDHeader)))
		assert.Check(t, is.Contains(delivery.header.Get("Content-Type"), "application/json"))

		assert.Check(t, is.Equal(enums.IntegrationRunStatusSuccess, webhookRunStatus(ctx, t, runID)))
	})

	t.Run("a delivery the endpoint rejects is not retried and leaves the installation connected", func(t *testing.T) {
		const body = `{"event":"workflow.step","object":{"type":"Control","id":"ctl_rejected"}}`

		endpoint.answerNext(http.StatusBadRequest)

		runID := dispatchWebhook(ctx, t, installation, body, "delivery-rejected")

		waitForEvents()

		assert.Check(t, is.Len(endpoint.received(body), 1))
		assert.Check(t, is.Equal(enums.IntegrationRunStatusFailed, webhookRunStatus(ctx, t, runID)))
		assert.Check(t, is.Equal(enums.IntegrationStatusConnected, reloadIntegration(t, ctx, installation.ID).Status))
	})

	t.Run("a delivery that fails on the endpoint's side is retried with the same key", func(t *testing.T) {
		const body = `{"event":"workflow.step","object":{"type":"Control","id":"ctl_retried"}}`

		endpoint.answerNext(http.StatusServiceUnavailable)

		dispatchWebhook(ctx, t, installation, body, "delivery-retried")

		waitForCondition(t, func() bool { return len(endpoint.received(body)) == 2 }, "delivery was not retried")

		for _, delivery := range endpoint.received(body) {
			assert.Check(t, delivery.signatureValid)
			assert.Check(t, is.Equal("delivery-retried", delivery.header.Get(webhookIDHeader)))
		}

		waitForEvents()

		assert.Check(t, is.Len(endpoint.received(body), 2))
		assert.Check(t, is.Equal(enums.IntegrationStatusConnected, reloadIntegration(t, ctx, installation.ID).Status))
	})

	t.Run("after the secret is rotated deliveries are signed with the new one", func(t *testing.T) {
		const (
			rotatedSecret = "whsec_cm90YXRlZC1zaWduaW5nLWtleS0wMDAy"
			body          = `{"event":"workflow.step","object":{"type":"Control","id":"ctl_rotated"}}`
		)

		rotated := credential
		rotated.SigningSecret = rotatedSecret

		assert.NilError(t, saveWebhookCredential(ctx, t, reloadIntegration(t, ctx, installation.ID), rotated))

		endpoint.setSecret(rotatedSecret)

		dispatchWebhook(ctx, t, installation, body, "delivery-rotated")

		waitForEvents()

		deliveries := endpoint.received(body)
		assert.Assert(t, is.Len(deliveries, 1))
		assert.Check(t, deliveries[0].signatureValid)
	})

	t.Run("another organization cannot deliver through this installation", func(t *testing.T) {
		const body = `{"event":"workflow.step","object":{"type":"Control","id":"ctl_other_org"}}`

		other := suite.UserBuilder(context.Background(), t)

		def, ok := suite.IntegrationsRT.Registry().Definition(webhookdef.DefinitionID.ID())
		assert.Assert(t, ok)

		config, err := json.Marshal(webhookdef.DeliverOperation{ID: "delivery-other-org", Body: json.RawMessage(body)})
		assert.NilError(t, err)

		_, err = suite.IntegrationsRT.Dispatch(other.UserCtx, integrationtypes.DispatchRequest{
			IntegrationID: installation.ID,
			OwnerID:       other.OrganizationID,
			Operation:     def.Operations[0].Name,
			Config:        config,
		})
		assert.Check(t, ent.IsNotFound(err), "got %v", err)

		waitForEvents()

		assert.Check(t, is.Len(endpoint.received(body), 0))
	})
}

// TestWebhookIntegrationRejectsUnusableCredential verifies a credential the connect-time check
// refuses does not leave a connected installation behind
func TestWebhookIntegrationRejectsUnusableCredential(t *testing.T) {
	org := suite.UserBuilder(context.Background(), t)
	ctx := th.SetInternalContext(org.UserCtx, suite.Client.DB)

	installation, err := installWebhook(ctx, t, org.OrganizationID, webhookdef.CredentialSchema{
		URL:     "https://hooks.example.com/openlane",
		Headers: map[string]string{"X Bad Header": "value"},
	})
	assert.Check(t, is.ErrorIs(err, webhookdef.ErrHeaderNameInvalid))
	assert.Check(t, reloadIntegration(t, ctx, installation.ID).Status != enums.IntegrationStatusConnected)
}

// TestWebhookIntegrationDestinationDictatesAuth installs against destinations that dictate how they
// are authenticated: an API key header with no signature, then the destination's own signature header
func TestWebhookIntegrationDestinationDictatesAuth(t *testing.T) {
	const (
		apiKeyHeader    = "X-Api-Key"
		apiKey          = "vendor-api-key"
		vendorHeader    = "X-Vendor-Signature"
		vendorSecret    = "vendor-shared-secret"
		unsignedBody    = `{"type":"workflow.step","data":{"id":"ctl_api_key"}}`
		vendorSignedMsg = `{"type":"workflow.step","data":{"id":"ctl_vendor_signed"}}`
	)

	org := suite.UserBuilder(context.Background(), t)
	ctx := th.SetInternalContext(org.UserCtx, suite.Client.DB)

	endpoint := newWebhookEndpoint(t, "")

	credential := webhookdef.CredentialSchema{
		URL:     endpoint.server.URL,
		Headers: map[string]string{apiKeyHeader: apiKey},
	}

	installation, err := installWebhook(ctx, t, org.OrganizationID, credential)
	assert.NilError(t, err)

	t.Run("an API key header is sent and nothing is signed", func(t *testing.T) {
		dispatchWebhook(ctx, t, installation, unsignedBody, "delivery-api-key")

		waitForEvents()

		deliveries := endpoint.received(unsignedBody)
		assert.Assert(t, is.Len(deliveries, 1))
		assert.Check(t, is.Equal(apiKey, deliveries[0].header.Get(apiKeyHeader)))
		assert.Check(t, is.Equal("delivery-api-key", deliveries[0].header.Get(webhookIDHeader)))
		assert.Check(t, is.Len(deliveries[0].header.Values(webhookSignatureHeader), 0))
	})

	t.Run("the destination's own signature header replaces the standard one", func(t *testing.T) {
		credential.SigningSecret = vendorSecret
		credential.SignatureHeader = vendorHeader
		credential.SignatureFormat = "sha256={signature}"

		assert.NilError(t, saveWebhookCredential(ctx, t, reloadIntegration(t, ctx, installation.ID), credential))

		dispatchWebhook(ctx, t, installation, vendorSignedMsg, "delivery-vendor-signed")

		waitForEvents()

		deliveries := endpoint.received(vendorSignedMsg)
		assert.Assert(t, is.Len(deliveries, 1))

		// the body alone, hex encoded, as the vendor would compute it
		mac := hmac.New(sha256.New, []byte(vendorSecret))
		mac.Write([]byte(vendorSignedMsg))

		assert.Check(t, is.Equal("sha256="+hex.EncodeToString(mac.Sum(nil)), deliveries[0].header.Get(vendorHeader)))
		assert.Check(t, is.Len(deliveries[0].header.Values(webhookSignatureHeader), 0))
		assert.Check(t, is.Equal(apiKey, deliveries[0].header.Get(apiKeyHeader)))
	})
}
